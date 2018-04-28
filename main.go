package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/laher/smoosh/run"
)

func main() {
	flag.Parse()

	runner := run.NewRunner()
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
