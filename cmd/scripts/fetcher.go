//go:build ignore
package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

var endpoints = map[string]string{
	"data/pokedex.json": "https://play.pokemonshowdown.com/data/pokedex.json",
	"data/moves.json":   "https://play.pokemonshowdown.com/data/moves.json",
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
}
