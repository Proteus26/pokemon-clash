package main

import (
	"log"
	"net/http"
	"pokemon-clash/loader"
	"pokemon-clash/network"
)

func main() {
	err := loader.LoadAll()
	if err != nil {
		log.Fatalf("[Server] failed to start: %v\n", err)
	}

	hub := network.NewHub()
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
