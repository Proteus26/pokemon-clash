package engine

import (
	"fmt"
	"time"
)

func InitBattle(id string, p1, p2 *Player) *Battle {
	return &Battle{
		Id: id,
		P1: p1,
		P2: p2,
		P1Chan: make(chan Action),
		P2Chan: make(chan Action),
		Broadcast: make(chan string),
	}
}

func (b *Battle) Start() {
	b.Broadcast <- fmt.Sprintf("Battle %s initialized, waiting for inputs", b.Id)

	turnTimer := time.NewTicker(30*time.Second)
	defer turnTimer.Stop()

	var act1, act2 *Action
	
	for {
		select {
		case act := <-b.P1Chan:
			act1 = &act
		case act := <-b.P2Chan:
			act2 = &act
		case <- turnTimer.C:
			b.Broadcast <- "Timer expired"
			//todo: add default action
		}

		if act1 != nil && act2 != nil {
			//todo: make calculation mechanic
			b.moveOrder(*act1, *act2)

			act1 = nil
			act2 = nil
			turnTimer.Reset(30*time.Second)
		}
	} 
}
