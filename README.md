## fileserver —— “文件上传+文件浏览+文件下载”服务

- 特别适合办公室局域网环境下, 不同操作系统平台(windows,linux,macos,移动端等)之间临时传输文件
- 非常简单，容易上手
- 基于HTTP协议, 同时支持浏览器和curl命令行
- *暂不支持批量上传/下载*
- *最大上传文件大小10GB*
- *最长请求处理时间10分钟*


#### 1. fileserver命令

查看帮助
```bash
$./fileserver -h
 Usage of ./fileserver:
   -dir string
     	file server data dir (default "./")
   -port string
     	port number (default "9090")

```

运行服务
```bash
$ ./fileserver -port 9000 -dir ./test
Now serving on http://192.168.9.217:9000/

2019/12/20 11:01:47 127.0.0.1 /upload create  ./testtest.pdf
 100% >####################################################################################################< (861.2 MB/s) [0s:0s]2019/12/20 11:01:57 127.0.0.1 ./test visit
2019/12/20 11:02:00 127.0.0.1 ./test visit
2019/12/20 11:02:06 127.0.0.1 ./test visit
2019/12/20 11:02:06 127.0.0.1 ./test visit

```


#### 2. 命令行上传/下载

上传
```bash
$ curl -F uploadfile=@test.pdf http://127.0.0.1:9090/upload
test.pdf	上传完成, 文件路径:/home/chain/test/test.pdf
```

下载
```bash
$ curl -o test2.pdf http://127.0.0.1:9090/file/test.pdf
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  153M  100  153M    0     0   965M      0 --:--:-- --:--:-- --:--:--  967M
```


#### 3. 浏览器上传/下载

文件服务器首页
![index](./doc/index.png)

选择上传文件
![upload1](./doc/upload1.png)

可对上传文件重命名
![upload2](./doc/upload2.png)

上传成功
![upload3](./doc/upload3.png)

查看文件
![file1](./doc/file1.png)

点击浏览器右键——“另存为”
![file2](./doc/file2.png)

选择保存目录
![file3](./doc/file3.png)
