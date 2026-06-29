package engine

import (
	"encoding/json"
	"log"
	"time"
)

func InitBattle(id string, p1, p2 *Player) *Battle {
	return &Battle{
		ID:        id,
		P1:        p1,
		P2:        p2,
		P1Chan:    make(chan Action),
		P2Chan:    make(chan Action),
		Broadcast: make(chan []byte, 100),
	}
}

func (b *Battle) EmitSystem(text string) {
	msg := ServerMessage{Type: "system", Text: text}
	bytes, _ := json.Marshal(msg)
	b.Broadcast <- bytes
}

func (b *Battle) EmitBattle(event, text string) {
	msg := ServerMessage{
		Type:     "battle",
		Event:    event,
		Text:     text,
		P1Active: b.P1.Active.Mon,
		P1HP:     b.P1.Active.HP,
		P1Max:    b.P1.Active.MaxHP,
		P2Active: b.P2.Active.Mon,
		P2HP:     b.P2.Active.HP,
		P2Max:    b.P2.Active.MaxHP,
	}
	bytes, _ := json.Marshal(msg)
	b.Broadcast <- bytes
}

func (b *Battle) Start() {
	log.Printf("[Engine] Battle %s initialized.\n", b.ID)
	b.EmitBattle("start","Battle initialized, waiting for inputs...")

	turnTimer := time.NewTicker(30 * time.Second)
	defer turnTimer.Stop()

	var act1, act2 *Action

	for {
		select {
		case act := <-b.P1Chan:
			act1 = &act
		case act := <-b.P2Chan:
			act2 = &act
		case <-turnTimer.C:
			b.EmitSystem("Timer expired! Waiting on players...")
		}

		if act1 != nil && act2 != nil {
			b.moveOrder(*act1, *act2)
			act1 = nil
			act2 = nil
			turnTimer.Reset(30 * time.Second)
		}
	}
}
