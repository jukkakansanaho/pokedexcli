package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jukkakansanaho/pokedexcli/internal/pokeapi"
	"github.com/jukkakansanaho/pokedexcli/internal/pokecache"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input:    "",
			expected: []string{},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("cleanInput(%q): length %d; want %d", c.input, len(actual), len(c.expected))
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("cleanInput(%q): actual[%d] = %q; want %q", c.input, i, word, expectedWord)
			}
		}
	}
}

func TestCommandRegistry(t *testing.T) {
	reg := commandRegistry()
	cases := map[string]struct {
		wantName string
		wantDesc string
	}{
		"help":    {wantName: "help", wantDesc: "Displays a help message"},
		"map":     {wantName: "map", wantDesc: "List the next 20 location areas"},
		"mapb":    {wantName: "mapb", wantDesc: "List the previous 20 location areas"},
		"explore": {wantName: "explore", wantDesc: "Explore a location area by name"},
		"catch":   {wantName: "catch", wantDesc: "Catch a Pokemon by name"},
		"exit":    {wantName: "exit", wantDesc: "Exit the Pokedex"},
	}
	for key, w := range cases {
		c, ok := reg[key]
		if !ok {
			t.Fatalf("commandRegistry: missing %q", key)
		}
		if c.name != w.wantName {
			t.Errorf("%s: name = %q; want %q", key, c.name, w.wantName)
		}
		if c.description != w.wantDesc {
			t.Errorf("%s: description = %q; want %q", key, c.description, w.wantDesc)
		}
		if c.callback == nil {
			t.Errorf("%s: callback is nil", key)
		}
	}
}

func TestHelpMessage(t *testing.T) {
	want := "Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nmap: List the next 20 location areas\nmapb: List the previous 20 location areas\nexplore <area>: Explore a location area by name\ncatch <pokemon>: Catch a Pokemon by name\ninspect <pokemon>: Inspect a caught Pokemon by name\nexit: Exit the Pokedex\n"
	if got := helpMessage(); got != want {
		t.Errorf("helpMessage() = %q; want %q", got, want)
	}
}

