package network

import (
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"pokemon-clash/engine"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
	hub *Hub
	Player *engine.Player
}

func Servesock(w http.ResponseWriter, r *http.Request, hub *Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection: ", err)
		return
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
		hub: hub,
	}
	client.hub.register <- client
	log.Println("Player connected")

	go client.readpump()
	go client.writepump()
} 

func (c *Client) readpump(){
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err :=  c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v\n", err)
			}
			break
		}

		var action engine.Action
		err = json.Unmarshal(msg, &action)
		if err != nil {
			log.Println("Invalid JSON from browser:", err)
			continue 
		}

		if action.Act == "join" {
			playerID := fmt.Sprintf("player-%s", c.Conn.RemoteAddr().String())
			
			player, err := engine.Buildteam(playerID, action.Team)
			if err != nil {
				c.Send <- []byte(fmt.Sprintf(`{"system": "Failed to build team: %v"}`, err))
				continue
			}

			c.Player = player	
			c.hub.ready <- c 
		} else {
			if c.Player != nil {
				log.Printf("Player %s used %s: %s", c.Player.Id, action.Act, action.Value)
			} else {
				log.Printf("Received combat action when there is not team")
			}

			//todo: make it actually wire up to the engine
		}
	}
}

func (c *Client) writepump() {
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
			log.Printf("error: %v\n", err)
			return
		}
	}
}
