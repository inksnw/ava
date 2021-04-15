package avad

import (
	"ava/core"
	"github.com/gorilla/websocket"
	"github.com/phuslu/log"
	"strings"
	"time"
)

const (
	// Send pings to peer with this period. Must be less than PongWait.
	pingPeriod = (core.PongWait * 9) / 10
)

var errChan = make(chan string, 1024)

func ping() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	check()
	for {
		select {
		case <-ticker.C:
			check()
		case host := <-errChan:
			ins, _ := ConnStatus.Get(host)
			realIns := ins.(*core.WsStruct)
			log.Error().Msgf("节点 %s连接失败,重新连接", host)
			Reconnect(realIns, host)
		}
	}
}

func check() {
	ch := ConnStatus.IterBuffered()
	for item := range ch {
		host := item.Key
		ins, _ := item.Val.(*core.WsStruct)
		if !ins.Status {
			Reconnect(ins, host)
			continue
		}
		ping := func(string) error {
			_ = ins.Conn.SetReadDeadline(time.Now().Add(core.PongWait))
			return nil
		}
		ins.Conn.SetPongHandler(ping)
		err := ins.WMessage(websocket.PingMessage, []byte{})
		if err != nil {
			log.Error().Msgf("节点 %s心跳检测失败,重新连接 %s", host, err)
			Reconnect(ins, host)
			continue
		}
		log.Debug().Msgf("节点 %s的ws心跳检测正常", host)
	}
}

func Reconnect(wsIns *core.WsStruct, host string) {
	if wsIns.Conn != nil {
		err := wsIns.Conn.Close()
		if err != nil {
			log.Debug().Msgf("尝试关闭上一个ws连接失败 %s", err)
		}
	}
	wsIns.Status = false
	addrWs := strings.Join([]string{host, ":", core.WsPort}, "")
	addrTcp := strings.Join([]string{host, ":", core.TcpPort}, "")
	go DialConn(addrTcp, addrWs, wsIns)
}
