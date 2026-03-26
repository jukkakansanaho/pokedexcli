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
		"help": {wantName: "help", wantDesc: "Displays a help message"},
		"map":  {wantName: "map", wantDesc: "List the next 20 location areas"},
		"mapb": {wantName: "mapb", wantDesc: "List the previous 20 location areas"},
		"exit": {wantName: "exit", wantDesc: "Exit the Pokedex"},
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
	want := "Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nmap: List the next 20 location areas\nmapb: List the previous 20 location areas\nexit: Exit the Pokedex\n"
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
		errCh <- commandHelp(&config{})
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
		if err := commandMap(cfg); err != nil {
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
		if err := commandMap(cfg); err != nil {
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
		if err := commandMap(cfg); err != nil {
			t.Fatalf("commandMap: %v", err)
		}
	})
	captureStdout(t, func() {
		if err := commandMap(cfg); err != nil {
			t.Fatalf("commandMap 2: %v", err)
		}
	})
	outMapb := captureStdout(t, func() {
		if err := commandMapb(cfg); err != nil {
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
		if err := commandMapb(cfg); err != nil {
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
		if err := commandMap(cfg); err != nil {
			t.Fatal(err)
		}
	})
	out := captureStdout(t, func() {
		if err := commandMapb(cfg); err != nil {
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
	err := commandMapb(cfg)
	if err == nil {
		t.Fatal("commandMapb: want error on HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %v; want message mentioning 500", err)
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
	err := commandMap(cfg)
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
		handled, err := runRegisteredCommand(reg, cfg, "map")
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
		handled, err := runRegisteredCommand(reg, cfg, "mapb")
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
			"yes": {callback: func(*config) error { return nil }},
		}
		handled, err := runRegisteredCommand(commands, cfg, "no")
		if handled || err != nil {
			t.Errorf("handled=%v err=%v; want handled=false err=nil", handled, err)
		}
	})
	t.Run("callback succeeds", func(t *testing.T) {
		var called bool
		commands := map[string]cliCommand{
			"ping": {callback: func(*config) error { called = true; return nil }},
		}
		handled, err := runRegisteredCommand(commands, cfg, "ping")
		if !handled || err != nil || !called {
			t.Errorf("handled=%v err=%v called=%v; want handled=true err=nil called=true", handled, err, called)
		}
	})
	t.Run("callback returns error", func(t *testing.T) {
		want := errors.New("boom")
		commands := map[string]cliCommand{
			"bad": {callback: func(*config) error { return want }},
		}
		handled, err := runRegisteredCommand(commands, cfg, "bad")
		if !handled || err != want {
			t.Errorf("handled=%v err=%v; want handled=true err=boom", handled, err)
		}
	})
}
