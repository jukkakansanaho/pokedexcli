package main

import (
	"errors"
	"io"
	"os"
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
	want := "Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nexit: Exit the Pokedex\n"
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
		errCh <- commandHelp()
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

func TestRunRegisteredCommand(t *testing.T) {
	t.Run("unknown command", func(t *testing.T) {
		commands := map[string]cliCommand{
			"yes": {callback: func() error { return nil }},
		}
		handled, err := runRegisteredCommand(commands, "no")
		if handled || err != nil {
			t.Errorf("handled=%v err=%v; want handled=false err=nil", handled, err)
		}
	})
	t.Run("callback succeeds", func(t *testing.T) {
		var called bool
		commands := map[string]cliCommand{
			"ping": {callback: func() error { called = true; return nil }},
		}
		handled, err := runRegisteredCommand(commands, "ping")
		if !handled || err != nil || !called {
			t.Errorf("handled=%v err=%v called=%v; want handled=true err=nil called=true", handled, err, called)
		}
	})
	t.Run("callback returns error", func(t *testing.T) {
		want := errors.New("boom")
		commands := map[string]cliCommand{
			"bad": {callback: func() error { return want }},
		}
		handled, err := runRegisteredCommand(commands, "bad")
		if !handled || err != want {
			t.Errorf("handled=%v err=%v; want handled=true err=boom", handled, err)
		}
	})
}