func TestCommandHelp(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	errCh := make(chan error, 1)
	go func() {
		errCh <- commandHelp(&config{}, nil)
		w.Close()
	}()

	if err := <-errCh; err != nil {
		os.Stdout = old
		r.Close()
		t.Fatalf("commandHelp: %v", err)
	}
	os.Stdout = old

	got, readErr := io.ReadAll(r)
	r.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	want := helpMessage()
	if string(got) != want {
		t.Errorf("commandHelp wrote %q; want %q", got, want)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	got, readErr := io.ReadAll(r)
	r.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(got)
}

func TestCommandMap(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/location-area/" {
			t.Errorf("request path = %q; want /location-area/", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.RawQuery {
		case "":
			next := ts.URL + "/location-area/?offset=20"
			fmt.Fprintf(w, `{"count":10,"next":%q,"previous":null,"results":[{"name":"area-one","url":"http://a"},{"name":"area-two","url":"http://b"}]}`, next)
		case "offset=20":
			prev := ts.URL + "/location-area/"
			fmt.Fprintf(w, `{"count":10,"next":null,"previous":%q,"results":[{"name":"next-page-only","url":"http://c"}]}`, prev)
		default:
			t.Errorf("unexpected query %q", r.URL.RawQuery)
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	page1 := ts.URL + "/location-area/"
	cfg := &config{
		Next:   &page1,
		client: ts.Client(),
	}

	out1 := captureStdout(t, func() {
		if err := commandMap(cfg, nil); err != nil {
			t.Fatalf("commandMap: %v", err)
		}
	})
	want1 := "area-one\narea-two\n"
	if out1 != want1 {
		t.Errorf("first map output = %q; want %q", out1, want1)
	}
	wantNext := ts.URL + "/location-area/?offset=20"
	if cfg.Next == nil || *cfg.Next != wantNext {
		t.Errorf("after first map Next = %v; want %q", derefOrNil(cfg.Next), wantNext)
	}
	if cfg.Previous != nil {
		t.Errorf("after first map Previous = %v; want nil", derefOrNil(cfg.Previous))
	}

	out2 := captureStdout(t, func() {
		if err := commandMap(cfg, nil); err != nil {
			t.Fatalf("commandMap second call: %v", err)
		}
	})
	if out2 != "next-page-only\n" {
		t.Errorf("second map output = %q; want \"next-page-only\\n\"", out2)
	}
	if cfg.Next != nil {
		t.Errorf("after second map Next = %v; want nil", derefOrNil(cfg.Next))
	}
	if cfg.Previous == nil || *cfg.Previous != page1 {
		t.Errorf("after second map Previous = %v; want %q", derefOrNil(cfg.Previous), page1)
	}
}

func TestCommandMap_mapThenMapb(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/location-area/" {
			t.Errorf("request path = %q; want /location-area/", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.RawQuery {
		case "":
			next := ts.URL + "/location-area/?offset=20"
			fmt.Fprintf(w, `{"count":10,"next":%q,"previous":null,"results":[{"name":"area-one","url":"http://a"},{"name":"area-two","url":"http://b"}]}`, next)
		case "offset=20":
			prev := ts.URL + "/location-area/"
			fmt.Fprintf(w, `{"count":10,"next":null,"previous":%q,"results":[{"name":"next-page-only","url":"http://c"}]}`, prev)
		default:
			t.Errorf("unexpected query %q", r.URL.RawQuery)
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	page1 := ts.URL + "/location-area/"
	cfg := &config{
		Next:   &page1,
		client: ts.Client(),
	}

	out1 := captureStdout(t, func() {
		if err := commandMap(cfg, nil); err != nil {
			t.Fatalf("commandMap: %v", err)
		}
	})
	captureStdout(t, func() {
		if err := commandMap(cfg, nil); err != nil {
			t.Fatalf("commandMap 2: %v", err)
		}
	})
	outMapb := captureStdout(t, func() {
		if err := commandMapb(cfg, nil); err != nil {
			t.Fatalf("commandMapb: %v", err)
		}
	})
	if outMapb != out1 {
		t.Errorf("mapb output = %q; want same as first map %q", outMapb, out1)
	}
}

func TestCommandMapb_firstPage(t *testing.T) {
	cfg := &config{}
	out := captureStdout(t, func() {
		if err := commandMapb(cfg, nil); err != nil {
			t.Fatalf("commandMapb: %v", err)
		}
	})
	if out != "you're on the first page\n" {
		t.Errorf("output = %q; want \"you're on the first page\\n\"", out)
	}
}

func TestCommandMapb_firstPageAfterMap(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"count":1,"next":null,"previous":null,"results":[{"name":"only","url":"http://x"}]}`)
	}))
	defer ts.Close()
	pageURL := ts.URL + "/location-area/"
	cfg := &config{
		Next:   &pageURL,
		client: ts.Client(),
	}
	captureStdout(t, func() {
		if err := commandMap(cfg, nil); err != nil {
			t.Fatal(err)
		}
	})
	out := captureStdout(t, func() {
		if err := commandMapb(cfg, nil); err != nil {
			t.Fatal(err)
		}
	})
	if out != "you're on the first page\n" {
		t.Errorf("after first API page, mapb output = %q; want first-page message", out)
	}
}

func TestCommandMapb_fetchError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	u := ts.URL + "/location-area/"
	cfg := &config{
		Previous: &u,
		client:   ts.Client(),
	}
	err := commandMapb(cfg, nil)
	if err == nil {
		t.Fatal("commandMapb: want error on HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %v; want message mentioning 500", err)
	}
}

const exploreAreaJSON = `{
	"id": 1,
	"name": "pastoria-city-area",
	"pokemon_encounters": [
		{"pokemon": {"name": "tentacool", "url": "http://a"}},
		{"pokemon": {"name": "magikarp",  "url": "http://b"}},
		{"pokemon": {"name": "gyarados",  "url": "http://c"}}
	]
}`

func TestCommandExplore(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/location-area/pastoria-city-area/" {
			t.Errorf("explore: path = %q; want /location-area/pastoria-city-area/", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, exploreAreaJSON)
	}))
	defer ts.Close()

	cfg := &config{client: ts.Client(), pokeAPIBaseURL: ts.URL + "/location-area/"}
	out := captureStdout(t, func() {
		if err := commandExplore(cfg, []string{"pastoria-city-area"}); err != nil {
			t.Fatalf("commandExplore: %v", err)
		}
	})

	want := "Exploring pastoria-city-area...\nFound Pokemon:\n - tentacool\n - magikarp\n - gyarados\n"
	if out != want {
		t.Errorf("commandExplore output = %q; want %q", out, want)
	}
}

func TestCommandExplore_noArgs(t *testing.T) {
	cfg := &config{}
	err := commandExplore(cfg, nil)
	if err == nil {
		t.Fatal("commandExplore with no args: want error, got nil")
	}
	if !strings.Contains(err.Error(), "usage") {
		t.Errorf("error = %v; want message containing \"usage\"", err)
	}
}

func TestCommandExplore_fetchError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	cfg := &config{client: ts.Client(), pokeAPIBaseURL: ts.URL + "/location-area/"}
	err := commandExplore(cfg, []string{"unknown-area"})
	if err == nil {
		t.Fatal("commandExplore with 404: want error, got nil")
	}
	if !strings.Contains(err.Error(), "not a known location area") {
		t.Errorf("error = %v; want message containing \"not a known location area\"", err)
	}
}

func TestCommandExplore_cacheHit(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, exploreAreaJSON)
	}))
	defer ts.Close()

	cfg := &config{
		client:         ts.Client(),
		pokeAPIBaseURL: ts.URL + "/location-area/",
		cache:          pokecache.NewCache(1 * time.Hour),
	}
	captureStdout(t, func() {
		if err := commandExplore(cfg, []string{"pastoria-city-area"}); err != nil {
			t.Fatalf("first explore: %v", err)
		}
	})
	captureStdout(t, func() {
		if err := commandExplore(cfg, []string{"pastoria-city-area"}); err != nil {
			t.Fatalf("second explore: %v", err)
		}
	})
	if hits != 1 {
		t.Errorf("HTTP handler calls = %d; want 1 (second explore served from cache)", hits)
	}
}

func TestCommandExplore_viaRegistry(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, exploreAreaJSON)
	}))
	defer ts.Close()

	cfg := &config{client: ts.Client(), pokeAPIBaseURL: ts.URL + "/location-area/"}
	reg := commandRegistry()
	out := captureStdout(t, func() {
		handled, err := runRegisteredCommand(reg, cfg, "explore", []string{"pastoria-city-area"})
		if !handled || err != nil {
			t.Fatalf("handled=%v err=%v", handled, err)
		}
	})
	if !strings.Contains(out, "tentacool") {
		t.Errorf("output = %q; want it to contain \"tentacool\"", out)
	}
}

func derefOrNil(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

func TestCommandMap_fetchError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()
	u := ts.URL + "/location-area/"
	cfg := &config{
		Next:   &u,
		client: ts.Client(),
	}
	err := commandMap(cfg, nil)
	if err == nil {
		t.Fatal("commandMap: want error on HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %v; want message mentioning 500", err)
	}
}

func TestRunRegisteredCommand_map(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"count":1,"next":null,"previous":null,"results":[{"name":"from-registry","url":"http://x"}]}`)
	}))
	defer ts.Close()
	pageURL := ts.URL + "/location-area/"
	cfg := &config{
		Next:   &pageURL,
		client: ts.Client(),
	}
	reg := commandRegistry()
	out := captureStdout(t, func() {
		handled, err := runRegisteredCommand(reg, cfg, "map", nil)
		if !handled || err != nil {
			t.Fatalf("handled=%v err=%v", handled, err)
		}
	})
	if out != "from-registry\n" {
		t.Errorf("output = %q; want \"from-registry\\n\"", out)
	}
}

