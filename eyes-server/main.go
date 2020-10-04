

package main

import (
	"fmt"
    "os"
	"time"
	"net/http"
	"sync/atomic"
	"github.com/gorilla/websocket"
)

type msg struct {
        str []byte
}
var loadstorevar atomic.Value

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func sendServer(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("sender (input) started\n")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil { fmt.Println(err) ; return }
	for {
		_, message, err := conn.ReadMessage()
		if err != nil { fmt.Println(err); return }
		v := msg{ message }
		loadstorevar.Store(v)
		// fmt.Printf("mess=%s\n",message)
	}
}

func getServer(w http.ResponseWriter, r *http.Request) {
	//var m []byte("hello")
	fmt.Printf("getter (eyeballs) started\n")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil { fmt.Println(err) ; return }
	for {
		v := loadstorevar.Load().(msg)
		var err = conn.WriteMessage(1,v.str)
		time.Sleep(100*time.Millisecond);
		if err != nil { fmt.Println(err); return }
		// fmt.Printf("sent mess\n")
	}
}

func main() {

    for _, pair := range os.Environ() {
        fmt.Println(pair)
    }

	v := msg{ []byte("NULL") }
	loadstorevar.Store(v)

	http.HandleFunc("/send", sendServer)
	http.HandleFunc("/get", getServer)
	http.Handle("/", http.FileServer(http.Dir("/app/html/")))

	fmt.Printf("Starting \n");
	err := http.ListenAndServe(":"+os.Args[1], nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

