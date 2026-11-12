package main

import (
	"flag"
	"log"
	"net/http"
	"pokemon-clash/engine"
	"pokemon-clash/loader"
	"pokemon-clash/network"
)

func main() {
	turnTimeout := flag.Duration("turn-timeout", engine.DefaultTurnTimeout, "time each player has to choose an action")
	flag.Parse()

	err := loader.LoadAll()
	if err != nil {
		log.Fatalf("[Server] failed to start: %v\n", err)
	}

	hub := network.NewHub()
	hub.TurnTimeout = *turnTimeout
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		network.ServeSock(w, r, hub)
	})

	log.Println("[Server] Engine running on ws://localhost:8080/ws")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("[Server] ListenAndServe:", err)
	}
}
