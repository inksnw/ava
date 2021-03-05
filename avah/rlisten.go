package avah

import (
	"ava/core"
	"flag"
	"github.com/hashicorp/yamux"
	"github.com/phuslu/log"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var socksListen net.Listener
var lis = false

func listenHttp() {
	http.HandleFunc("/log", serveHome)
	http.HandleFunc("/wss", serveWs)
	addr = flag.String("address", ":4564", "http service address")
	address := strings.Join([]string{"0.0.0.0", ":", "4564"}, "")
	log.Info().Msgf("http监听地址: %s ", address)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Error().Msgf("监听HTTP失败 %s", err)
	}
}

// listen for agents
func listenTcp() {
	address := strings.Join([]string{"0.0.0.0", ":", core.TcpPort}, "")
	log.Info().Msgf("tcp监听地址: %s ", address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Error().Msgf("监听地址失败 %s: %v", address, err)
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Msgf("接收管理端连接失败 %s", err)
			continue
		}

		agentStr := conn.RemoteAddr().String()
		log.Info().Msgf("接收到管理端tcp连接")
		_ = conn.SetDeadline(time.Time{})
		session, err := yamux.Client(conn, nil)
		if err != nil {
			log.Error().Msgf("建立yamux session失败 %s", err)
		}
		go listenSocks(session, agentStr)

	}

}

// Catches local clients and connects to yamux

func listenSocks(session *yamux.Session, agentStr string) {
	var err error
	address := strings.Join([]string{"0.0.0.0", ":", core.SocksPort}, "")

	if !lis {
		socksListen, err = net.Listen("tcp", address)
		if err != nil {
			log.Error().Msgf("本地socks5端口监听失败 %s", err)
			panic(err)
		}
		lis = true
	}

	for {
		conn, err := socksListen.Accept()
		if err != nil {
			log.Error().Msgf("[%s] 接收本地请求失败 %s: %v", agentStr, address, err)
			continue
		}

		stream, err := session.Open()
		if err != nil {
			log.Error().Msgf("[%s] 开启stream失败 %s: %v", agentStr, conn.RemoteAddr(), err)
			_ = session.Close()
			return
		}
		go relay(conn, stream)
	}

}

func relay(conn, stream net.Conn) {
	type res struct {
		N   int64
		Err error
	}
	ch := make(chan res)

	go func() {
		n, err := io.Copy(stream, conn)
		_ = stream.SetDeadline(time.Now()) // wake up the other goroutine blocking on stream
		_ = conn.SetDeadline(time.Now())   // wake up the other goroutine blocking on conn
		defer stream.Close()
		defer conn.Close()
		ch <- res{n, err}
	}()

	_, err := io.Copy(conn, stream)
	defer stream.Close()
	defer conn.Close()
	_ = stream.SetDeadline(time.Now()) // wake up the other goroutine blocking on stream
	_ = conn.SetDeadline(time.Now())   // wake up the other goroutine blocking on left
	rs := <-ch

	if err == nil {
		err = rs.Err
	}

}
