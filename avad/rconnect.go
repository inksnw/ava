package avad

import (
	"ava/core"
	"ava/core/go-socks5"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/phuslu/log"
	"net"
	"net/url"
	"strings"
)

func dialTcp(address string) {
	var session *yamux.Session
	server, _ := socks5.New(&socks5.Config{})
	host := strings.Split(address, ":")[0]
	tcpStatus.Set(host, false)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Error().Msgf("连接远端tcp通道 %s失败,%s后重试", address, core.PongWait)
		return
	}
	tcpStatus.Set(host, true)
	log.Debug().Msgf("已创建连接节点tcp反弹通道%s", address)

	session, err = yamux.Server(conn, nil)
	if err != nil {
		//todo 这里的处理好像还有坑
		session.Close()
		panic(err)
	}
	relay(host, session, server)
}

func relay(host string, session *yamux.Session, server *socks5.Server) {
	for {
		stream, err := session.Accept()
		if err != nil {
			log.Error().Msgf("公网节点无法连接%s可能已经关闭", host)
			tcpStatus.Set(host, false)
			break
		}
		log.Debug().Msgf("代理转发通信 %s %s", stream.LocalAddr(), stream.RemoteAddr())
		go func() {
			err = server.ServeConn(stream)
			if err != nil {
				log.Error().Err(err)
			}
		}()
	}

}

func dialWs(addr string) {
	host := strings.Split(addr, ":")[0]
	wsStatus.Set(host, false)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error().Msgf("尝试连接节点ws通道%s失败,%s后重试:\n", addr, core.PongWait)
		return
	}
	wsStatus.Set(host, true)
	wsConns.Set(host, c)
	log.Debug().Msgf("已连接节点ws通道%s\n", addr)
	go getNodeInfo(host, c)
}
