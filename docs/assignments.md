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

- Register the exit command. Update your REPL loop to use the "command" the user typed in to look up the callback function in the registry. If the command is found, call the callback (and print any errors that are returned). If there isn't a handler, just print `Unknown command`.

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

