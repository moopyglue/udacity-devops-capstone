
package main

import (
	"fmt"
    "bytes"
    "sync"
    "os"
	"time"
	"net/http"
	"sync/atomic"
	"github.com/gorilla/websocket"
)

var loadstorevar atomic.Value

type msg struct {
    str []byte
}

var messages = make(map[string]msg)
var links    = make(map[string]string)
var mux      = &sync.Mutex{}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  10240,
    WriteBufferSize: 10240,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func sendSession(w http.ResponseWriter, r *http.Request) {

    sid:=r.URL.Query().Get("s");
    gid:=r.URL.Query().Get("g");
    logtag:=sid+":  in: ";
    fmt.Println(logtag,"started: adding in client ",gid)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
        fmt.Println(logtag,err);
        fmt.Println(logtag,"exiting")
        return
    }

    crossLink(gid,sid)
	fmt.Println(logtag,"crossLink g=",gid," s=",sid)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
            fmt.Println(logtag,err);
            fmt.Println(logtag,"exiting")
            sendUnlink(sid)
            return
        }
		v := msg{ message }
        sendVal(sid,message)
		loadstorevar.Store(v)
		fmt.Println(logtag,string(message))
		fmt.Println(links)
	}
}

func getSession(w http.ResponseWriter, r *http.Request) {

	//var m []byte("hello")
    id:=r.URL.Query().Get("g");
    logtag:=id+": out: ";
    fmt.Println(logtag,"started")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
        fmt.Println(logtag,err);
        fmt.Println(logtag,"exiting")
        return
    }

    lastv := []byte{}
    nosend := 0
	for {
        v := getVal(id)
        if( string(v.str) != "NULL" ) {
            if( bytes.Compare(v.str,lastv) != 0 || nosend > 100 ) {
		        var err = conn.WriteMessage(1,v.str)
		        if err != nil {
                    fmt.Println(logtag,err);
                    fmt.Println(logtag,"exiting")
                    //getUnlink(id)
                    return
                }
		        fmt.Println(logtag,string(v.str))
                nosend = 0
            } else {
                nosend++
            }
		    time.Sleep(50*time.Millisecond);
            lastv=v.str;
        }
	}
}

func sendVal(id string,mess []byte) {
    mux.Lock()
    messages[id]=msg{mess}
    mux.Unlock()
}

func getVal(gid string) (m msg){
    mux.Lock()
    if sid,ok := links[gid]; ok {
        if m,ok = messages[sid]; ok {
            mux.Unlock()
            return
        }
    }
    mux.Unlock()
    m = msg{str:[]byte("NULL")}
    return
}

func crossLink(gid string,sid string) {
    mux.Lock()
    if( gid != "" ) {
        links[gid]=sid
    }
    mux.Unlock()
    return
}

func sendUnlink(gid string) {
    mux.Lock()
    delete(messages,gid)
    mux.Unlock()
    return
}

func getUnlink(gid string) {
    mux.Lock()
    delete(links,gid)
    mux.Unlock()
    return
}

func main() {

    for _, pair := range os.Environ() {
        fmt.Println(pair)
    }

	v := msg{ []byte("NULL") }
	loadstorevar.Store(v)

	http.HandleFunc("/send", sendSession)
	http.HandleFunc("/get", getSession)
	http.Handle("/", http.FileServer(http.Dir("/app/html/")))

	fmt.Printf("Starting \n");
	err := http.ListenAndServe(":"+os.Args[1], nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

