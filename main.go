package main

import (
	"os"

	"github.com/laher/smoosh/repl"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