func TestRunRegisteredCommand_mapb(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"count":1,"next":null,"previous":null,"results":[{"name":"back-via-registry","url":"http://x"}]}`)
	}))
	defer ts.Close()
	prevURL := ts.URL + "/location-area/"
	cfg := &config{
		Previous: &prevURL,
		client:   ts.Client(),
	}
	reg := commandRegistry()
	out := captureStdout(t, func() {
		handled, err := runRegisteredCommand(reg, cfg, "mapb", nil)
		if !handled || err != nil {
			t.Fatalf("handled=%v err=%v", handled, err)
		}
	})
	if out != "back-via-registry\n" {
		t.Errorf("output = %q; want \"back-via-registry\\n\"", out)
	}
}

func TestRunRegisteredCommand(t *testing.T) {
	cfg := &config{}
	t.Run("unknown command", func(t *testing.T) {
		commands := map[string]cliCommand{
			"yes": {callback: func(*config, []string) error { return nil }},
		}
		handled, err := runRegisteredCommand(commands, cfg, "no", nil)
		if handled || err != nil {
			t.Errorf("handled=%v err=%v; want handled=false err=nil", handled, err)
		}
	})
	t.Run("callback succeeds", func(t *testing.T) {
		var called bool
		commands := map[string]cliCommand{
			"ping": {callback: func(*config, []string) error { called = true; return nil }},
		}
		handled, err := runRegisteredCommand(commands, cfg, "ping", nil)
		if !handled || err != nil || !called {
			t.Errorf("handled=%v err=%v called=%v; want handled=true err=nil called=true", handled, err, called)
		}
	})
	t.Run("callback returns error", func(t *testing.T) {
		want := errors.New("boom")
		commands := map[string]cliCommand{
			"bad": {callback: func(*config, []string) error { return want }},
		}
		handled, err := runRegisteredCommand(commands, cfg, "bad", nil)
		if !handled || err != want {
			t.Errorf("handled=%v err=%v; want handled=true err=boom", handled, err)
		}
	})
}

// --- catch command tests ---

const pikachuJSON = `{
	"id": 25,
	"name": "pikachu",
	"base_experience": 112,
	"height": 4,
	"weight": 60,
	"stats": [
		{"base_stat": 35, "stat": {"name": "hp"}},
		{"base_stat": 55, "stat": {"name": "attack"}}
	],
	"types": [
		{"type": {"name": "electric"}}
	]
}`

func TestCommandCatch_noArgs(t *testing.T) {
	cfg := &config{}
	err := commandCatch(cfg, nil)
	if err == nil {
		t.Fatal("commandCatch with no args: want error, got nil")
	}
	if !strings.Contains(err.Error(), "usage") {
		t.Errorf("error = %v; want message containing \"usage\"", err)
	}
}

func TestCommandCatch_fetchError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	cfg := &config{client: ts.Client(), pokemonBaseURL: ts.URL + "/pokemon/"}
	err := commandCatch(cfg, []string{"unknownmon"})
	if err == nil {
		t.Fatal("commandCatch with 404: want error, got nil")
	}
	if !strings.Contains(err.Error(), "not a known Pokemon") {
		t.Errorf("error = %v; want message containing \"not a known Pokemon\"", err)
	}
}

func TestCommandCatch_printsThrowing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, pikachuJSON)
	}))
	defer ts.Close()
	cfg := &config{client: ts.Client(), pokemonBaseURL: ts.URL + "/pokemon/"}
	out := captureStdout(t, func() {
		_ = commandCatch(cfg, []string{"pikachu"})
	})
	if !strings.HasPrefix(out, "Throwing a Pokeball at pikachu...") {
		t.Errorf("output = %q; want it to start with \"Throwing a Pokeball at pikachu...\"", out)
	}
}

func TestCommandCatch_caught(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// base_experience 0 → catchSucceeds always returns true
		fmt.Fprint(w, `{"id":1,"name":"caterpie","base_experience":0,"height":3,"weight":29,"stats":[],"types":[]}`)
	}))
	defer ts.Close()
	cfg := &config{client: ts.Client(), pokemonBaseURL: ts.URL + "/pokemon/"}
	out := captureStdout(t, func() {
		if err := commandCatch(cfg, []string{"caterpie"}); err != nil {
			t.Fatalf("commandCatch: %v", err)
		}
	})
	if !strings.Contains(out, "was caught!") {
		t.Errorf("output = %q; want it to contain \"was caught!\"", out)
	}
	if cfg.pokedex == nil || cfg.pokedex["caterpie"].Name != "caterpie" {
		t.Errorf("pokedex does not contain caterpie after catch")
	}
}

func TestCommandCatch_escaped(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// base_experience 1000000 → catchSucceeds always returns false (rand.Intn(1000000) < 50 is astronomically unlikely)
		fmt.Fprint(w, `{"id":999,"name":"bigmon","base_experience":1000000,"height":10,"weight":100,"stats":[],"types":[]}`)
	}))
	defer ts.Close()
	cfg := &config{client: ts.Client(), pokemonBaseURL: ts.URL + "/pokemon/"}
	out := captureStdout(t, func() {
		if err := commandCatch(cfg, []string{"bigmon"}); err != nil {
			t.Fatalf("commandCatch: %v", err)
		}
	})
	if !strings.Contains(out, "escaped!") {
		t.Errorf("output = %q; want it to contain \"escaped!\"", out)
	}
	if cfg.pokedex != nil && cfg.pokedex["bigmon"].Name == "bigmon" {
		t.Errorf("pokedex should not contain bigmon after escape")
	}
}

func TestCommandCatch_cacheHit(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":1,"name":"caterpie","base_experience":0,"height":3,"weight":29,"stats":[],"types":[]}`)
	}))
	defer ts.Close()
	cfg := &config{
		client:         ts.Client(),
		pokemonBaseURL: ts.URL + "/pokemon/",
		cache:          pokecache.NewCache(1 * time.Hour),
	}
	captureStdout(t, func() { _ = commandCatch(cfg, []string{"caterpie"}) })
	captureStdout(t, func() { _ = commandCatch(cfg, []string{"caterpie"}) })
	if hits != 1 {
		t.Errorf("HTTP handler calls = %d; want 1 (second catch served from cache)", hits)
	}
}

