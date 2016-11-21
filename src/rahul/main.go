package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"text/template"
	"time"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll every time.
	waitPeriod = 10 * time.Second
)

var (
	fileName         string
	FileNameposition = int64(0)
	upgrader         = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type TemplateData struct {
	Host string
	Data string
}

func handleHomeRequest(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.Error(writer, "Not found", 404)
		return
	}
	if request.Method != "GET" {
		http.Error(writer, "Method not allowed", 405)
		return
	}

	homeTemplate, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(writer, "Oops template not found", http.StatusInternalServerError)
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	p, n, err := readFileAfter(-1)
	if err != nil {
		p = []byte(err.Error())
	} else {
		FileNameposition += n
	}
	v := TemplateData{request.Host, string(p)}
	homeTemplate.Execute(writer, &v)
}

func main() {
	config := flag.String("port", ":8080", "http server address")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("file not specified")
	}
	fileName = flag.Args()[0]
	http.HandleFunc("/", handleHomeRequest)
	http.HandleFunc("/ws", handleWebsocket)
	if err := http.ListenAndServe(*config, nil); err != nil {
		log.Fatal(err)
	}
}
