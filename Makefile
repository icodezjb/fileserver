ifndef GOOS
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Darwin)
		GOOS := darwin
	else ifeq ($(UNAME_S),Linux)
		GOOS := linux
	else
		$(error "$$GOOS is not defined. If you are using Windows, try to re-make using 'GOOS=windows make ...' ")
	endif
endif

FILESERVER_BINARY32 := fileserver-$(GOOS)_386
FILESERVER_BINARY64 := fileserver-$(GOOS)_amd64

fileserver:
	@echo "Building fileserver to ./fileserver"
	@go build -o fileserver fileserver.go

target:
	mkdir -p $@

binary: target/$(FILESERVER_BINARY32) target/$(FILESERVER_BINARY64)

ifeq ($(GOOS), windows)
release: binary
		cd target && cp -f $(FILESERVER_BINARY32) $(FILESERVER_BINARY32).exe
		cd target && md5sum $(FILESERVER_BINARY32).exe > $(FILESERVER_BINARY32).md5
		cd target && zip $(FILESERVER_BINARY32).zip $(FILESERVER_BINARY32).exe $(FILESERVER_BINARY32).md5
		cd target && rm -f $(FILESERVER_BINARY32) $(FILESERVER_BINARY32).exe $(FILESERVER_BINARY32).md5
		cd target && cp -f $(FILESERVER_BINARY64) $(FILESERVER_BINARY64).exe
		cd target && md5sum $(FILESERVER_BINARY64).exe > $(FILESERVER_BINARY64).md5
		cd target && zip $(FILESERVER_BINARY64).zip $(FILESERVER_BINARY64).exe $(FILESERVER_BINARY64).md5
		cd target && rm -f $(FILESERVER_BINARY64) $(FILESERVER_BINARY64).exe $(FILESERVER_BINARY64).md5
else
release: binary
		cd target && md5sum $(FILESERVER_BINARY32) > $(FILESERVER_BINARY32).md5
		cd target && tar -czf $(FILESERVER_BINARY32).tgz $(FILESERVER_BINARY32) $(FILESERVER_BINARY32).md5
		cd target && rm -f $(FILESERVER_BINARY32) $(FILESERVER_BINARY32).md5
		cd target && md5sum $(FILESERVER_BINARY64) > $(FILESERVER_BINARY64).md5
		cd target && tar -czf $(FILESERVER_BINARY64).tgz $(FILESERVER_BINARY64) $(FILESERVER_BINARY64).md5
		cd target && rm -f $(FILESERVER_BINARY64) $(FILESERVER_BINARY64).md5
endif

release-all: clean
		GOOS=darwin   make release
		@echo
		GOOS=linux    make release
		@echo
		GOOS=windows  make release

clean:
	@echo "Cleaning binary built..."
	rm -fr target
	@echo "Done.\n"

target/$(FILESERVER_BINARY32):
	CGO_ENABLED=0 GOARCH=386 go build -o $@ fileserver.go
target/$(FILESERVER_BINARY64):
	CGO_ENABLED=0 GOARCH=amd64 go build -o $@ fileserver.go

all: release-all