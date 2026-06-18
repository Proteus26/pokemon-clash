package network

import (
	"fmt"
	"pokemon-clash/engine"
)

type Hub struct {
	clients map[*Client] bool
	register chan *Client
	unregister chan *Client 
	ready chan *Client
	waiting *Client
}

func Newhub() *Hub {
	return &Hub {
		register: make(chan *Client),
		unregister: make(chan *Client),
		clients: make(map[*Client]bool),
		ready: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <- h.register:
			h.clients[client] = true
			fmt.Println("[hub] New player registered.")

		case client := <-h.ready:
			fmt.Println("[hub] Queueing player")
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

		battleid := fmt.Sprintf("battle-%s-%s", p1.Player.Id, p2.Player.Id)
		battle := engine.InitBattle(battleid, p1.Player, p2.Player)

		p1.Actionchan = battle.P1chan
		p2.Actionchan = battle.P2chan

		go battle.Start()

		go func() {
			for msg := range battle.Broadcast {
				out := []byte(msg)
				p1.Send <- out
				p2.Send <- out
			}
		}()
	}
}
