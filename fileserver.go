package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	. "github.com/schollz/progressbar/v2"
)

var (
	dataDir  string
	serverIp string
	port     string
)

func main() {
	flag.StringVar(&serverIp, "ip", "", "default auto select local ip address")
	flag.StringVar(&port, "port", "9090", "port number")
	flag.StringVar(&dataDir, "dir", "./", "file server data dir")

	flag.Parse()

	checkPath(dataDir)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexPageHandler)
	mux.HandleFunc("/upload", upload)
	mux.HandleFunc("/file/", staticServer)

	server := http.Server{
		Addr:         serverIp + ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	serve(server)
}

func serve(server http.Server) {
	var wg sync.WaitGroup
	exit := make(chan os.Signal)

	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-exit
		wg.Add(1)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			fmt.Println(err)
		}
		wg.Done()
	}()

	if serverIp == "" {
		if ip, err := findIp(); err == nil {
			serverIp = ip
		}
	}
	fmt.Printf("Now serving on http://%s:%s/\n", serverIp, port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	wg.Wait()
	fmt.Println("\ngracefully shutdown the http server...")
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		uploadPageHandler(w, r)
		return
	}

	// 32MB
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusOK)
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		http.Error(w, "上传错误: "+err.Error(), http.StatusOK)
		return
	}

	filename := r.FormValue("uploadfile")
	if filename == "" {
		filename = handler.Filename
	}

	newFile, err := os.OpenFile(filepath.Join(dataDir, filename), os.O_CREATE|os.O_WRONLY, 0775)
	defer newFile.Close()

	if err != nil {
		http.Error(w, "上传失败: "+err.Error(), http.StatusOK)
	}

	log.Println(strings.Split(r.RemoteAddr, ":")[0], r.RequestURI, "create ", filepath.Join(dataDir, filename))

	bar := NewOptions64(
		handler.Size,
		OptionSetTheme(Theme{Saucer: "#", SaucerPadding: "-", BarStart: ">", BarEnd: "<"}),
		OptionSetWidth(100),
		OptionSetBytes64(handler.Size),
	)

	out := io.MultiWriter(newFile, bar)

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "上传失败: "+err.Error(), http.StatusOK)
		return
	}

	filePath, _ := filepath.Abs(filepath.Join(dataDir, filename))
	http.Error(w, fmt.Sprintf("%v", filename+"	上传完成, 文件路径:"+filePath), http.StatusOK)
}

func indexPageHandler(w http.ResponseWriter, _ *http.Request) {
	const indexTpl = `
<html>
<head>
	<meta http-equiv="Content-Type" content="text/html;charset=UTF-8">
	<title>Go index</title>
</head>
<body>
	<div>
		<a href="/upload">上传文件</a></p>
		<a href="/file">查看文件</a></p>
	</div>

	<br>
	<p>命令行访问</p>
	<p>&nbsp;&nbsp;(1)上传</p>
	<p>&nbsp;&nbsp;&nbsp;&nbsp;curl -F uploadfile=@<em><strong>yourfilename</strong></em>&nbsp;http://{{.Ip}}:{{.Port}}/upload</p>
	<p>&nbsp;&nbsp;(2)下载</p>
	<p>&nbsp;&nbsp;&nbsp;&nbsp;curl -o&nbsp;<em><strong>yourfilename</strong></em>&nbsp; http://{{.Ip}}:{{.Port}}/file/<em><strong>yourfilename</strong></em></p>

</body>
</html>
`

	t, err := template.New("index").Parse(indexTpl)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err = t.Execute(w, struct {
		Ip   string
		Port string
	}{
		Ip:   serverIp,
		Port: port,
	}); err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}

func staticServer(w http.ResponseWriter, r *http.Request) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	//trim '/file/'
	log.Println(ip, filepath.Join(dataDir, r.RequestURI[6:]), "visit")

	http.StripPrefix("/file", http.FileServer(http.Dir(dataDir))).ServeHTTP(w, r)
}

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	const tpl = `
<html>
	<title>Go upload</title>
<head>
<script type="text/javascript">
        var isIE = /msie/i.test(navigator.userAgent) && !window.opera;
        function fileChange(target,id) {
            var fileSize = 0;

            if (isIE && !target.files) {
                var filePath = target.value;
                var fileSystem = new ActiveXObject("Scripting.FileSystemObject");
                if(!fileSystem.FileExists(filePath)){
                    alert("文件不存在，请重新输入！");
                    return false;
                }
                var file = fileSystem.GetFile (filePath);
                fileSize = file.Size;
            } else {
                fileSize = target.files[0].size;
            }

            var size = fileSize / 1024 / 1024;
            if(size>{{.FileMaxSize}}){
                alert("文件大小不能大于"+{{.FileMaxSize}}+"MB！");
                target.value ="";
                return false;
            }
        }
    </script>
</head>
<body>
	<form enctype="multipart/form-data" action="{{.URI}}/upload" method="post">
		<input type="file" name="uploadfile" id="uploadfile" onchange="fileChange(this);"> <br>
		optional rename:
		<input type="text" name="uploadfile"> <br>
		<input type="submit" name="submit" value="upload">
	</form>
</body>
</html>
`

	t, err := template.New("page").Parse(tpl)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err = t.Execute(w, struct {
		FileMaxSize int
		URI         string
	}{
		//10G = 10240M
		FileMaxSize: 10 * 1024,
		URI:         strings.TrimSuffix((r.RequestURI), "/upload"),
	}); err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
}

/*
# Utility function to guess the IP (as a string) where the server can be
# reached from the outside. Quite nasty problem actually.
*/
func findIp() (string, error) {
	/*
		# we get a UDP-socket for the TEST-networks reserved by IANA.
		# It is highly unlikely, that there is special routing used
		# for these networks, hence the socket later should give us
		# the ip address of the default route.
		# We're doing multiple tests, to guard against the computer being
		# part of a test installation.
	*/
	var (
		localIp    string
		candidates = make([]string, 3)
	)

	for _, testIp := range []string{"192.0.2.0", "198.51.100.0", "203.0.113.0"} {
		udpAddr, err := net.ResolveUDPAddr("udp", testIp+":80")
		if err != nil {
			return "", fmt.Errorf("%w", err)
		}

		udpconn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			return "", fmt.Errorf("%w", err)
		}

		localIp = strings.Split(udpconn.LocalAddr().String(), ":")[0]
		for _, ip := range candidates {
			if ip == localIp {
				return localIp, nil
			}
		}
		candidates = append(candidates[:], localIp)
	}
	return candidates[0], nil
}

func checkPath(dir string) {
	if _, err := os.Stat(dir); err != nil {
		fmt.Printf("check path: %v\n", err)
		os.Exit(1)
	}
}
