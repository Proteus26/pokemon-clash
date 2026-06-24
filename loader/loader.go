package loader

import (
	"fmt"
	"os"
	"encoding/json"
	"strings"
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
	Abilities map[string]string `json:"abilities"`
}

type Movedata struct {
	Id int `json:"num"`
	Name string `json:"string"`
	Bp int `json:"basePower"`
	Category string `json:"category"`
	Type string `json:"type"`
	Prio int `json:"priority"`
}

var Pokedex map[string]Pokemondata
var Moves map[string]Movedata

func LoadAll() error {
	dexFile, err := os.ReadFile("data/pokedex.json")
	if err != nil {
		return fmt.Errorf("failed to read pokedex.json: %w", err)
	}

	err = json.Unmarshal(dexFile, &Pokedex)
	if err != nil {
		return fmt.Errorf("failed to parse pokedex.json: %w", err)
	}

	moveFile, err := os.ReadFile("data/moves.json")
	if err != nil {
		return fmt.Errorf("failed to read moves.json: %w", err)
	}

	err = json.Unmarshal(moveFile, &Moves)
	if err != nil {
		return fmt.Errorf("failed to parse moves.json: %w", err)
	}

	fmt.Printf("Successfully loaded %d Pokemon and %d Moves.\n", len(Pokedex), len(Moves))
	return nil
}

func ToId(s string) string {
	s = strings.ToLower(s)
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func GetMove(id string) (Movedata, bool) {
	nId := ToId(id)
	move, exists := Moves[nId]
	return move, exists
}

func GetPokemon(id string) (Pokemondata, bool) {
	nId := ToId(id)
	pokemon, exists := Pokedex[nId]
	return pokemon, exists
}
