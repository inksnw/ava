package avad

import (
	"ava/core"
	"github.com/gorilla/websocket"
	"github.com/orcaman/concurrent-map"
	"github.com/phuslu/log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//ip--conn对应map
var allconn = make(map[string]*websocket.Conn)
var nodeTask = make(map[string]core.LauncherConf)
var wsStatus = cmap.New()
var tcpStatus = cmap.New()

func dialWs(addr string) {
	host := strings.Split(addr, ":")[0]

	wsStatus.Set(host, false)
	for {
		status, _ := wsStatus.Get(host)
		if !status.(bool) {
			u := url.URL{Scheme: "ws", Host: addr, Path: "/echo"}
			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Debug().Msgf("尝试连接节点ws通道%s失败,%s后重试:\n", addr, pongWait)
				time.Sleep(pongWait)
				continue
			}
			host := strings.Split(u.Host, ":")[0]
			wsStatus.Set(host, true)
			allconn[host] = c
			log.Debug().Msgf("已创建连接节点ws通道%s\n", addr)

			p := core.LauncherConf{}
			err = c.ReadJSON(&p)
			if err != nil {
				log.Debug().Msgf("读取节点: %s注册信息失败", host)
			}
			log.Debug().Msgf("接收节点: %s注册信息成功,可运行%s", host, p.Worker)
			nodeTask[host] = p
		}
	}

}



func DLocal(addrs []string) {

	for _, host := range addrs {
		tcpStatus.Set(host, false)
		wsStatus.Set(host, false)
	}

	go ping()

	for _, host := range addrs {
		addrWs := strings.Join([]string{host, ":", core.WsPort}, "")
		addrTcp := strings.Join([]string{host, ":", core.TcpPort}, "")
		go dialWs(addrWs)
		go connectForSocks(addrTcp)
	}

	http.HandleFunc("/exectask", handel)
	http.HandleFunc("/webWsConns", webWsConns)
	http.HandleFunc("/webWsStatus", webWsStatus)
	http.HandleFunc("/webNodeTask", webNodeTask)
	http.HandleFunc("/webTcpStatus", webTcpStatus)
	addr := strings.Join([]string{"localhost", ":", core.Web}, "")
	http.ListenAndServe(addr, nil)

}
