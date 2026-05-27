package engine

import (
	"fmt"
	"time"
)

func initBattle(id string, p1, p2 *player) *battle {
	return &battle{
		id: id,
		p1: p1,
		p2: p2,
		p1chan: make(chan action),
		p2chan: make(chan action),
		broadcast: make(chan string),
	}
}

func (b *battle) start() {
	b.broadcast <- fmt.Sprintf("Battle %s initialized, waiting for inputs", b.id)

	turnTimer := time.NewTicker(30*time.Second)
	defer turnTimer.Stop()

	var act1, act2 *action
	
	for {
		select {
		case act := <-b.p1chan:
			act1 = &act
		case act := <-b.p2chan:
			act2 = &act
		case <- turnTimer.C:
			b.broadcast <- "Timer expired"
			//todo: add default action
		}

		if act1 != nil && act2 != nil {
			//todo: make calculation mechanic

			act1 = nil
			act2 = nil
			turnTimer.Reset(30*time.Second)
		}
	} 
}
