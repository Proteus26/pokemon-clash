package engine

import (
	"fmt"
	"pokemon-clash/loader"
)

func Injectpokemon(id, pkmnid string, level int) (*Pokemon, error){
	blueprint, exists := loader.GetPokemon(pkmnid)
	if !exists {
		return nil, fmt.Errorf("pokemon %s not found in the dex", pkmnid)
	}

	maxhp := ((2*blueprint.Basestats.Hp)*level/100) + level + 10
	atk := ((2*blueprint.Basestats.Atk)*level/100) + 5 
	def := ((2*blueprint.Basestats.Def)*level/100) + 5
	spa := ((2*blueprint.Basestats.SpA)*level/100) + 5
	spd := ((2*blueprint.Basestats.SpD)*level/100) + 5
	spe := ((2*blueprint.Basestats.Spe)*level/100) + 5

	return &Pokemon{
		Id: id,
		Mon: blueprint.Name,
		Level: level,
		Types: blueprint.Types,
		Maxhp: maxhp,
		Hp: maxhp,
		Atk: atk,
		Def: def,
		Spa: spa,
		Spd: spd,
		Spe: spe,
	}, nil
} 
