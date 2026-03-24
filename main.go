package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanInput(scanner.Text())
		firstWord := ""
		if len(words) > 0 {
			firstWord = words[0]
		}
		fmt.Println("Your command was:", firstWord)
	}
}
