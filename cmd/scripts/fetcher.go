//go:build ignore
package main

import (
	"fmt"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var endpoints = map[string]string{
	"data/pokedex.json": "https://play.pokemonshowdown.com/data/pokedex.json",
	"data/moves.json":   "https://play.pokemonshowdown.com/data/moves.json",
}

type TeamMember struct {
	Mon string `json:"mon"`
}

func downloadFile(filepath string, url string) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		log.Printf("[Fetcher] Failed to fetch %s\n", url)
		return
	}
	defer resp.Body.Close()

	out, _ := os.Create(filepath)
	defer out.Close()
	io.Copy(out, resp.Body)
	log.Printf("[Fetcher] Saved %s\n", filepath)
}

func main() {
	os.MkdirAll("data", os.ModePerm)
	for path, url := range endpoints {
		log.Printf("[Fetcher] Fetching %s\n", url)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("[Fetcher] Failed to fetch %s: %v\n", url, err)
			continue
		}
		defer resp.Body.Close()

		out, err := os.Create(path)
		if err != nil {
			log.Printf("[Fetcher] Failed to make path %s: %v\n", path, err)
			continue
		}

		_, err = io.Copy(out, resp.Body)
		out.Close()

		if err != nil {
			log.Printf("[Fetcher] Failed to write file %s: %v\n", path, err)
		} else {
			log.Printf("[Fetcher] Successfully saved %s\n", path)
		}
	}

	os.MkdirAll("data/sprites/front", os.ModePerm)
	os.MkdirAll("data/sprites/back", os.ModePerm)

	teamFiles, err := filepath.Glob("data/team*.json")
	if err != nil || len(teamFiles) == 0{
		log.Println("[Fetcher] No team files found. Skipping sprites")
		return
	}

	uniqueMons := make(map[string]bool)

	for _, file := range teamFiles {
		bytes, err := os.ReadFile(file)
		if err != nil {
			log.Printf("[Fetcher] Could not read %s\n", file)
			continue
		}

		var team []TeamMember
		if err := json.Unmarshal(bytes, &team); err != nil {
			log.Printf("[Fetcher] Could not parse %s\n", file)
			continue
		}

		for _, member := range team {
			if member.Mon != "" {
				uniqueMons[member.Mon] = true
			}
		}
	}

	log.Printf("[Fetcher] Found %d unique Pokemon across team files. Downloading sprites...\n", len(uniqueMons))

	for mon := range uniqueMons {
		frontUrl := fmt.Sprintf("https://play.pokemonshowdown.com/sprites/gen5/%s.png", mon)
		backUrl := fmt.Sprintf("https://play.pokemonshowdown.com/sprites/gen5-back/%s.png", mon)

		downloadFile(fmt.Sprintf("data/sprites/front/%s.png", mon), frontUrl)
		downloadFile(fmt.Sprintf("data/sprites/back/%s.png", mon), backUrl)
	}

	log.Println("[Fetcher] All data and sprites acquired!")
}
