package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jukkakansanaho/pokedexcli/internal/pokecache"
)

const defaultLocationAreaListURL = "https://pokeapi.co/api/v2/location-area/"

// LocationAreaListResponse is the paginated list payload from GET /location-area/.
type LocationAreaListResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

// ListLocationAreas performs a GET for one page of location areas.
// If url is empty, the first page URL is used.
// If cache is non-nil, responses are read from or stored in the cache keyed by the request URL.
func ListLocationAreas(client *http.Client, cache *pokecache.Cache, url string) (*LocationAreaListResponse, error) {
	if url == "" {
		url = defaultLocationAreaListURL
	}
	if client == nil {
		client = http.DefaultClient
	}

	if cache != nil {
		if body, ok := cache.Get(url); ok {
			log.Printf("pokeapi: cache hit for %s", url)
			var out LocationAreaListResponse
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
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("pokeapi: GET %s: %s", url, resp.Status)
	}
	if cache != nil {
		cache.Add(url, body)
	}
	var out LocationAreaListResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// LocationAreaResponse is the detail payload from GET /location-area/{name}/.
type LocationAreaResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

// GetLocationArea fetches detail for a single location area by name or id.
// baseURL overrides the API base; if empty, defaultLocationAreaListURL is used.
// Responses are cached using the request URL as the key.
func GetLocationArea(client *http.Client, cache *pokecache.Cache, baseURL, name string) (*LocationAreaResponse, error) {
	if baseURL == "" {
		baseURL = defaultLocationAreaListURL
	}
	url := baseURL + name + "/"
	if client == nil {
		client = http.DefaultClient
	}

	if cache != nil {
		if body, ok := cache.Get(url); ok {
			log.Printf("pokeapi: cache hit for %s", url)
			var out LocationAreaResponse
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
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("pokeapi: GET %s: %s", url, resp.Status)
	}
	if cache != nil {
		cache.Add(url, body)
	}
	var out LocationAreaResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
