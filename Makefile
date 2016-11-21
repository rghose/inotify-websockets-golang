FILENAME=main_server
all:
	@echo "Run the following before anything else. Then do - make build. The server file name is ${FILENAME}"
	export GOPATH=`pwd`
	export GOBIN=`pwd`/bin
build:
	go get github.com/fsnotify/fsnotify
	go get github.com/gorilla/websocket
	go build -o bin/${FILENAME} rahul
