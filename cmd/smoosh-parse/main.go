package main

import (
	"os"

	"github.com/laher/smoosh/repl"
)

func main() {
	repl.Parse(os.Stdin, os.Stdout)
}
