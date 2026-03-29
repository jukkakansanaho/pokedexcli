package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/jukkakansanaho/pokedexcli/internal/pokeapi"
	"github.com/jukkakansanaho/pokedexcli/internal/pokecache"
)

// config holds REPL state needed for commands (e.g. PokeAPI pagination URLs).
type config struct {
	Next              *string
	Previous          *string
	client            *http.Client // if nil, http.DefaultClient; tests may set for httptest
	cache             *pokecache.Cache
	pokeAPIBaseURL    string // if empty, the pokeapi package default is used
	pokemonBaseURL    string // if empty, the pokeapi package default is used
	pokedex           map[string]pokeapi.Pokemon
}

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
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
		"explore": {
			name:        "explore",
			description: "Explore a location area by name",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch a Pokemon by name",
			callback:    commandCatch,
		},
	}
}

func runRegisteredCommand(commands map[string]cliCommand, cfg *config, cmd string, args []string) (handled bool, err error) {
	c, ok := commands[cmd]
	if !ok {
		return false, nil
	}
	return true, c.callback(cfg, args)
}

func helpMessage() string {
	return "Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nmap: List the next 20 location areas\nmapb: List the previous 20 location areas\nexplore <area>: Explore a location area by name\ncatch <pokemon>: Catch a Pokemon by name\nexit: Exit the Pokedex\n"
}

func commandHelp(_ *config, _ []string) error {
	fmt.Print(helpMessage())
	return nil
}

func commandExit(_ *config, _ []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(cfg *config, _ []string) error {
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

func commandMapb(cfg *config, _ []string) error {
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

func commandExplore(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: explore <area_name>")
	}
	areaName := args[0]
	fmt.Printf("Exploring %s...\n", areaName)
	client := cfg.client
	if client == nil {
		client = http.DefaultClient
	}
	area, err := pokeapi.GetLocationArea(client, cfg.cache, cfg.pokeAPIBaseURL, areaName)
	if err != nil {
		if errors.Is(err, pokeapi.ErrNotFound) {
			return fmt.Errorf("%q is not a known location area — did you mean 'catch %s'?", areaName, areaName)
		}
		return fmt.Errorf("could not explore %q: %w", areaName, err)
	}
	fmt.Println("Found Pokemon:")
	for _, enc := range area.PokemonEncounters {
		fmt.Printf(" - %s\n", enc.Pokemon.Name)
	}
	return nil
}

// catchSucceeds reports whether a catch attempt succeeds.
// Higher baseExperience makes the Pokemon harder to catch.
// A random number in [0, baseExperience) is generated; if it is less than
// half the base experience plus a fixed bonus of 30, the catch succeeds.
// This gives roughly 80% chance for weak Pokemon (base exp ~39) and
// ~15% for legendaries (base exp ~340+).
func catchSucceeds(baseExperience int) bool {
	if baseExperience <= 0 {
		return true
	}
	return rand.Intn(baseExperience) < 50
}

func commandCatch(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: catch <pokemon_name>")
	}
	name := args[0]
	fmt.Printf("Throwing a Pokeball at %s...\n", name)

	client := cfg.client
	if client == nil {
		client = http.DefaultClient
	}
	pokemon, err := pokeapi.GetPokemon(client, cfg.cache, cfg.pokemonBaseURL, name)
	if err != nil {
		if errors.Is(err, pokeapi.ErrNotFound) {
			return fmt.Errorf("%q is not a known Pokemon — check the name and try again", name)
		}
		return fmt.Errorf("could not catch %q: %w", name, err)
	}

	if !catchSucceeds(pokemon.BaseExperience) {
		fmt.Printf("%s escaped!\n", name)
		return nil
	}

	fmt.Printf("%s was caught!\n", name)
	if cfg.pokedex == nil {
		cfg.pokedex = make(map[string]pokeapi.Pokemon)
	}
	cfg.pokedex[pokemon.Name] = *pokemon
	return nil
}

func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	return strings.Fields(lower)
}
