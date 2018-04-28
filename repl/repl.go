/*
Package repl provides a hook into the smoosh interpreter.


*/
package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"

	"github.com/laher/smoosh/object"
)

// NewRunner initializes a Runner
func NewRunner() *Runner {
	return &Runner{true, true, false}
}

// Runner can run a repl or a program
type Runner struct {
	Parse    bool
	Evaluate bool
	Format   bool
}

func isPipedInput() bool {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return true
	}
	return false
}

// Start starts a line-by-line processor
func (r *Runner) Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	for {
		if !isPipedInput() {
			pwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			prompt := fmt.Sprintf("[%s]/[%s]> ", user.Username, path.Base(pwd))
			fmt.Printf(prompt)
		}
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		err := r.runData(line, out, env)
		if err != nil {
			panic(err)
		}
	}
}

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
