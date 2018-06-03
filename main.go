package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	"github.com/laher/smoosh/run"
)

func main() {
	runner := run.NewRunner()
	flag.BoolVar(&runner.Evaluate, "eval", true, "evaluate input")
	flag.BoolVar(&runner.Parse, "parse", true, "parse input")
	flag.BoolVar(&runner.Format, "fmt", false, "format inut")
	diagPort := ""
	flag.StringVar(&diagPort, "diag", "", "diagnostics port (e.g. ':6060')")
	flag.Parse()
	if runner.Format {
		runner.Evaluate = false
	}
	if diagPort != "" {
		go func() {
			// for the duration of the program
			log.Println(http.ListenAndServe(diagPort, nil))
		}()
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
