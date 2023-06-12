
//
//  todo
//      - when hub is disconnected messages build up on the event_queue
//          consider discard all but last minute and no duplicates
//      - automated testing
//      - rewrite to be horizoltally scalable (no hub)
//      - identify listner by ip address//port or hostname?
//      - consider epoc time to ensure older messages not transmitted or acted upon
//      - transmit 'alive' message from hub to each listnere every 5 secs using channels
//      - on connection with hub a listener should dump it's database to the hub
//          this will mean that a restarted hub will rebuild it's database from listers
//          potentially need to 'ask'? new listernbs are empty so easy
//          special case is when hub restarts and wants all the info, but not duplicates.
//          maybe the hub responds witha recap??
//

package main

import (
    "bytes"
    "log"
    "sync"
    "strings"
    "net"
    "io"
    "bufio"
    "os"
	"time"
	"net/http"
	"github.com/gorilla/websocket"
    "github.com/timtadh/getopt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

func sendSession(w http.ResponseWriter, r *http.Request) {

    sid:=r.URL.Query().Get("s");
    gid:=r.URL.Query().Get("g");
    logtag:=sid+":  in: ";
    log.Print(logtag,"started: adding in client ",gid)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
        log.Print(logtag,err);
        log.Print(logtag,"exiting")
        return
    }

    crossLink(gid,sid,true)

	for {
		_, rec_mess, err := conn.ReadMessage()
		if err != nil {
            log.Print(logtag,err);
            log.Print(logtag,"exiting")
            sendUnlink(sid)
            return
        }
        saveMessage(sid,rec_mess,true)
	}
}

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

