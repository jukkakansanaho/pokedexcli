package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/jukkakansanaho/pokedexcli/internal/pokeapi"
	"github.com/jukkakansanaho/pokedexcli/internal/pokecache"
)

// config holds REPL state needed for commands (e.g. PokeAPI pagination URLs).
type config struct {
	Next     *string
	Previous *string
	client   *http.Client // if nil, http.DefaultClient; tests may set for httptest
	cache    *pokecache.Cache
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config) error
}

func commandRegistry() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "List the next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "List the previous 20 location areas",
			callback:    commandMapb,
		},
	}
}

func runRegisteredCommand(commands map[string]cliCommand, cfg *config, cmd string) (handled bool, err error) {
	c, ok := commands[cmd]
	if !ok {
		return false, nil
	}
	return true, c.callback(cfg)
}

func helpMessage() string {
	return "Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nmap: List the next 20 location areas\nmapb: List the previous 20 location areas\nexit: Exit the Pokedex\n"
}

func commandHelp(_ *config) error {
	fmt.Print(helpMessage())
	return nil
}

func commandExit(_ *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(cfg *config) error {
	var pageURL string
	if cfg.Next != nil && *cfg.Next != "" {
		pageURL = *cfg.Next
	}
	client := cfg.client
	if client == nil {
		client = http.DefaultClient
	}
	page, err := pokeapi.ListLocationAreas(client, cfg.cache, pageURL)
	if err != nil {
		return err
	}
	cfg.Next = page.Next
	cfg.Previous = page.Previous
	for _, r := range page.Results {
		fmt.Println(r.Name)
	}
	return nil
}

func commandMapb(cfg *config) error {
	if cfg.Previous == nil || *cfg.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}
	client := cfg.client
	if client == nil {
		client = http.DefaultClient
	}
	page, err := pokeapi.ListLocationAreas(client, cfg.cache, *cfg.Previous)
	if err != nil {
		return err
	}
	cfg.Next = page.Next
	cfg.Previous = page.Previous
	for _, r := range page.Results {
		fmt.Println(r.Name)
	}
	return nil
}

func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	return strings.Fields(lower)
}
