package main

import (
	"fmt"
	"log"
	"time"
	"net/http"
	"pokemon-clash/engine"
	"pokemon-clash/loader"
	"pokemon-clash/network"
)

func main () {
	err := loader.Loadall()
	if err != nil {
		log.Fatalf("server failed to start: %v\n", err)
	}

	http.HandleFunc("/ws", network.Servesock)
	
	fmt.Println("")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	p1active, err := engine.Injectpokemon("p1a", "gengar", 100)
	if err != nil {
		log.Fatalf("failed to inject p1 active %v\n", err)
	}
	p1 := &engine.Player{
		Id: "playerone",
		Active: p1active,
	}
	
	p2active, err := engine.Injectpokemon("p2a", "missingno", 100)
	if err != nil {
		log.Fatalf("failed to inject p2 active %v\n", err)
	}
	p2 := &engine.Player{
		Id: "playertwo",
		Active: p2active,
	}

	battle := engine.InitBattle("testbattle", p1, p2)

	go battle.Start()

	go func ()  {
		for msg := range battle.Broadcast {
			fmt.Printf("[system] %s\n", msg)
		}
	}()
	
	time.Sleep(100*time.Millisecond)

	battle.P1chan <- engine.Action{
		Pid: "playerone",
		Act: "move",
		Value: "shadowball",
	}

	time.Sleep(500*time.Millisecond)
	
	battle.P2chan <- engine.Action{
		Pid: "playertwo",
		Act: "move",
		Value: "tackle",
	}

	time.Sleep(1*time.Second)
}