func saveMessage(id string,mess []byte,shareit bool) {
	log.Print("saveMessage: send,"+id+string(mess))
    m := message{value:mess}
    m.mtime=time.Now().Unix()
    mux.Lock()
    morig := messages[id]
    messages[id]=m
    mux.Unlock()
    if string(m.value) != string(morig.value) && shareit {
        event_queue <- "send,"+id+","+string(m.value)
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

func crossLink(getter_id string,sender_id string,shareit bool) {
    if( getter_id != "" ) {
	    log.Print("crossLink link,"+getter_id+","+sender_id)
        s := link{value:sender_id}
        s.mtime=time.Now().Unix()
        mux.Lock()
        links[getter_id] = s
        mux.Unlock()
        if shareit {
            event_queue <- "link,"+getter_id+","+sender_id
        }
    }
}

func sendUnlink(gid string) {
    mux.Lock()
    delete(messages,gid)
    mux.Unlock()
}

func monitor() {
    for {
        time.Sleep(60000*time.Millisecond);
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

    // main loop, includes making connection to server so that if connection lost
    // the will attempt reconnection every 5 seconds, under normal operation we
    // would hoe the loop would only be executed once.
    for {

        // give daemon chance to start if started at same time
	    time.Sleep(2*sec);

        // connect to server port
        c, err := net.Dial("tcp", dbServ)
        if err != nil {
            log.Print("share_client: Dial failed: ", err.Error())
	        time.Sleep(3*sec);
            continue
        }
        log.Print("share_client: connected to ", dbServ )

        stop := make(chan bool)

        // at start of a new connection share a list of all the getter/setter links
        // if just started this will be empty but if because we have lost hub the
        // refills the hub with info it neads about this client connection
        bootstring := ""
        mux.RLock()
        for  i := range active_getters {
            if _, ok := links[i] ; ok {
                bootstring = bootstring + "ag," + i + "," + links[i].value + "\n"
            }
        }
        mux.RUnlock()
        _, err = io.WriteString(c,bootstring)
        if err != nil {
            log.Print("share_client: failed writing boot string to hub: ", err.Error())
            c.Close()
            continue
        }

        // go routine loop where we share data sent via 'event_queue' with remote database
        go func() {
            log.Print("share_client: waiting on new messages")
            for {

                s:=""
                select {
                    case s = <-event_queue:
                    case <-stop:
                        break
                }

                _, err = io.WriteString(c,s+"\n")
                if err != nil {
                    log.Print("share_client: Write to server failed: ", err.Error())
                    break
                }
            }

            c.Close()
            stop <- true
        }()

        // go routine loop where we share data sent via 'event_queue' with remote database
        go func() {

            input := bufio.NewScanner(c)
            for input.Scan() {
                a := strings.SplitN(input.Text(),",",3)
                if len(a) != 3 {
                    continue
                }
                if a[0] == "send" {
                    saveMessage(a[1],[]byte(a[2]),false)
                } else if a[0] == "link" {
                    crossLink(a[1],a[2],false)
                }
                log.Print("share_client: Read: ",input.Text())
            }

            c.Close()
            stop <- true
        }()

        <-stop
        <-stop
        log.Print("share_client: stopped")

    }

}

func share_daemon(dbPort string) {

    //tcpAddr, err := net.ResolveTCPAddr("tcp4",dbPort)
    //if err != nil {
        //log.Print("share_daemon: ResolveTCPAddr: ", err.Error())
        //os.Exit(30)
    //}
    //log.Print("share_daemon: ResolveTCPAddr: ",tcpAddr)

    listener, err := net.Listen("tcp4",dbPort)
    if err != nil {
        log.Print("share_daemon: Listen: ", err.Error())
        select { }
    }
    log.Print("share_daemon: Listen now active: ",listener)

    // accept connections from share clients
    // we will 
    go func() {
        for {
            conn, err := listener.Accept()
            if err != nil {
                log.Print("share_daemon: Listener.Accept(): ", err.Error())
                os.Exit(32)
            }
            log.Print("share_daemon: Accept: ",listener)
            go share_daemon_peer(conn)
        }
    }()

}

func share_daemon_peer(c net.Conn) {

    // uniquely identify this client connection 
    // and create a channel to listen on for 
    peerChan := make(chan string,1000)
    mux.Lock()
    daemon_id = daemon_id + 1
    peerID := daemon_id
    daemon_chans[peerID] = peerChan
    mux.Unlock()

    stop := make(chan bool)

    // go routine to listen for notifications on newly created channel
    go func() {
        log.Print("share_daemon_peer: listening on peerChan: ",peerID)
        for {
            s:=""
            select {
                case s = <-peerChan:
                case <-stop:
                    break
                }

            _, err := io.WriteString(c,s+"\n")
            if err != nil {
                log.Print("share_client: Write to server failed: ", err.Error())
                break
            }
        }
        c.Close()
        stop <- true
    }()

    // go routine to recived client messages and share with other registered clients 
    go func() {
        log.Print("share_daemon_peer: waiting on Network input: ",peerID)
        input := bufio.NewScanner(c)
        for input.Scan() {
            if input.Text() != "" {
                a := strings.SplitN(input.Text(),",",3)
                if len(a) != 3 {
                    log.Print("share_daemon_peer: BAD COMS(1): ",input.Text())
                    continue
                }
                if a[0] == "new_getter" {
                    // if a new getter has a link stored on the hub share 
                    // back to the connected carrier
                    mux.RLock()
                    if v, ok := links[a[1]] ; ok {
                        peerChan <- "link,"+a[1]+","+v.value
                    }
                    mux.RUnlock()
                    continue
                } else if a[0] == "ag" {
                    // record 'ag' (activate getters) but don't pass them on
                    // to other carriers
                    crossLink(a[1],a[2],false)
                    continue
                } else if a[0] == "link" {
                    // record the link information and let it drop through to be
                    // passed out to all other carriers
                    crossLink(a[1],a[2],false)
                } else if a[0] != "send" {
                    // if not send that not a known request so marked as bad
                    // otherwise just let it be passed to other carriers
                    log.Print("share_daemon_peer: BAD COMS(2): ",input.Text())
                    continue
                }
                mux.RLock()
                for i, xchan := range daemon_chans {
                if i != peerID {
                        xchan <- input.Text()
                    }
                }
                mux.RUnlock()
            }
        }
        c.Close()
        stop <- true
    }()

    <-stop
    <-stop
    log.Print("share_daemon_peer: closed peer: ",peerID)
}

func main() {

	short := "l:m:h:"
	long := []string{ "listen","manager","hub" }

	listenPort := "6001"
    managerPort := "7777"
    hubURL := "localhost:7777"

    otherargs, optargs, err := getopt.GetOpt(os.Args[1:], short, long)
	for _, oa := range optargs {
		switch oa.Opt() {
		case "-l", "--listen":
			listenPort = oa.Arg()
		case "-m", "--manager":
			managerPort = oa.Arg()
		case "-h", "--hub":
			hubURL = oa.Arg()
		}
	}

    if len(otherargs) > 0 {
        runtype = otherargs[0]
    } else {
        runtype = "standalone"
    }
	log.Print("runtype: ",runtype)
    if runtype == "hub"  {
	    log.Print("Hub Port: ",managerPort);
        go share_daemon(":"+managerPort)
    } else if runtype == "carrier" {
	    log.Print("Listening on Port: ",listenPort);
	    log.Print("Manager on Port: ",managerPort);
	    log.Print("Hub Location: ",hubURL);
        go share_client(hubURL)
        go share_daemon(":"+managerPort)
    } else if runtype == "standalone" {
	    log.Print("Listening on Port: ",listenPort);
        go share_client(hubURL)
    } else {
	    log.Print("need to specify a run type of hub, carrier, or standalone");
        os.Exit(1)
    }

    // to be used for background monitoring
    go monitor()

    if runtype == "hub" {
        select { } // pause and just run daemon only
    } else {
	    http.HandleFunc("/send", sendSession)
	    http.HandleFunc("/get", getSession)
	    http.Handle("/", http.FileServer(http.Dir("./html/")))

	    log.Print("Starting \n");
	    err2 := http.ListenAndServe(":"+listenPort, nil)
	    if err2 != nil {
		    panic("ListenAndServe: " + err.Error())
	    }
    }
}

