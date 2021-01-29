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
		ch := ConnStatus.IterBuffered()
		for item := range ch {
			host := item.Key
			ins, _ := item.Val.(*ConnStruct)
			if !ins.status {
				reconnect(host, ins)
				continue
			}

			ping := func(string) error {
				_ = ins.conn.SetReadDeadline(time.Now().Add(core.PongWait))
				return nil
			}

			ins.conn.SetPongHandler(ping)
			err := ins.conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Error().Msgf("节点 %s心跳检测失败,重新连接 %s", host, err)
				reconnect(host, ins)
				continue
			}
			log.Debug().Msgf("节点 %s的ws心跳检测正常", host)
		}
		<-ticker.C
	}
}

func reconnect(host string, ins *ConnStruct) {
	if ins.conn != nil {
		err := ins.conn.Close()
		if err != nil {
			log.Debug().Msgf("尝试关闭上一个ws连接失败 %s", err)
		}

	}

	addrWs := strings.Join([]string{host, ":", core.WsPort}, "")
	addrTcp := strings.Join([]string{host, ":", core.TcpPort}, "")
	go dialWs(addrWs, ins)
	go dialTcp(addrTcp, ins)
}
