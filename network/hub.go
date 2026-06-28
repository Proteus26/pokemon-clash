package network

import (
	"fmt"
	"log"
	"pokemon-clash/engine"
)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	ready      chan *Client
	waiting    *Client
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		ready:      make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Println("[Hub] New player connected to websocket.")

		case client := <-h.ready:
			log.Printf("[Hub] Queueing player %s\n", client.Player.ID)
			h.matchmake(client)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)

				if h.waiting == client {
					h.waiting = nil
					log.Println("[Hub] Waiting player left the queue.")
				}
				log.Println("[Hub] Player disconnected.")
			}
		}
	}
}

func (h *Hub) matchmake(client *Client) {
	if h.waiting == nil {
		h.waiting = client
		client.Send <- []byte(`{"type": "system", "text": "Waiting for an opponent..."}`)
		log.Println("[Hub] Player added to waitlist.")
	} else {
		p1 := h.waiting
		p2 := client
		h.waiting = nil

		p1.Send <- []byte(`{"type": "system", "text": "Opponent found! Battle starting.", "role":"p1"}`)
		p2.Send <- []byte(`{"type": "system", "text": "Opponent found! Battle starting.", "role":"p2"}`)

		log.Println("[Hub] Match found! Spinning up battle instance...")

		battleID := fmt.Sprintf("battle-%s-%s", p1.Player.ID, p2.Player.ID)
		battle := engine.InitBattle(battleID, p1.Player, p2.Player)

		p1.ActionChan = battle.P1Chan
		p2.ActionChan = battle.P2Chan

		go battle.Start()

		go func() {
			for msg := range battle.Broadcast {
				p1.Send <- msg
				p2.Send <- msg
			}
		}()
	}
}
