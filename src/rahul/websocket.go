package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func readFileAfter(position int64) ([]byte, int64, error) {
	fi, err := os.Stat(fileName)
	if err != nil {
		return nil, 0, err
	}
	newSize := fi.Size()
	f, err := os.Open(fileName)
	if err != nil {
		return nil, 0, err
	}
	if position == -1 {
		pos := newSize
		b := make([]byte, 1)
		for numNewLines := 0; numNewLines < 10; {
			pos = pos - 1
			_, err := f.ReadAt(b, pos)
			if err != nil {
				log.Println("oops")
				break
			}
			if string(b[0]) == "\n" || string(b[0]) == "\r" {
				numNewLines++
			}
			//pos = pos - 1
			f.Seek(pos, os.SEEK_CUR)
		}
		position = newSize
	}
	FileNameposition = position
	filePosition, err := f.Seek(position, os.SEEK_SET)
	if err != nil {
		return nil, 0, err
	}
	bufferSize := newSize - filePosition
	p := make([]byte, bufferSize)
	_, err = f.ReadAt(p, filePosition)
	if err == io.EOF {
		return p, int64(len(p)), nil
	}
	if err != nil {
		return nil, 0, err
	}
	return p, int64(len(p)), nil
}

func readFromSocket(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writeToSocket(ws *websocket.Conn) {
	lastError := ""

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	err = watcher.Add(fileName)
	if err != nil {
		log.Fatal(err)
	}

	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		watcher.Close()
		ws.Close()
	}()
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)
				p, n, err := readFileAfter(FileNameposition)
				if err != nil {
					if s := err.Error(); s != lastError {
						lastError = s
						p = []byte(lastError)
					}
				} else {
					FileNameposition += n
					lastError = ""
				}

				if p != nil {
					ws.SetWriteDeadline(time.Now().Add(writeWait))
					if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {
						return
					}
				}
			}

		case err := <-watcher.Errors:
			log.Println("error:", err)

		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}

	<-done
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go writeToSocket(ws)
	readFromSocket(ws)
}
