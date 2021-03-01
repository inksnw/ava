package core

import (
	"github.com/gorilla/websocket"
	"sync"
)

type WsStruct struct {
	Status bool
	Conn   *websocket.Conn
	Mutex  *sync.Mutex
}

func (t *WsStruct) WJson(json interface{}) (err error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	err = t.Conn.WriteJSON(json)
	if err != nil {
		return err
	}
	return nil
}

func (t *WsStruct) RJson(json interface{}) (err error) {
	err = t.Conn.ReadJSON(json)
	if err != nil {
		return err
	}
	return nil
}

func (t *WsStruct) WMessage(messageType int, data []byte) (err error) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	err = t.Conn.WriteMessage(messageType, data)
	if err != nil {
		return err
	}
	return nil
}
