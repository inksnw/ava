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

func DialConn(address, addrWs string, ins *core.WsStruct) {

	host := strings.Split(addrWs, ":")[0]
	u := url.URL{Scheme: "ws", Host: addrWs, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error().Msgf("连接节点ws通道%s失败,%s后重试:", addrWs, core.PongWait)
		return
	}
	ins.Conn = c
	log.Info().Msgf("已连接节点ws通道%s", addrWs)
	go getNodeInfo(host, ins)

	var session *yamux.Session
	server, _ := socks5.New(&socks5.Config{})
	host = strings.Split(address, ":")[0]
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Error().Msgf("连接远端tcp通道%s失败,%s后重试", address, core.PongWait)
		return
	}
	ins.Status = true
	log.Info().Msgf("已连接节点tcp通道%s", address)
	session, err = yamux.Server(conn, nil)
	if err != nil {
		//todo 这里的处理好像还有坑
		session.Close()
		panic(err)
	}
	err = relay(host, session, server)
	if err != nil {
		log.Error().Msgf("转载tcp通道%s失败,%s后重试", address, core.PongWait)
		conn.Close()
		//Reconnect(ins, host)
		errChan <- host
	}
}

func relay(host string, session *yamux.Session, server *socks5.Server) (err error) {
	for {
		stream, err := session.Accept()
		if err != nil {
			log.Error().Msgf("公网节点无法连接%s可能已经关闭,立刻重连 %s", host, err)
			return err
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
