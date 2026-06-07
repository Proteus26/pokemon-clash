package network

import (
	"fmt"
)

type Hub struct {
	clients map[*Client] bool
	register chan *Client
	unregister chan *Client 
	waiting *Client
}

func Newhub() *Hub {
	return &Hub {
		register: make(chan *Client),
		unregister: make(chan *Client),
		clients: make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <- h.register:
			h.clients[client] = true
			fmt.Println("[hub] New player registered.")
			h.matchmake(client)

		case client := <- h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)

				if h.waiting == client {
					h.waiting = nil
					fmt.Println("[hub] Waiting player left the queue.")
				}

				fmt.Println("[hub] Player disconnected")	
			}
		}
	}
}

func (h *Hub) matchmake(client *Client) {
	if h.waiting == nil {
		h.waiting = client
		client.Send <- []byte(`{"system": "Waiting for an opponent..."}`)
		fmt.Println("[hub] Player added to waitlist")
	} else {
		p1 := h.waiting
		p2 := client
		h.waiting = nil

		p1.Send <- []byte(`{"system": "Opponent found! Battle starting."}`)
		p2.Send <- []byte(`{"system": "Opponent found! Battle starting."}`)

		fmt.Println("[Hub] Match found! Spinning up battle instance...")

		//todo: starting the battle stuff here
	}
}
