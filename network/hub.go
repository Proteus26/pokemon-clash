package network

import (
	"fmt"
	"log"
	"pokemon-clash/engine"
	"time"
)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	ready      chan *Client
	waiting    *Client

	TurnTimeout time.Duration
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		ready:      make(chan *Client),

		TurnTimeout: engine.DefaultTurnTimeout,
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
				client.close()

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
	if h.waiting != nil && h.waiting.isClosed() {
		h.waiting = nil
	}
	if h.waiting == nil {
		h.waiting = client
		client.deliver([]byte(`{"type": "system", "text": "Waiting for an opponent..."}`))
		log.Println("[Hub] Player added to waitlist.")
	} else {
		p1 := h.waiting
		p2 := client
		h.waiting = nil

		p1.deliver([]byte(`{"type": "system", "text": "Opponent found! Battle starting.", "role":"p1"}`))
		p2.deliver([]byte(`{"type": "system", "text": "Opponent found! Battle starting.", "role":"p2"}`))

		log.Println("[Hub] Match found! Spinning up battle instance...")

		battleID := fmt.Sprintf("battle-%s-%s", p1.Player.ID, p2.Player.ID)
		battle := engine.InitBattle(battleID, p1.Player, p2.Player)
		battle.TurnTimeout = h.TurnTimeout

		p1.joinBattle(battle, battle.P1Chan)
		p2.joinBattle(battle, battle.P2Chan)

		go battle.Start()

		for _, c := range []*Client{p1, p2} {
			if c.isClosed() {
				go battle.Leave(c.Player)
			}
		}

		go func() {
			for env := range battle.Out {
				if env.To != "p2" {
					p1.deliver(env.Data)
				}
				if env.To != "p1" {
					p2.deliver(env.Data)
				}
			}
		}()
	}
}
