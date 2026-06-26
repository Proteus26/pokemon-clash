package engine

import (
	"fmt"
	"pokemon-clash/loader"
)

func InjectPokemon(id, pkmnID string, level int) (*Pokemon, error) {
	blueprint, exists := loader.GetPokemon(pkmnID)
	if !exists {
		return nil, fmt.Errorf("pokemon %s not found in the dex", pkmnID)
	}

	maxHP := ((2 * blueprint.BaseStats.HP) * level / 100) + level + 10
	atk := ((2 * blueprint.BaseStats.Atk) * level / 100) + 5
	def := ((2 * blueprint.BaseStats.Def) * level / 100) + 5
	spa := ((2 * blueprint.BaseStats.SpA) * level / 100) + 5
	spd := ((2 * blueprint.BaseStats.SpD) * level / 100) + 5
	spe := ((2 * blueprint.BaseStats.Spe) * level / 100) + 5

	return &Pokemon{
		ID:    id,
		Mon:   blueprint.Name,
		Level: level,
		Types: blueprint.Types,
		MaxHP: maxHP,
		HP:    maxHP,
		Atk:   atk,
		Def:   def,
		SpA:   spa,
		SpD:   spd,
		Spe:   spe,
	}, nil
}

func BuildTeam(pid string, teamIDs []string) (*Player, error) {
	if len(teamIDs) == 0 || len(teamIDs) > 6 {
		return nil, fmt.Errorf("invalid team size")
	}

	var team []*Pokemon

	for i, mon := range teamIDs {
		instanceID := fmt.Sprintf("%s-%d", pid, i)

		pkmn, err := InjectPokemon(instanceID, mon, 50)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %v", mon, err)
		}
		team = append(team, pkmn)
	}

	return &Player{
		ID:     pid,
		Team:   team,
		Active: team[0],
	}, nil
}
