
//
//  HALLOWEEN
//  work by single token
//      - allow multiple listeners and single controller
//

package main

import (
    "bytes"
    "log"
    "sync"
//    "strings"
//    "net"
//    "io"
//    "bufio"
//    "os"
	"time"
	"net/http"
	"github.com/gorilla/websocket"
//    "github.com/timtadh/getopt"
)

// track latest sender messages
type message struct {
    value  []byte
    mtime  int64
}
var messages = make(map[string]message)

// track links between sender and getter
type link struct {
    value  string
    mtime  int64
}
var links = make(map[string]link)

var runtype = "none"

// sharing client config
var event_queue = make(chan string,1000)
// sharing server config
var daemon_chans = make(map[int](chan string))
var daemon_id    = 0;

//simple Read/Write mux 
var mux = &sync.RWMutex{}

// track active getter list
var active_getters = make(map[string]string)

// upgrader used to convert https to wss connections
var upgrader = websocket.Upgrader{
    ReadBufferSize:  10240,
    WriteBufferSize: 10240,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

//================================================================

func sendSession(w http.ResponseWriter, r *http.Request) {

    id:=r.URL.Query().Get("id");
    logtag:=id+":  in: ";

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
        log.Print(logtag,err);
        log.Print(logtag,"exiting")
        return
    }

	for {
		_, rec_mess, err := conn.ReadMessage()
		if err != nil {
            log.Print(logtag,err);
            log.Print(logtag,"exiting")
            return
        }
	    log.Print("saveMessage: send,"+id+": "+string(rec_mess))
        event_queue <- "send,"+id+","+string(rec_mess)
	}
}

//================================================================

func getSession(w http.ResponseWriter, r *http.Request) {

	//var m []byte("hello")
    id:=r.URL.Query().Get("g");
    logtag:=id+": out: ";
    log.Print(logtag,"started")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
        log.Print(logtag,err);
        log.Print(logtag,"exiting")
        return
    }

    event_queue <- "new_getter,"+id+","
    mux.Lock()
    active_getters[id]=""
    mux.Unlock()

    lastv := []byte{}
    nosend := 0
	for {
        get_mess := getMessage(id)
        if( !bytes.Equal(get_mess, lastv) || nosend > 100 ) {
	        var err = conn.WriteMessage(1,get_mess)
	        if err != nil {
                log.Print(logtag,err);
                log.Print(logtag,"exiting")
                mux.Lock()
                delete(active_getters,id)
                mux.Unlock()
                return
            }
	        log.Print(logtag,string(get_mess))
            nosend = 0
        } else {
            nosend++
        }
	    time.Sleep(50*time.Millisecond);
        lastv=get_mess;
	}
}

func getMessage(gid string) (mess []byte){
    mux.RLock()
    if sid,ok := links[gid]; ok {
        if m,ok := messages[sid.value]; ok {
            mux.RUnlock()
            mess = m.value
            return
        }
    }
    mux.RUnlock()
    mess = []byte("NULL")
    return
}

func monitor() {
    for {
        time.Sleep(5000*time.Millisecond);
        log.Print("event_queue:",len(event_queue)," daemon_chans:",len(daemon_chans));
    }
}

func share_client(dbServ string) {

    sec :=1000*time.Millisecond;

    // infinited loop to attempt reconnectying every sleep_ms 
    // normal operation is to stay connected but loop enables recovery

    if runtype == "standalone" {
        for {
            <-event_queue
        }
        // just consume queue thrwoing it away when in stadnalone mode
    }

    // to ensure that at least one event is sent every 5 secs
    // (resulting in network traffic, which ensures network
    // remains tested and operational)
    go func() {
        for {
            time.Sleep(5*sec)
            if len(event_queue) == 0  {
                // chan len is used to avoid pointless event build up when
                // no network connection.
                event_queue <- ""
            }
        }
    }()


}

func main() {

	listenPort := "7777"
    // managerPort := "7777"
    hubURL := "localhost:7777"

	log.Print("Listening on Port: ",listenPort);

    go share_client(hubURL)   // loop managing message passing
    go monitor()              // background monitoring routine

	http.HandleFunc("/send", sendSession)                   // send session handler
	http.HandleFunc("/get", getSession)                     // get session handler
	http.Handle("/", http.FileServer(http.Dir("./html/")))  // http file server

	err2 := http.ListenAndServe(":"+listenPort, nil)     // LISTEN 

    // if drop out of ListenAndServe then stop
	if err2 != nil {
	    panic("ListenAndServe: " + err2.Error())
	}

}

