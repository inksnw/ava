package avah

import (
	"ava/core"
	"github.com/gorilla/websocket"
	"github.com/phuslu/log"
	"net/http"
	"strings"
	"time"
)

var upgrader = websocket.Upgrader{} // use default options

func update() {
	ticker := time.NewTicker(core.UpdateWait)
	defer ticker.Stop()
	for {
		tmp:=make(map[string]core.LauncherConf)
		tmp["info"]=core.LauncherConf{Version: "1.0"}
		taskchan <- tmp
		<-ticker.C
	}

}

func dial(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Debug().Msgf("ws握手失败: %s", err)
		return
	}
	go infoReg(c)
	listAll(".")
	taskchan <- allConfig
	go update()
	defer c.Close()

	p := core.TaskMsg{}
	for {
		err := c.ReadJSON(&p)
		if err != nil {
			log.Debug().Msgf("读取数据失败,管理节点可能已关闭")
			break
		}
		//接收信息,给到路由分发
		taskrouter(p)
	}
}

func Node() {

	go listenForAgents()

	addr := strings.Join([]string{"0.0.0.0", ":", core.WsPort}, "")
	http.HandleFunc("/ws", dial)
	log.Debug().Msgf("ws监听地址: %s \n", addr)
	http.ListenAndServe(addr, nil)
}
