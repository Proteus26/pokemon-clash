package engine

import (
	"fmt"
	"math/rand"
	"pokemon-clash/loader"
)

func (b *Battle) moveOrder(act1, act2 Action) {
	p1First := true

	if act1.Act == "switch" && act2.Act == "move" {
		p1First = true
	} else if act2.Act == "switch" && act1.Act == "move" {
		p1First = false
	} else {
		spe1 := b.P1.Active.Spe
		spe2 := b.P2.Active.Spe

		if spe1 > spe2 {
			p1First = true
		} else if spe2 > spe1 {
			p1First = false
		} else {
			p1First = rand.Intn(2) == 0
		}
	} 

	if p1First {
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

func calcDamage(attacker, defender *Pokemon, moveId string) int {
	moveData, exists := loader.GetMove(moveId)
	if !exists {
		fmt.Printf("move '%s' not found\n", moveId)
		return 10
	}

	if moveData.Category == "Status" {
		return 0
	}
	
	var a, d float64
	if moveData.Category == "Physical" {
		a = float64(attacker.Atk)
		d = float64(defender.Def)
	} else if moveData.Category == "Special" {
		a = float64(attacker.Spa)
		d = float64(defender.Spd)
	}

	level := float64(attacker.Level)
	power := float64(moveData.Bp)

	stab := 1.0
	for _, t := range attacker.Types {
		if t == moveData.Type {
			stab = 1.5
			break
		}
	}

	baseDmg := ((((2.0*level/5.0) + 2.0)*power*(a/d))/50.0) + 2.0
	mult := getEff(moveData.Type, defender.Types)
	rng := 0.85 + (rand.Float64() * 0.15)
	mod := stab * mult * rng

	finalDmg := int(baseDmg*mod)

	if mult == 0 {
		return 0
	}
	if finalDmg < 1 {
		finalDmg = 1
	}

	return finalDmg
	//todo: some lsp optimizations to the code by usiing swtich statements and slices ig
}
