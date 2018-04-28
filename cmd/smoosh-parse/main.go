package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/laher/smoosh/repl"
)

func main() {
	flag.Parse()
	runner := repl.NewRunner()
	runner.Evaluate = false

	if len(flag.Args()) == 0 {
		runner.Start(os.Stdin, os.Stdout)
		return
	}

	// Run a Smoosh script
	if err := runner.RunFile(flag.Arg(0)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
