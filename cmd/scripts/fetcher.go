//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	showdown  = "https://play.pokemonshowdown.com"
	spriteDir = "data/sprites"
)

var dataFiles = map[string]string{
	"data/pokedex.json": showdown + "/data/pokedex.json",
	"data/moves.json":   showdown + "/data/moves.json",
}

var spriteSets = map[string]string{
	"front": showdown + "/sprites/gen5",
	"back":  showdown + "/sprites/gen5-back",
}

var client = &http.Client{Timeout: 30 * time.Second}

type TeamMember struct {
	Mon string `json:"mon"`
}

func download(path, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".fetch-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func teamMons() (map[string]bool, error) {
	files, err := filepath.Glob("data/team*.json")
	if err != nil {
		return nil, err
	}

	mons := map[string]bool{}
	for _, file := range files {
		bytes, err := os.ReadFile(file)
		if err != nil {
			log.Printf("[Fetcher] Could not read %s: %v\n", file, err)
			continue
		}

		var team []TeamMember
		if err := json.Unmarshal(bytes, &team); err != nil {
			log.Printf("[Fetcher] Could not parse %s: %v\n", file, err)
			continue
		}

		for _, member := range team {
			if member.Mon != "" {
				mons[member.Mon] = true
			}
		}
	}
	return mons, nil
}

func main() {
	failed := 0
	fetch := func(path, url string) {
		if err := download(path, url); err != nil {
			log.Printf("[Fetcher] Failed to fetch %s: %v\n", url, err)
			failed++
			return
		}
		log.Printf("[Fetcher] Saved %s\n", path)
	}

	for path, url := range dataFiles {
		fetch(path, url)
	}

	mons, err := teamMons()
	switch {
	case err != nil:
		log.Printf("[Fetcher] Could not search for team files: %v\n", err)
		failed++
	case len(mons) == 0:
		log.Println("[Fetcher] No team files found. Skipping sprites.")
	default:
		log.Printf("[Fetcher] Found %d unique Pokemon across team files. Downloading sprites...\n", len(mons))
		for mon := range mons {
			for side, base := range spriteSets {
				fetch(filepath.Join(spriteDir, side, mon+".png"), fmt.Sprintf("%s/%s.png", base, mon))
			}
		}
	}

	if failed > 0 {
		log.Fatalf("[Fetcher] Finished with %d failed download(s).\n", failed)
	}
	log.Println("[Fetcher] All data and sprites acquired!")
}
