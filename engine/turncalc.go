package engine

import (
	"fmt"
	"math/rand"
	"pokemon-clash/loader"
)

func (b *Battle) moveOrder(act1, act2 Action) {
	p1first := true

	if act1.Act == "switch" && act2.Act == "move" {
		p1first = true
	} else if act2.Act == "switch" && act1.Act == "move" {
		p1first = false
	} else {
		spd1 := b.P1.Active.Spd
		spd2 := b.P2.Active.Spd

		if spd1 > spd2 {
			p1first = true
		} else if spd2 > spd1 {
			p1first = false
		} else {
			p1first = rand.Intn(2) == 0
		}
	} 

	if p1first {
		b.execAction(b.P1, b.P2, act1)
		if b.P2.Active.Hp > 0 {
			b.execAction(b.P2, b.P1, act2)
		}
	} else {
		b.execAction(b.P2, b.P1, act2)
		if b.P1.Active.Hp > 0 {
			b.execAction(b.P1, b.P2, act1)
		}

	}
}

func (b *Battle) execAction(attacker, defender *Player, act Action) {
	if act.Act == "switch" {
		b.Broadcast <- fmt.Sprintf(`{"event": "switch", "player": "%s", "pokemon": "%s"}`, attacker.Id, act.Act)
	} else if act.Act == "move" {
		damage := calcDamage(attacker.Active, defender.Active, act.Value)

		defender.Active.Hp -= damage
		if defender.Active.Hp < 0 {
			defender.Active.Hp = 0
		}

		b.Broadcast <- fmt.Sprintf(`{"event": "move", "player": "%s", "pokemon": "%s", "move": "%s"}`, attacker.Id, attacker.Active.Mon, act.Value)
		b.Broadcast <- fmt.Sprintf(`{"event": "damage", "target": "%s", "amount": %d, "remaining_hp": %d}`, defender.Id, damage, defender.Active.Hp)

		if defender.Active.Hp == 0 {
			b.Broadcast <- fmt.Sprintf(`{"event": "faint", "target": "%s"}`, defender.Id)
		}
	} //todo: add other Action types maybe and also change to switch then because lsp is saying so
}

func calcDamage(attacker, defender *Pokemon, moveid string) int {
	movedata, exists := loader.GetMove(moveid)
	if !exists {
		fmt.Printf("move '%s' not found\n", moveid)
		return 10
	}

	//todo: better damage calc with actual stat considerations
	damage := movedata.Bp/2
	return damage
}
