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