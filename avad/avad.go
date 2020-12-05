package avad

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/orcaman/concurrent-map"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

//var addrs = []string{"localhost:8080", "localhost:8081"}
var addrs = []string{"localhost:8080"}
var allconn = make(map[string]*websocket.Conn)
var wsStatus = cmap.New()

func DialWs(addr string) {
	wsStatus.Set(addr, false)
	for {
		if status, ok := wsStatus.Get(addr); ok {
			tmp := status.(bool)
			if !tmp {
				u := url.URL{Scheme: "ws", Host: addr, Path: "/echo"}
				c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					log.Printf("尝试连接节点ws通道%s失败,10s后重试:\n", addr)
					wsStatus.Set(addr, false)
					time.Sleep(10 * time.Second)
					continue
				}
				wsStatus.Set(addr, true)
				allconn[addr] = c
				fmt.Printf("已创建连接节点ws通道%s\n", addr)
			}
		}

	}
}



func DialS5(listenTarget string) {
	var RemoteConn net.Conn
	var err error
	for {
		for {
			RemoteConn, err = net.Dial("tcp", listenTarget)
			if err == nil {
				break
			}
		}
		go Handshake(RemoteConn)
	}
}


func DLocal() {
	//连接websocket
	for _, addr := range addrs {
		go DialWs(addr)
	}

	//连接内网穿透
	for _, addr := range addrs {
		host, _, _ := net.SplitHostPort(addr)
		addr = host + ":" + "18080"
		go DialS5(addr)
	}

	http.HandleFunc("/start", handel)
	addr := "localhost:4000"
	log.Fatal(http.ListenAndServe(addr, nil))

}

type Task struct {
	Route  string
	Cmd  string
	Args string
}

type rusult struct {
	Code int
	Msg  string
}

func handel(w http.ResponseWriter, r *http.Request) {
	var p Task
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	//todo 往哪个连接发应解析业务逻辑
	addr := p.Route
	c := allconn[addr]

	err = c.WriteJSON(p)
	msg := "投送成功"
	if err != nil {
		fmt.Println("节点消息投送失败,触发重新连接节点")
		//panic(err)
		msg = "投送失败,触发重新连接节点"
		//panic(err)
		wsStatus.Set(addr, false)

	}
	rv := rusult{
		Code: 200,
		Msg:  msg,
	}
	err = json.NewEncoder(w).Encode(rv)
	if err != nil {
		//... handle error
		panic(err)
	}

}