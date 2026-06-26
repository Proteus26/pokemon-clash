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
		if b.P2.Active.HP > 0 {
			b.execAction(b.P2, b.P1, act2)
		}
	} else {
		b.execAction(b.P2, b.P1, act2)
		if b.P1.Active.HP > 0 {
			b.execAction(b.P1, b.P2, act1)
		}
	}
}

func (b *Battle) execAction(attacker, defender *Player, act Action) {
	if act.Act == "switch" {
		b.EmitBattle("switch", fmt.Sprintf("%s switched to %s!", attacker.ID, act.Value))
	} else if act.Act == "move" {
		damage := calcDamage(attacker.Active, defender.Active, act.Value)

		defender.Active.HP -= damage
		if defender.Active.HP < 0 {
			defender.Active.HP = 0
		}

		b.EmitBattle("move", fmt.Sprintf("%s used %s!", attacker.Active.Mon, act.Value))
		b.EmitBattle("damage", fmt.Sprintf("%s took %d damage!", defender.Active.Mon, damage))

		if defender.Active.HP == 0 {
			b.EmitBattle("faint", fmt.Sprintf("%s fainted!", defender.Active.Mon))
		}
	}
}

func calcDamage(attacker, defender *Pokemon, moveID string) int {
	moveData, exists := loader.GetMove(moveID)
	if !exists {
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
		a = float64(attacker.SpA)
		d = float64(defender.SpD)
	}

	level := float64(attacker.Level)
	power := float64(moveData.BP)

	stab := 1.0
	for _, t := range attacker.Types {
		if t == moveData.Type {
			stab = 1.5
			break
		}
	}

	baseDmg := ((((2.0 * level / 5.0) + 2.0) * power * (a / d)) / 50.0) + 2.0
	mult := GetEff(moveData.Type, defender.Types)
	rng := 0.85 + (rand.Float64() * 0.15)
	mod := stab * mult * rng

	finalDmg := int(baseDmg * mod)

	if mult == 0 {
		return 0
	}
	if finalDmg < 1 {
		finalDmg = 1
	}

	return finalDmg
}
