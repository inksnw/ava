package avah

import (
	"ava/core"
	"context"
	"encoding/json"
	"flag"
	"github.com/hashicorp/yamux"
	"github.com/phuslu/log"
	"golang.org/x/net/proxy"
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
		log.Info().Msgf("接收到管理端tcp连接 %s", agentStr)
		_ = conn.SetDeadline(time.Time{})
		session, err := yamux.Client(conn, nil)
		if err != nil {
			log.Error().Msgf("建立yamux session失败 %s", err)
		}
		go listenSocks(session, agentStr)

	}

}

func testSocks5(ctx context.Context) {
	time.Sleep(5 * time.Second)
	ticker := time.NewTicker(5 * time.Minute)
	type result struct {
		Msg string `json:"msg"`
	}
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("socks5测试协程退出")
			return
		case <-ticker.C:
			dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:4562", nil, proxy.Direct)
			if err != nil {
				log.Error().Msgf("无法连接本机代理", err)
				continue
			}
			httpTransport := &http.Transport{}
			httpClient := &http.Client{Transport: httpTransport, Timeout: 2 * time.Second}
			httpTransport.Dial = dialer.Dial

			if resp, err := httpClient.Get("http://127.0.0.1:4000/socks5"); err != nil {
				log.Error().Msgf("socks5连接异常 %s", err.Error())
			} else {
				rv := result{}
				json.NewDecoder(resp.Body).Decode(&rv)
				log.Info().Msgf("测试socks5: %s", rv.Msg)
				resp.Body.Close()
			}
		}
	}
}

// Catches local clients and connects to yamux

func listenSocks(session *yamux.Session, agentStr string) {
	var err error
	address := strings.Join([]string{"127.0.0.1", ":", core.SocksPort}, "")

	if !lis {
		socksListen, err = net.Listen("tcp", address)
		if err != nil {
			log.Error().Msgf("本地socks5端口监听失败 %s", err)
			panic(err)
		}
		lis = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	go testSocks5(ctx)
	defer cancel()

	for {
		conn, err := socksListen.Accept()
		if err != nil {
			log.Error().Msgf("[%s] 接收本地请求失败 %s: %v", agentStr, address, err)
			continue
		}

		stream, err := session.Open()
		if err != nil {
			log.Error().Msgf("[%s] 开启session id:%d 失败 %v", agentStr, session.NumStreams(), err)
			err = session.Close()
			if err != nil {
				log.Error().Msgf("session %d 关闭失败 %s", session.NumStreams(), err)
			}
			return
		}

		go func() {
			io.Copy(conn, stream)
			conn.Close()
			log.Printf("[%s] Done copying conn to stream for %s", agentStr, conn.RemoteAddr())
		}()
		go func() {
			io.Copy(stream, conn)
			stream.Close()
			log.Printf("[%s] Done copying stream to conn for %s", agentStr, conn.RemoteAddr())
		}()
	}

}
