package main

import (
	"fmt"
	"flag"
	"net"
	"net/http"
	"io"
	"runtime"
	"log"

	"github.com/gorilla/websocket"
)

const (
	bufferSize = 1024
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin:       func(_ *http.Request) bool { return true },
		EnableCompression: true,
	}
	wsAddr = ""
	rpiAddr = ""
	ErrCopyEmptyBuffer = fmt.Errorf("copy empty buffer")
)

func init() {
	flag.StringVar(&wsAddr, "wsaddr", "localhost:9090", "http websocket server")
	flag.StringVar(&rpiAddr, "rpiaddr", "localhost:3030", "raspberry pi video stream server")
	flag.Parse()
}

func netCopy(input, output net.Conn) (err error) {
	var count int64
	count, err = io.Copy(output, input)
	if err == nil && count < 0 {
		err = ErrCopyEmptyBuffer
		return
	}
	return
}

func rpiStream(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, bufferSize)
			buf = buf[:runtime.Stack(buf, false)]
			log.Println(string(buf))
		}
	}()
	fmt.Println("Got connection")
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer wsConn.Close()
	// wrap wsConn to net.Conn
	wrapWSConn := NewWSConn(wsConn)
	defer wrapWSConn.Close()
	fmt.Println("Got websocket connection: ", wrapWSConn.RemoteAddr())
	rpiConn, err := net.Dial("tcp", rpiAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer rpiConn.Close()
	// tunnel both connection
	go netCopy(wrapWSConn, rpiConn)
	netCopy(rpiConn, wrapWSConn)
}

func main() {
	// dial rpi stream
	http.HandleFunc("/", rpiStream)
	fmt.Println("Serve: ", wsAddr)
	log.Fatal(http.ListenAndServe(wsAddr, nil))
}