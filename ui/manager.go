package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type wrapper struct {
	Type string          `json:"typ"`           // The type of the json message
	Msg  json.RawMessage `json:"msg,omitempty"` // The full body of the received message
	Who  *who            `json:"who,omitempty"` // info for forwarded consumers
}

type who struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

func (w *wrapper) attach(id, kind string) { w.Who = &who{id, kind} }

var approvedTypes = map[string]bool{
	"admin": true,
}

type socketWrapper struct {
	ty string
	ws *websocket.Conn
	// TODO: add a channel to safely send messages to in threads
}

type manager struct {
	conz map[string]socketWrapper
	mutx sync.RWMutex

	servers map[string]string // hostname -> kind
	smux    sync.RWMutex
}

func newManager() (*manager, error) {
	return &manager{
		conz:    make(map[string]socketWrapper, 10),
		servers: make(map[string]string),
	}, nil
}

func (man *manager) register(ws *websocket.Conn) (id, kind string, cancel func(), err error) {
	kind = ws.Request().URL.Path[4:] // len("/ws/")
	if !approvedTypes[kind] {
		return id, kind, nil, errors.New("bad ws type: '" + kind + "'")
	}
	id = kind + "-" + ws.RemoteAddr().String() + "-" + time.Now().String()
	id = fmt.Sprintf("%X", md5.Sum([]byte(id)))
	log.Println("Registering:", id)
	man.mutx.Lock()
	man.conz[id] = socketWrapper{
		ty: kind,
		ws: ws,
	}
	man.mutx.Unlock()
	return id, kind, func() {
		// TODO: send admin disconnet info
		log.Println("Unregistering:", id)
		man.mutx.Lock()
		delete(man.conz, id)
		man.mutx.Unlock()
	}, nil
}

func (man *manager) Handle(ws *websocket.Conn) {
	id, kind, death, err := man.register(ws)
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
		wrap.attach(id, kind) // Add metadata for forwarding situations
		switch wrap.Type {
		case "ping":
			if err := websocket.JSON.Send(ws, &wrapper{Type: "pong"}); err != nil {
				log.Println("Pong err", ws.RemoteAddr(), err)
				break OUTER // TODO: better error handling
			}
		case "stat":
			if err := man.sendTo("admin", wrap); err != nil {
				log.Println("Stat err", ws.RemoteAddr(), err)
				break OUTER // TODO: better error handling
			}
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

func (man *manager) sendTo(kind string, wrap wrapper) error {
	man.mutx.RLock()
	list := man.conz
	man.mutx.RUnlock()
	// TODO: parallel send these requests
	for _, socket := range list {
		if socket.ty == kind {
			// TODO: send these to a threadsafe channel
			if err := websocket.JSON.Send(socket.ws, &wrap); err != nil {
				return err
			}
		}
	}
	return nil
}

func (man *manager) Hello(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	kind := r.Form.Get("kind")
	name := r.Form.Get("name")
	if kind == "" || name == "" {
		http.Error(w, "invalid name or kind", http.StatusExpectationFailed)
		return
	}
	man.smux.Lock()
	man.servers[name] = kind
	man.smux.Unlock()
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Ready to Poll!"))
	log.Printf("Registering new %s: %q", kind, name)
}

func (man *manager) poll() {
	for range time.Tick(time.Second) {
		man.smux.RLock()
		// TODO: only clone if there is a difference
		clone := make(map[string]string, len(man.servers))
		for key, value := range man.servers {
			clone[key] = value
		}
		man.smux.RUnlock()

		// TODO: fanout request all servers metrics

		log.Println("TODO: poll all the servers for statuses and broadcast to admin")
		// TODO: update admins with useful information
	}
}
