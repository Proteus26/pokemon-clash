package main

import (
	"fmt"
	"log"
	"pokemon-clash/engine"
	"pokemon-clash/loader"
	"time"
)

func main () {
	err := loader.Loadall()
	if err != nil {
		log.Fatalf("server failed to start: %v\n", err)
	}

	p1active := &engine.Pokemon{
		Id: "p1a",
		Mon: "Gengar",
		Hp: 100,
		Maxhp: 100,
		Spd: 100,
	}
	p1 := &engine.Player{
		Id: "playerone",
		Active: p1active,
	}
	
	p2active := &engine.Pokemon{
		Id: "p2a",
		Mon: "Missingno",
		Hp: 100,
		Maxhp: 100,
		Spd: 100,
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
		Value: "Shadow Ball",
	}

	time.Sleep(500*time.Millisecond)
	
	battle.P2chan <- engine.Action{
		Pid: "playertwo",
		Act: "move",
		Value: "Tackle",
	}

	time.Sleep(1*time.Second)
}


