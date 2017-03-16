package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type wrapper struct {
	Type string          `json:"typ"` // The type of the json message
	Msg  json.RawMessage `json:"msg"` // The full body of the received message
}

var approvedTypes = map[string]bool{
	"admin":  true,
	"client": true,
	"origin": true,
	"server": true,
}

type socketWrapper struct {
	ty string
	ws *websocket.Conn
}

type manager struct {
	conz map[string]socketWrapper
	mutx sync.RWMutex
}

func newManager() (*manager, error) {
	return &manager{
		conz: make(map[string]socketWrapper, 10),
	}, nil
}

func (man *manager) register(ws *websocket.Conn) (string, func(), error) {
	kind := ws.Request().URL.Path[4:] // len("/ws/")
	if !approvedTypes[kind] {
		return kind, nil, errors.New("bad ws type: '" + kind + "'")
	}
	id := kind + "-" + ws.RemoteAddr().String() + "-" + time.Now().String()
	id = fmt.Sprintf("%X", md5.Sum([]byte(id)))
	log.Println("Registering:", id)
	man.mutx.Lock()
	man.conz[id] = socketWrapper{
		ty: kind,
		ws: ws,
	}
	man.mutx.Unlock()
	return kind, func() {
		log.Println("Unregistering:", id)
		man.mutx.Lock()
		delete(man.conz, id)
		man.mutx.Unlock()
	}, nil
}

func (man *manager) Handle(ws *websocket.Conn) {
	_, death, err := man.register(ws)
	if err != nil {
		log.Println("bad register: ", err)
		return
	}
	defer death()
OUTER:
	for {
		var wrap wrapper
		if err := websocket.JSON.Receive(ws, &wrap); err != nil {
			log.Println("Receive", err)
			break
		}
		switch wrap.Type {
		case "ping":
			if err := websocket.JSON.Send(ws, wrapper{Type: "pong"}); err != nil {
				log.Println("Pong err", ws.RemoteAddr(), err)
				break OUTER
			}
		case "stat":
			// TODO: stat
			log.Println("STAT", string(wrap.Msg))
		default:
			// TODO: default
			log.Println("Unknown Type: ", wrap.Type)
		}
		// log.Println("Receive Message:", m.Message)
		// m = message{"Thanks!"}
		// if err := websocket.JSON.Send(ws, m); err != nil {
		// 	log.Println("Send", err)
		// 	break
		// }
	}
}
