package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
func ListLocationAreas(client *http.Client, url string) (*LocationAreaListResponse, error) {
	if url == "" {
		url = defaultLocationAreaListURL
	}
	if client == nil {
		client = http.DefaultClient
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
	var out LocationAreaListResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
