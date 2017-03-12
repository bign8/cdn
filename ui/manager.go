package main

import (
	"log"

	"golang.org/x/net/websocket"
)

type manager struct {
}

func newManager() (*manager, error) {
	return &manager{}, nil
}

type message struct {
	Message string `json:"msg"`
}

func (man *manager) Handle(ws *websocket.Conn) {
	for {
		var m message
		if err := websocket.JSON.Receive(ws, &m); err != nil {
			log.Println("Receive", err)
			break
		}
		log.Println("Receive Message:", m.Message)
		m = message{"Thanks!"}
		if err := websocket.JSON.Send(ws, m); err != nil {
			log.Println("Send", err)
			break
		}
	}
}
