package main

import (
	"os"

	"github.com/laher/smoosh/repl"
)

func main() {
	repl.Tokenize(os.Stdin, os.Stdout)
}
