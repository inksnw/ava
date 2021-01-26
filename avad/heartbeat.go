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

func ping() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		ch := wsConns.IterBuffered()
		for item := range ch {
			host := item.Key
			ws, ok := item.Val.(*websocket.Conn)
			if !ok {
				reconnect(host)
				continue
			}

			ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(core.PongWait)); return nil })
			err := ws.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Error().Msgf("节点 %s的ws心跳检测失败,重新连接 %s", host, err)
				reconnect(host)
				continue
			}

			status, _ := tcpStatus.Get(host)
			if !status.(bool) {
				log.Error().Msgf("节点 %s tcp中断,重新连接", host)
				addrTcp := strings.Join([]string{host, ":", core.TcpPort}, "")
				go dialTcp(addrTcp)
			}

			//log.Debug().Msgf("节点 %s的ws心跳检测正常", host)
		}
		<-ticker.C
	}
}

func reconnect(host string) {
	addrWs := strings.Join([]string{host, ":", core.WsPort}, "")
	addrTcp := strings.Join([]string{host, ":", core.TcpPort}, "")
	go dialWs(addrWs)
	go dialTcp(addrTcp)
}
