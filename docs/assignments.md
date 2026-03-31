# Pokedexcli guidelines
Thus file contains assignments (step-by-step development guidelines) for Pokedexcli program.

# Assignments

## Assignment 1

### Task 1.2

- Create a main.go file. It should be part of package main in the root of your project and have a main() function that just prints the text "Hello, World!".

### Task 1.2

- Create a Go module in the root of your project. Here's the command: `go mod github.com/jukkakansanaho/pokedexcli`

### Task 1.3

- Build your program: `go build`

### Task 1.4

- Run the program: `./pokedexcli`

## Assignment 2

### Task 2.1

- Create a new `cleanInput(text string) []string` function. For now it should just return an empty slice of strings.
- The purpose of this function will be to split the user's input into "words" based on whitespace. It should also lowercase the input and trim any leading or trailing whitespace. For example:
  - `hello world` → `["hello", "world"]`
  - `Charmander Bulbasaur PIKACHU` → `["charmander", "bulbasaur", "pikachu"]`

### Task 2.2

- Create a new file for some unit tests. I called mine `repl_test.go` since I put `cleanInput` in a new file, `repl.go` (but you can organize your project your way, the only requirement is that the test file ends in `_test.go`). Create a test suite for the `cleanInput` function. Here is the basic structure of the test file:
- All tests go inside `TestXXX` functions that take a `*testing.T` argument:

  ```go
  func TestCleanInput(t *testing.T) {
      // ...
  }
  ```

- Remember to import the `testing` package if it isn't imported already.
- I like to start by creating a slice of test case structs, in this case:

  ```go
  cases := []struct {
      input    string
      expected []string
  }{
      {
          input:    "  hello  world  ",
          expected: []string{"hello", "world"},
      },
      // add more cases here
  }
  ```

- Then I loop over the cases and run the tests:

  ```go
  for _, c := range cases {
      actual := cleanInput(c.input)
      // Check the length of the actual slice against the expected slice
      // if they don't match, use t.Errorf to print an error message
      // and fail the test
      for i := range actual {
          word := actual[i]
          expectedWord := c.expected[i]
          // Check each word in the slice
          // if they don't match, use t.Errorf to print an error message
          // and fail the test
      }
  }
  ```

### Task 2.3

- Once you have at least a few tests, run the tests using `go test ./...` from the root of the repo. We expect them to fail.

### Task 2.4

- Implement the `cleanInput` function to make the tests pass.

### Task 2.5

- Add one more test in `repl_test.go` to test empty input.

## Assignment 3

### Task 3.1

- Remove your "Hello, World!" logic.

### Task 3.2

