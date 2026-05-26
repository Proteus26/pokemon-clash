package engine

import (
	"fmt"
	"time"
)

func initBattle(id string, p1, p2 *player) *battle{
	return &battle{
		id: id,
		p1: p1,
		p2: p2,
		p1chan: make(chan action),
		p2chan: make(chan action),
		broadcast: make(chan string),
	}
}

//todo: start event loop
