package engine

import (
	"fmt"
	"math/rand"
)

func (b *battle) moveOrder(act1, act2 action) {
	p1first := true

	if act1.act == "switch" && act2.act == "move" {
		p1first = true
	} else if act2.act == "switch" && act1.act == "move" {
		p1first = false
	} else {
		spd1 := b.p1.active.spd
		spd2 := b.p2.active.spd

		if spd1 > spd2 {
			p1first = true
		} else if spd2 > spd1 {
			p1first = false
		} else {
			p1first = rand.Intn(2) == 0
		}
	} 

	if p1first {
		b.execAction(b.p1, b.p2, act1)
		if b.p2.active.hp > 0 {
			b.execAction(b.p2, b.p1, act2)
		}
	} else {
		b.execAction(b.p2, b.p1, act1)
		if b.p1.active.hp > 0 {
			b.execAction(b.p1, b.p2, act2)
		}

	}
}

func (b *battle) execAction(attacker, defender *player, act action) {
	if act.act == "switch" {
		b.broadcast <- fmt.Sprintf(`{"event": "switch", "player": "%s", "pokemon": "%s"}`, attacker.id, act.act)
	} else if act.act == "move" {
		damage := calcDamage(attacker.active, defender.active, act.value)

		defender.active.hp -= damage
		if defender.active.hp < 0 {
			defender.active.hp = 0
		}

		b.broadcast <- fmt.Sprintf(`{"event": "move", "player": "%s", "move": "%s"}`, attacker.id, act.value)
		b.broadcast <- fmt.Sprintf(`{"event": "damage", "target": "%s", "amount": %d, "remaining_hp": %d}`, defender.id, damage, defender.active.hp)

		if defender.active.hp == 0 {
			b.broadcast <- fmt.Sprintf(`{"event": "faint", "target": "%s"}`, defender.id)
		}
	} //todo: add other action types maybe and also change to switch then because lsp is saying so
}

func calcDamage(attacker, defender *pokemon, move string) int {

	//todo: implement actual damage calculation after you setup fetching moves and such insteaad of this flat 20 for testing

	return 20
}
