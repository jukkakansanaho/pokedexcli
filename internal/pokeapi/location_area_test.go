package pokeapi

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jukkakansanaho/pokedexcli/internal/pokecache"
)

func TestListLocationAreas(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/location-area/" {
			t.Errorf("path = %q; want /location-area/", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"count": 2,
			"next": null,
			"previous": null,
			"results": [
				{"name": "alpha-area", "url": "https://example.com/1/"},
				{"name": "beta-area", "url": "https://example.com/2/"}
			]
		}`))
	}))
	defer ts.Close()

	client := ts.Client()
	out, err := ListLocationAreas(client, nil, ts.URL+"/location-area/")
	if err != nil {
		t.Fatal(err)
	}
	if out.Count != 2 {
		t.Errorf("Count = %d; want 2", out.Count)
	}
	if len(out.Results) != 2 || out.Results[0].Name != "alpha-area" || out.Results[1].Name != "beta-area" {
		t.Errorf("Results = %#v; want alpha-area, beta-area", out.Results)
	}
}

func TestListLocationAreas_cacheHit(t *testing.T) {
	var hits atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"next":null,"previous":null,"results":[{"name":"cached","url":"http://x"}]}`))
	}))
	defer ts.Close()

	u := ts.URL + "/location-area/"
	cache := pokecache.NewCache(1 * time.Hour)
	client := ts.Client()

	_, err := ListLocationAreas(client, cache, u)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ListLocationAreas(client, cache, u)
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Errorf("HTTP handler calls = %d; want 1 (second call served from cache)", hits.Load())
	}
}
