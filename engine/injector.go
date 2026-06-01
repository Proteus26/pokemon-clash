package engine

import (
	"fmt"
	"pokemon-clash/loader"
)

func Injectpokemon(id, pkmnid string, level int) (*Pokemon, error){
	blueprint, exists := loader.GetPokemon(pkmnid)
	if !exists {
		return nil, fmt.Errorf("pokemon not found in the dex", pkmnid)
	}

	maxhp := ((2*blueprint.Basestats.Hp)*level/100) + level + 10
	spd := ((2*blueprint.Basestats.Spe)*level/100) + 5

	return &Pokemon{
		Id: id,
		Mon: blueprint.Name,
		Maxhp: maxhp,
		Hp: maxhp,
		Spd: spd,
	}, nil
} 