func TestCommandCatch_inRegistry(t *testing.T) {
	reg := commandRegistry()
	c, ok := reg["catch"]
	if !ok {
		t.Fatal("commandRegistry: missing \"catch\"")
	}
	if c.name != "catch" {
		t.Errorf("catch name = %q; want \"catch\"", c.name)
	}
	if c.description != "Catch a Pokemon by name" {
		t.Errorf("catch description = %q; want \"Catch a Pokemon by name\"", c.description)
	}
	if c.callback == nil {
		t.Error("catch callback is nil")
	}
}

func TestCommandInspect_notCaught(t *testing.T) {
	cfg := &config{pokedex: map[string]pokeapi.Pokemon{}}
	out := captureStdout(t, func() {
		if err := commandInspect(cfg, []string{"pidgey"}); err != nil {
			t.Fatalf("commandInspect: %v", err)
		}
	})
	if !strings.Contains(out, "you have not caught that pokemon") {
		t.Errorf("output = %q; want it to contain \"you have not caught that pokemon\"", out)
	}
}

func TestCommandInspect_caught(t *testing.T) {
	p := pokeapi.Pokemon{
		Name:   "pidgey",
		Height: 3,
		Weight: 18,
		Stats: []struct {
			BaseStat int `json:"base_stat"`
			Stat     struct {
				Name string `json:"name"`
			} `json:"stat"`
		}{
			{BaseStat: 40, Stat: struct {
				Name string `json:"name"`
			}{Name: "hp"}},
			{BaseStat: 45, Stat: struct {
				Name string `json:"name"`
			}{Name: "attack"}},
		},
		Types: []struct {
			Type struct {
				Name string `json:"name"`
			} `json:"type"`
		}{
			{Type: struct {
				Name string `json:"name"`
			}{Name: "normal"}},
			{Type: struct {
				Name string `json:"name"`
			}{Name: "flying"}},
		},
	}
	cfg := &config{pokedex: map[string]pokeapi.Pokemon{"pidgey": p}}
	out := captureStdout(t, func() {
		if err := commandInspect(cfg, []string{"pidgey"}); err != nil {
			t.Fatalf("commandInspect: %v", err)
		}
	})
	for _, want := range []string{"Name: pidgey", "Height: 3", "Weight: 18", "Stats:", "-hp: 40", "-attack: 45", "Types:", "- normal", "- flying"} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q; want it to contain %q", out, want)
		}
	}
}

