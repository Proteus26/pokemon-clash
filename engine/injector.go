package engine

import (
	"fmt"
	"pokemon-clash/loader"
)

func InjectPokemon(id, pkmnId string, level int) (*Pokemon, error){
	blueprint, exists := loader.GetPokemon(pkmnId)
	if !exists {
		return nil, fmt.Errorf("pokemon %s not found in the dex", pkmnId)
	}

	maxHp := ((2*blueprint.Basestats.Hp)*level/100) + level + 10
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
		MaxHp: maxHp,
		Hp: maxHp,
		Atk: atk,
		Def: def,
		Spa: spa,
		Spd: spd,
		Spe: spe,
	}, nil
} 

func BuildTeam(pid string, teamIds []string) (*Player, error) {
	if len(teamIds) == 0 || len(teamIds) > 6 {
		return nil, fmt.Errorf("Invalid team size")
	}

	var team []*Pokemon

	for i, mon := range teamIds {
		instanceId := fmt.Sprintf("%s-%d", pid, i)

		pkmn, err := InjectPokemon(instanceId, mon, 50)
		if err != nil {
			return nil, fmt.Errorf("Failed to load %s: %v", mon, err)
		}
		team = append(team, pkmn)
	}

	return &Player {
		Id: pid,
		Team: team,
		Active: team[0],
	}, nil
}
