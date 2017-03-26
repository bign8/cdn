package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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

	data map[string]interface{}
	dmux sync.RWMutex
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
	port := r.Form.Get("port")
	if kind == "" || name == "" || port == "" {
		http.Error(w, "invalid name or kind or port", http.StatusExpectationFailed)
		return
	}
	man.smux.Lock()
	man.servers[name+":"+port] = kind
	man.smux.Unlock()
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Ready to Poll!"))
	log.Printf(host+": Registering new %s: %q", kind, name)
}

type metricsResult struct {
	host, kind string
	metrics    map[string]interface{}
	err        error
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

		// requesting from all remote servers
		fetch := func(host, kind string, done chan<- metricsResult) {
			target := "http://" + host + "/debug/metrics"
			r := metricsResult{host: host, kind: kind}
			if res, err := http.Get(target); err != nil {
				r.err = err
			} else if res.StatusCode != http.StatusOK {
				defer res.Body.Close()
				bits, _ := ioutil.ReadAll(res.Body)
				r.err = fmt.Errorf("fetch: bad response (%s): %q", res.Status, string(bits))
			} else {
				defer res.Body.Close()
				r.err = json.NewDecoder(res.Body).Decode(&r.metrics)
				for key := range r.metrics {
					if !strings.HasPrefix(key, kind) {
						delete(r.metrics, key)
					}
				}
			}
			done <- r
		}

		// fanout requests to all servers
		results := make(chan metricsResult, len(clone))
		for host, kind := range clone {
			go fetch(host, kind, results)
		}

		// get results
		data := make(map[string]interface{})
		data["uptime"] = time.Since(start).String()
		for i := 0; i < len(clone); i++ {
			back := <-results
			if back.err != nil {
				log.Println(host + ": Problem fetching stats from " + back.host + " " + back.kind)
			} else {
				for key, value := range back.metrics {
					data[key] = value
				}
			}
		}

		// TODO: update admins with useful information
		man.dmux.Lock()
		man.data = data
		man.dmux.Unlock()
		payload, err := json.MarshalIndent(data, "", " ")
		if err == nil {
			man.sendTo("admin", wrapper{
				Type: "data",
				Msg:  json.RawMessage(payload),
			})
		} else {
			log.Println(host+": Problem marshaling for admin send:", err)
		}
	}
}

func (man *manager) Data(w http.ResponseWriter, r *http.Request) {
	man.dmux.RLock()
	defer man.dmux.RUnlock()
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	if err := enc.Encode(man.data); err != nil {
		http.Error(w, "Cannot encode JSON", http.StatusInternalServerError)
	}
}