- Create support for a simple REPL:
  - In main.go create a bufio.Scanner that reads from os.Stdin, for example: `scanner := bufio.NewScanner(os.Stdin)`. When you later call scanner.Scan it will block and wait for input until the user presses enter.
  - Start an infinite for loop. This loop will execute once for every command the user types in (we don't want to exit the program after just one command)
  - Use fmt.Print to print the prompt `Pokedex >` without a newline character
  - Use the scanner's .Scan and .Text methods to get the user's input as a string
  - Clean the user's input string
  - Capture the first "word" of the input and use it to print: `Your command was: <first word>`

### Task 3.3

- Test your program. Here's an example session:

  ```
  wagslane@MacBook-Pro-2 pokedexcli % go run .
  Pokedex > well hello there
  Your command was: well
  Pokedex > Hello there
  Your command was: hello
  Pokedex > POKEMON was underrated
  Your command was: pokemon

  You can terminate the program by pressing ctrl+c
  ```

### Task 3.4

- Run the CLI again and tee the output (copies the stdout) to a new file called repl.log (and .gitignore the log).
  ```bash
  go run . | tee repl.log
  ```
- Use this as the first input: `CHARMANDER is better than bulbasaur.`
- Use this as the second input: `Pikachu is kinda mean to ash.`
- Terminate the program by pressing ctrl+c.

## Assignment 4

### Task 4.1

- Remove your logic that prints the first word (the command) back to the user
- Add a callback for the exit command. Commands in our REPL are just callback functions with no arguments, but that return an error. For example: `func commandExit() error`
  This function should print `Closing the Pokedex... Goodbye!` then immediately exit the program e.g. `os.Exit(0)`.

### Task 4.2

- Create a "registry" of commands. This will give us a nice abstraction for managing the many commands we'll be adding. Create a struct type that describes a command:

  ```go
  type cliCommand struct {
      name        string
      description string
      callback    func() error
  }
  ```

- Then create a map of supported commands:

  ```go
  map[string]cliCommand{
      "exit": {
          name:        "exit",
          description: "Exit the Pokedex",
          callback:    commandExit,
      },
  }
  ```

- Register the `exit` command. Update your REPL loop to use the "command" the user typed in to look up the callback function in the registry. If the command is found, call the callback (and print any errors that are returned). If there isn't a handler, just print `Unknown command`.

### Task 4.3

- Add unit tests for the new functionality
- Run tests

### Task 4.4

- Add a `help` command, its callback, and register it. It should print:

  ```
  Welcome to the Pokedex!
  Usage:

  help: Displays a help message
  exit: Exit the Pokedex
  ```

- Add unit tests for the new functionality
- Run tests

### Task 4.5

- Test your code again manually


## Assignment 5

### Task 5.1

- Add the `map` command (https://pokeapi.co/). It displays the names of 20 location areas in the Pokemon world. Each subsequent call to map should display the next 20 locations, and so on. This will be how we explore the Pokemon world. Example usage:

  ```
  Pokedex > map
  canalave-city-area
  eterna-city-area
  pastoria-city-area
  sunyshore-city-area
  sinnoh-pokemon-league-area
  oreburgh-mine-1f
  oreburgh-mine-b1f
  valley-windworks-area
  eterna-forest-area
  fuego-ironworks-area
  mt-coronet-1f-route-207
  mt-coronet-2f
  mt-coronet-3f
  mt-coronet-exterior-snowfall
  mt-coronet-exterior-blizzard
  mt-coronet-4f
  mt-coronet-4f-small-room
  mt-coronet-5f
  mt-coronet-6f
  mt-coronet-1f-from-exterior
  ```

  Here are some pointers for implementing this command:

  - You'll need to use the PokeAPI location-area endpoint (https://pokeapi.co/docs/v2#location-areas) to get the location areas. Note that this is a different endpoint than the "location" endpoint. Calling the endpoint without an id will return a batch of location areas.
  - Update all commands (e.g. help, exit, map) to now accept a pointer to a "config" struct as a parameter. This struct will contain the Next and Previous URLs that you'll need to paginate through location areas.
  - Here's an example of how to make a GET request in Go (https://pkg.go.dev/net/http#example-Get).
  - Here's how to unmarshal a slice of bytes into a Go struct (https://www.boot.dev/blog/golang/json-golang/#example-unmarshal-json-to-struct-decode).
    You can make GET requests in your browser or by using curl! It's convenient for testing and debugging.

  TIPS:

  - JSON lint (https://jsonlint.com/) is a useful tool for debugging JSON, it makes it easier to read.
  - JSON to Go (https://mholt.github.io/json-to-go/) a useful tool for converting JSON to Go structs. You can use it to generate the structs you'll need to parse the PokeAPI response. Keep in mind it sometimes can't know the exact type of a field that you want, because there are multiple valid options. For nullable strings, use *string.
  - I recommend creating an internal package (https://dave.cheney.net/2019/10/06/use-internal-packages-to-reduce-your-public-api-surface) that manages your PokeAPI interactions. It's not required, but it's a good organizational and architectural pattern.

### Task 5.2

- Add unit tests to cover new map command functionality

### Task 5.3

- Add the mapb (map back) command. It's similar to the map command, however, instead of displaying the next 20 locations, it displays the previous 20 locations. It's a way to go back.
- If you're on the first "page" of results, this command should just print "you're on the first page". Example usage:

  ```
  Pokedex > map
  canalave-city-area
  eterna-city-area
  pastoria-city-area
  sunyshore-city-area
  sinnoh-pokemon-league-area
  oreburgh-mine-1f
  oreburgh-mine-b1f
  valley-windworks-area
  eterna-forest-area
  fuego-ironworks-area
  mt-coronet-1f-route-207
  mt-coronet-2f
  mt-coronet-3f
  mt-coronet-exterior-snowfall
  mt-coronet-exterior-blizzard
  mt-coronet-4f
  mt-coronet-4f-small-room
  mt-coronet-5f
  mt-coronet-6f
  mt-coronet-1f-from-exterior
  Pokedex > map
  mt-coronet-1f-route-216
  mt-coronet-1f-route-211
  mt-coronet-b1f
  great-marsh-area-1
  great-marsh-area-2
  great-marsh-area-3
  great-marsh-area-4
  great-marsh-area-5
  great-marsh-area-6
  solaceon-ruins-2f
  solaceon-ruins-1f
  solaceon-ruins-b1f-a
  solaceon-ruins-b1f-b
  solaceon-ruins-b1f-c
  solaceon-ruins-b2f-a
  solaceon-ruins-b2f-b
  solaceon-ruins-b2f-c
  solaceon-ruins-b3f-a
  solaceon-ruins-b3f-b
  solaceon-ruins-b3f-c
  Pokedex > mapb
  canalave-city-area
  eterna-city-area
  pastoria-city-area
  sunyshore-city-area
  sinnoh-pokemon-league-area
  oreburgh-mine-1f
  oreburgh-mine-b1f
  valley-windworks-area
  eterna-forest-area
  fuego-ironworks-area
  mt-coronet-1f-route-207
  mt-coronet-2f
  mt-coronet-3f
  mt-coronet-exterior-snowfall
  mt-coronet-exterior-blizzard
  mt-coronet-4f
  mt-coronet-4f-small-room
  mt-coronet-5f
  mt-coronet-6f
  mt-coronet-1f-from-exterior
  ```

## Assignment 6

### Task 6.1

- Create a new internal package called pokecache in your internal directory (if you haven't already created an internal directory in your project, do so now). This package will be responsible for all of our caching logic.
- Use a Cache struct to hold a map[string]cacheEntry and a mutex to protect the map across goroutines. A cacheEntry should be a struct with two fields:

  - createdAt - A time.Time that represents when the entry was created.
  - val - A []byte that represents the raw data we're caching.

  You'll probably want to expose a NewCache() function that creates a new cache with a configurable interval (time.Duration).

- Create a cache.Add() method that adds a new entry to the cache. It should take a key (a string) and a val (a []byte).
- Create a cache.Get() method that gets an entry from the cache. It should take a key (a string) and return a []byte and a bool. The bool should be true if the entry was found and false if it wasn't.
- Create a cache.reapLoop() method that is called when the cache is created (by the NewCache function). Each time an interval (the time.Duration passed to NewCache) passes it should remove any entries that are older than the interval. This makes sure that the cache doesn't grow too large over time. For example, if the interval is 5 seconds, and an entry was added 7 seconds ago, that entry should be removed.

  TIP: Clearing the Cache: You can use a time.Ticker inside a goroutine started by NewCache. In a loop like for range ticker.C { ... }, check the entries and remove any whose createdAt is older than the cache's interval.

  Maps are not thread-safe in Go. You should use a sync.Mutex to lock access to the map when you're adding, getting entries or reaping entries. It's unlikely that you'll have issues because reaping only happens every ~5 seconds, but it's still possible, so you should make your cache package safe for concurrent use.

- Update the code that makes requests to the PokeAPI to use the cache. If you already have the data for a given URL (which is our cache key) in the cache, you should use that instead of making a new request. Whenever you do make a request, you should add the response to the cache.

- Write tests for your cache package!

- Test your application manually to make sure that the cache works as expected. When you use the map command to get data for the first time there should be a noticeable waiting time. However, when you use mapb it should be instantaneous because the data for that page is already in the cache. Feel free to add some logging that informs you in the command line when the cache is being used.

## Assignment 7

### Task 7.1

- Add an `explore` command. It takes the name of a location area as an argument.
- Write tests for "explore" command.

Tips:
- Use the same PokeAPI location-area endpoint (https://pokeapi.co/docs/v2#location-areas), but this time you'll need to pass the name of the location area being explored. By adding a name or id, the API will return a lot more information about the location area.
- Feel free to use tools like JSON lint and JSON to Go to help you parse the response.
- Parse the Pokemon's names from the response and display them to the user.
- Make sure to use the caching layer again! Re-exploring an area should be blazingly fast.
- You'll need to alter the function signature of all your commands to allow them to allow parameters. E.g. explore <area_name>

Example usage:

```
Pokedex > explore pastoria-city-area
Exploring pastoria-city-area...
Found Pokemon:
 - tentacool
 - tentacruel
 - magikarp
 - gyarados
 - remoraid
 - octillery
 - wingull
 - pelipper
 - shellos
 - gastrodon
Pokedex >
```

## Assignment 8

### Task 8.1

- Add a `catch` command. It takes the name of a Pokemon as an argument. Example usage:

Pokedex > catch pikachu
Throwing a Pokeball at pikachu...
pikachu escaped!
Pokedex > catch pikachu
Throwing a Pokeball at pikachu...
pikachu was caught!

- Be sure to print the Throwing a Pokeball at <pokemon>... message before determining if the Pokemon was caught or not.
- Use the Pokemon endpoint (https://pokeapi.co/docs/v2#pokemon ) to get information about a Pokemon by name.
- Give the user a chance to catch the Pokemon using the math/rand package (https://pkg.go.dev/math/rand#Rand.Intn ).
- You can use the pokemon's "base experience" to determine the chance of catching it. The higher the base experience, the harder it should be to catch.
- Once the Pokemon is caught, add it to the user's Pokedex. I used a map[string]Pokemon to keep track of caught Pokemon.
- Test the `catch` command manually - make sure you can actually catch a Pokemon within a reasonable number of tries.

- Write tests for `catch` command and run them.

## Assignment 9

### Task 9.1

- Add an `inspect` command. It takes the name of a Pokemon and prints the name, height, weight, stats and type(s) of the Pokemon. Example usage:

Pokedex > inspect pidgey
you have not caught that pokemon
Pokedex > catch pidgey
Throwing a Pokeball at pidgey...
pidgey was caught!
Pokedex > inspect pidgey
Name: pidgey
Height: 3
Weight: 18
Stats:
  -hp: 40
  -attack: 45
  -defense: 40
  -special-attack: 35
  -special-defense: 35
  -speed: 56
Types:
  - normal
  - flying

- You should not need to make an API call to get this information, since you should have already stored it when the user caught the Pokemon.

- If the user has not caught the Pokemon, just print a message saying so.
