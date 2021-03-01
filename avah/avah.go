package avah

import (
	"ava/core"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/phuslu/log"
	"github.com/spf13/viper"
	"net/http"
	"strings"
	"sync"
)

var connIns *core.WsStruct

func updateInfo() {
	go sendMsg()
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Info().Msgf("配置文件发生变动,更新信息")
		loadConfig(".")
	})
}

func dial(w http.ResponseWriter, r *http.Request) {
	var err error
	connIns = &core.WsStruct{}
	Upgrader := websocket.Upgrader{} // use default options
	var mutex sync.Mutex
	connIns.Mutex = &mutex
	connIns.Conn, err = Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Msgf("ws握手失败: %s", err)
		return
	}
	log.Info().Msgf("接到管理端ws连接")
	loadConfig(".")
	updateInfo()
	defer connIns.Conn.Close()

	p := core.TaskMsg{}
	for {
		err := connIns.RJson(&p)
		if err != nil {
			log.Error().Msgf("读取数据失败,管理节点可能已关闭 %s", err)
			break
		}
		//接收信息,给到路由分发
		taskRouter(p)
	}
}

func Node() {

	go listenTcp()

	addr := strings.Join([]string{"0.0.0.0", ":", core.WsPort}, "")
	http.HandleFunc("/ws", dial)
	log.Info().Msgf("ws监听地址: %s ", addr)
	http.ListenAndServe(addr, nil)
}
