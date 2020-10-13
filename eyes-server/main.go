
package main

import (
	"fmt"
    "bytes"
    "sync"
    "net"
    "net/textproto"
    "os"
	"time"
	"net/http"
	"github.com/gorilla/websocket"
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

// sharing client config
var share_queue = make(chan string,1000)
// sharing server config
var dbServ = "localhost:7777"
var dbPort = ":7777"
var daemon_chans = make(map[int](chan message))
var daemon_id    = 0;

//simple mux 
var mux = &sync.Mutex{}

// upgrader used to convert https to wss connections
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
	fmt.Println(logtag,"crossLink sender=",sid," getter=",gid)

	for {
		_, rec_mess, err := conn.ReadMessage()
		if err != nil {
            fmt.Println(logtag,err);
            fmt.Println(logtag,"exiting")
            sendUnlink(sid)
            return
        }
        saveMessage(sid,rec_mess)
		fmt.Println(logtag,string(rec_mess))
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
        get_mess := getMessage(id)
        if( !bytes.Equal(get_mess, lastv) || nosend > 100 ) {
	        var err = conn.WriteMessage(1,get_mess)
	        if err != nil {
                fmt.Println(logtag,err);
                fmt.Println(logtag,"exiting")
                //getUnlink(id)
                return
            }
	        fmt.Println(logtag,string(get_mess))
            nosend = 0
        } else {
            nosend++
        }
	    time.Sleep(50*time.Millisecond);
        lastv=get_mess;
	}
}

func saveMessage(id string,mess []byte) {
    m := message{value:mess}
    m.mtime=time.Now().Unix()
    mux.Lock()
    messages[id]=m
    mux.Unlock()
    share_queue <- id+","+string(m.value)
}

func getMessage(gid string) (mess []byte){
    mux.Lock()
    if sid,ok := links[gid]; ok {
        if m,ok := messages[sid.value]; ok {
            mux.Unlock()
            mess = m.value
            return
        }
    }
    mux.Unlock()
    mess = []byte("NULL")
    return
}

func crossLink(getter_id string,sender_id string) {
    if( getter_id != "" ) {
        s := link{value:sender_id}
        s.mtime=time.Now().Unix()
        mux.Lock()
        links[getter_id] = s
        mux.Unlock()
    }
}

func sendUnlink(gid string) {
    mux.Lock()
    delete(messages,gid)
    mux.Unlock()
}

//func getUnlink(gid string) {
//    mux.Lock()
//    delete(links,gid)
//    mux.Unlock()
//}

func stats_out() {
    for {
        time.Sleep(2000*time.Millisecond);
        fmt.Println("share_queue",len(share_queue))
    }
}

func share_client() {

    sec :=1000*time.Millisecond;

    // infinited loop to attempt reconnectying every sleep_ms 
    // normal operation is to stay connected but loop enables recovery
    for {

        // generate remote TCP Address
        tcpAddr, err := net.ResolveTCPAddr("tcp4", dbServ )
        if err != nil {
            fmt.Println("share_client: ResolveTCPAddr:", err.Error())
	        time.Sleep(5*sec)
            continue
        }

        // give daemon chance to start if started at same time
	    time.Sleep(2*sec);

        // connect to server port
        conn, err := Dial("tcp", tcpAddr)
        if err != nil {
            fmt.Println("share_client: Dial failed:", err.Error())
	        time.Sleep(5*sec);
            continue
        }
        fmt.Println("share_client: connected to", dbServ )

        stop := make(chan bool)

        // go routine loop where we share data sent via 'share_queue' with remote database
        go func() {
            fmt.Println("share_client: waiting on new messages")
            for {

                s:=""
                select {
                    case s = <-share_queue:
                        fmt.Println("share_client: received message")
                    case <-stop:
                        fmt.Println("share_client: Writer recieved stop request")
                        break
                }

                _, err = conn.Write( []byte(s+"\n") )
                if err != nil {
                    fmt.Println("share_client: Write to server failed:", err.Error())
                    break
                }
            }

            conn.Close()
            stop <- true
        }()

        // go routine loop where we share data sent via 'share_queue' with remote database
        go func() {

            // read single byte  
            // read remaining bytes
            // store message with locking 
            // loop
            fmt.Println("Listner recived >stop<")

        }()

        <- stop

    }

}

func share_daemon() {

    //tcpAddr, err := net.ResolveTCPAddr("tcp4",dbPort)
    //if err != nil {
        //fmt.Println("share_daemon: ResolveTCPAddr:", err.Error())
        //os.Exit(30)
    //}
    //fmt.Println("share_daemon: ResolveTCPAddr:",tcpAddr)

    listener, err := net.Listen("tcp4",dbPort)
    if err != nil {
        fmt.Println("share_daemon: Listen:", err.Error())
        os.Exit(31)
    }
    fmt.Println("share_daemon: Listen:",listener)

    // accept connections from share clients
    // we will 
    go func() {
        for {
            conn, err := listener.Accept()
            if err != nil {
                fmt.Println("share_daemon: Listener.Accept():", err.Error())
                os.Exit(32)
            }
            fmt.Println("share_daemon: ListenTCP:",listener)
            go share_daemon_peer(conn)
        }
    }()

}

func share_daemon_peer(c net.Conn) {

    // uniquely identify this client connection 
    // and create a channel to listen on for 
    peerChan := make(chan message,1000)
    mux.Lock()
    daemon_id = daemon_id + 1
    peerID := daemon_id
    daemon_chans[peerID] = peerChan
    mux.Unlock()

    // go routine to listen for notifications on newly created channel
    go func() {
        fmt.Println("share_daemon_peer: listening on peerChan:",peerID)
        for {
            m := <- peerChan
            fmt.Println("peerID",peerID,"message",m,"NEED ACTION TO BE CODED");
        }
    }()

    // go routine to recived client messages and share with other registered clients 
    go func() {
        fmt.Println("share_daemon_peer: waiting on Network input:",peerID)
        for {
            n, err := c.Read(b)
            if (err != nil ) {
                fmt.Println("share_daemon_peer: netwrok read:",peerID,":Read failed:", err.Error())
                break
            }
            fmt.Println("share_daemon_peer: processing:",n,b)
            if (n>0) {
                cached := append(cached,b...)
                for {
                    mc := int(cached[0])
                    if( len(cached) > mc ) {
                        fmt.Println("share_daemon_peer: netwrok read:",peerID,mc,":",string(cached[1:mc]))
                        cached = cached[mc:]
                    } else  {
                        break
                    }
                }
            }
        }
    }()

    select { }
}

func main() {

    for _, pair := range os.Environ() {
        fmt.Println(pair)
    }

    // 
    go stats_out()
    go share_client()
    go share_daemon()

	http.HandleFunc("/send", sendSession)
	http.HandleFunc("/get", getSession)
	http.Handle("/", http.FileServer(http.Dir("/app/html/")))

	fmt.Printf("Starting \n");
	err := http.ListenAndServe(":"+os.Args[1], nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

