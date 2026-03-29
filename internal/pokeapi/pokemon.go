package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jukkakansanaho/pokedexcli/internal/pokecache"
)

const defaultPokemonURL = "https://pokeapi.co/api/v2/pokemon/"

// Pokemon represents the relevant fields from GET /pokemon/{name}.
type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

// GetPokemon fetches a single Pokemon by name or id.
// baseURL overrides the API base; if empty, defaultPokemonURL is used.
// Responses are cached using the request URL as the key.
func GetPokemon(client *http.Client, cache *pokecache.Cache, baseURL, name string) (*Pokemon, error) {
	if baseURL == "" {
		baseURL = defaultPokemonURL
	}
	url := baseURL + name + "/"
	if client == nil {
		client = http.DefaultClient
	}

	if cache != nil {
		if body, ok := cache.Get(url); ok {
			var out Pokemon
			if err := json.Unmarshal(body, &out); err != nil {
				return nil, err
			}
			return &out, nil
		}
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	if cache != nil {
		cache.Add(url, body)
	}
	var out Pokemon
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
