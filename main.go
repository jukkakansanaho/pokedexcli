package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/jukkakansanaho/pokedexcli/internal/pokecache"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	commands := commandRegistry()
	cfg := &config{
		cache: pokecache.NewCache(5 * time.Minute),
	}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanInput(scanner.Text())
		if len(words) == 0 {
			continue
		}
		cmd := words[0]
		handled, err := runRegisteredCommand(commands, cfg, cmd)
		if err != nil {
			fmt.Println(err)
		}
		if !handled {
			fmt.Println("Unknown command")
		}
	}
}