func TestCommandInspect_noArgs(t *testing.T) {
	cfg := &config{}
	err := commandInspect(cfg, []string{})
	if err == nil {
		t.Error("commandInspect with no args: want error, got nil")
	}
}

func TestCommandInspect_inRegistry(t *testing.T) {
	reg := commandRegistry()
	c, ok := reg["inspect"]
	if !ok {
		t.Fatal("commandRegistry: missing \"inspect\"")
	}
	if c.name != "inspect" {
		t.Errorf("inspect name = %q; want \"inspect\"", c.name)
	}
	if c.description != "Inspect a caught Pokemon by name" {
		t.Errorf("inspect description = %q; want \"Inspect a caught Pokemon by name\"", c.description)
	}
	if c.callback == nil {
		t.Error("inspect callback is nil")
	}
}

func TestCatchSucceeds_zeroBaseExp(t *testing.T) {
	for i := 0; i < 20; i++ {
		if !catchSucceeds(0) {
			t.Error("catchSucceeds(0) returned false; want always true")
		}
	}
}

func TestCatchSucceeds_veryHighBaseExp(t *testing.T) {
	successes := 0
	trials := 1000
	for i := 0; i < trials; i++ {
		if catchSucceeds(1000000) {
			successes++
		}
	}
	if successes > 1 {
		t.Errorf("catchSucceeds(1000000): %d/%d succeeded; expected near 0", successes, trials)
	}
}
