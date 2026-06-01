package loader

import (
	"fmt"
	"os"
	"encoding/json"
)

type Basestats struct {
	Hp int `json:"hp"`
	Atk int `json:"atk"`
	Def int `json:"def"`
	SpA int `json:"spa"`
	SpD int `json:"spd"`
	Spe int `json:"spe"`
} 

type Pokemondata struct {
	Id int `json:"num"`
	Name string `json:"name"`
	Types []string `json:"types"`
	Basestats Basestats `json:"baseStats"`
}

type Movedata struct {
	Id int `json:"num"`
	Name string `json:"string"`
	Bp int `json:"basePower"`
	Category string `json:"category"`
	Type string `json:"type"`
}

var Pokedex map[string]Pokemondata
var Moves map[string]Movedata

func Loadall() error {
	dexfile, err := os.ReadFile("data/pokedex.json")
	if err != nil {
		return fmt.Errorf("failed to read pokedex.json: %w", err)
	}

	err = json.Unmarshal(dexfile, &Pokedex)
	if err != nil {
		return fmt.Errorf("failed to parse pokedex.json: %w", err)
	}

	movefile, err := os.ReadFile("data/moves.json")
	if err != nil {
		return fmt.Errorf("failed to read moves.json: %w", err)
	}

	err = json.Unmarshal(movefile, &Moves)
	if err != nil {
		return fmt.Errorf("failed to parse moves.json: %w", err)
	}

	fmt.Printf("Successfully loaded %d Pokemon and %d Moves.\n", len(Pokedex), len(Moves))
	return nil
}

func GetMove(id string) (Movedata, bool) {
	move, exists := Moves[id]
	return move, exists
}

func GetPokemon(id string) (Pokemondata, bool) {
	pokemon, exists := Pokedex[id]
	return pokemon, exists
}
