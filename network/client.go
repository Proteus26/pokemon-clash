package network

import (
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"pokemon-clash/engine"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn       *websocket.Conn
	Send       chan []byte
	hub        *Hub
	Player     *engine.Player
	ActionChan chan engine.Action
}

func ServeSock(w http.ResponseWriter, r *http.Request, hub *Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[Client] Failed to upgrade connection:", err)
		return
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
		hub:  hub,
	}
	client.hub.register <- client

	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Client] error: %v\n", err)
			}
			break
		}

		var action engine.Action
		err = json.Unmarshal(msg, &action)
		if err != nil {
			log.Println("[Client] Invalid JSON from browser:", err)
			continue
		}

		if action.Act == "join" {
			playerID := fmt.Sprintf("player-%s", c.Conn.RemoteAddr().String())

			player, err := engine.BuildTeam(playerID, action.Team)
			if err != nil {
				c.Send <- []byte(fmt.Sprintf(`{"type": "system", "text": "Failed to build team: %v"}`, err))
				continue
			}

			c.Player = player
			c.hub.ready <- c
		} else {
			if c.Player != nil && c.ActionChan != nil {
				log.Printf("[Client] Player %s used %s: %s\n", c.Player.ID, action.Act, action.Value)
				action.PID = c.Player.ID
				c.ActionChan <- action
			} else {
				log.Printf("[Client] Received combat action when there is no team.\n")
			}
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		msg, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("[Client] error writing message: %v\n", err)
			return
		}
	}
}
