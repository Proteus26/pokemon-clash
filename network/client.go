package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pokemon-clash/engine"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const writeWait = 10 * time.Second

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn   *websocket.Conn
	Send   chan []byte
	hub    *Hub
	Player *engine.Player

	closed    chan struct{}
	closeOnce sync.Once

	mu      sync.Mutex
	battle  *engine.Battle
	actions chan engine.Action
}

func (c *Client) deliver(msg []byte) {
	select {
	case c.Send <- msg:
	case <-c.closed:
	}
}

func (c *Client) close() {
	c.closeOnce.Do(func() { close(c.closed) })
}

func (c *Client) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func (c *Client) joinBattle(b *engine.Battle, actions chan engine.Action) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.battle, c.actions = b, actions
}

func (c *Client) currentBattle() (*engine.Battle, chan engine.Action) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.battle, c.actions
}

func ServeSock(w http.ResponseWriter, r *http.Request, hub *Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[Client] Failed to upgrade connection:", err)
		return
	}

	client := &Client{
		Conn:   conn,
		Send:   make(chan []byte, 256),
		hub:    hub,
		closed: make(chan struct{}),
	}
	client.hub.register <- client

	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.close()
		if battle, _ := c.currentBattle(); battle != nil {
			battle.Leave(c.Player)
		}
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
				c.deliver([]byte(fmt.Sprintf(`{"type": "system", "text": "Failed to build team: %v"}`, err)))
				continue
			}

			c.Player = player
			c.hub.ready <- c
		} else {
			battle, actions := c.currentBattle()
			if c.Player != nil && battle != nil {
				log.Printf("[Client] Player %s used %s: %s\n", c.Player.ID, action.Act, action.Value)
				action.PID = c.Player.ID
				select {
				case actions <- action:
				case <-battle.Done:
				}
			} else {
				log.Printf("[Client] Received combat action when there is no team.\n")
			}
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.close()
		c.Conn.Close()
	}()

	for {
		select {
		case msg := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("[Client] error writing message: %v\n", err)
				return
			}
		case <-c.closed:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
}
