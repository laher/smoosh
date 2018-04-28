package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/laher/smoosh/run"
)

func main() {
	runner := run.NewRunner()
	flag.BoolVar(&runner.Evaluate, "eval", true, "evaluate input")
	flag.BoolVar(&runner.Parse, "parse", true, "parse input")
	flag.BoolVar(&runner.Format, "fmt", false, "format inut")
	flag.Parse()
	if runner.Format {
		runner.Evaluate = false
	}

	if len(flag.Args()) == 0 {
		runner.Start(os.Stdin, os.Stdout)
		return
	}
	// Run a Smoosh script
	if err := runner.RunFile(flag.Arg(0), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
