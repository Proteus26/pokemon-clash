package loader

import (
	"fmt"
	"log"
	"os"
	"encoding/json"
	"strings"
)

type BaseStats struct {
	HP  int `json:"hp"`
	Atk int `json:"atk"`
	Def int `json:"def"`
	SpA int `json:"spa"`
	SpD int `json:"spd"`
	Spe int `json:"spe"`
}

type PokemonData struct {
	Num       int               `json:"num"`
	Name      string            `json:"name"`
	Types     []string          `json:"types"`
	BaseStats BaseStats         `json:"baseStats"`
	Abilities map[string]string `json:"abilities"`
}

type MoveData struct {
	Num      int    `json:"num"`
	Name     string `json:"name"`
	BP       int    `json:"basePower"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Prio     int    `json:"priority"`
}

var Pokedex map[string]PokemonData
var Moves map[string]MoveData

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

	log.Printf("[Loader] Successfully loaded %d Pokemon and %d Moves.\n", len(Pokedex), len(Moves))
	return nil
}

func ToID(s string) string {
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

func GetMove(id string) (MoveData, bool) {
	nID := ToID(id)
	move, exists := Moves[nID]
	return move, exists
}

func GetPokemon(id string) (PokemonData, bool) {
	nID := ToID(id)
	pokemon, exists := Pokedex[nID]
	return pokemon, exists
}
