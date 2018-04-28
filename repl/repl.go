/*
Package repl provides a hook into the smoosh interpreter.


*/
package repl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path"

	"github.com/laher/smoosh/evaluator"
	"github.com/laher/smoosh/lexer"
	"github.com/laher/smoosh/object"
	"github.com/laher/smoosh/parser"
	"github.com/laher/smoosh/token"
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
		l := lexer.New(line)
		if r.Parse {
			p := parser.New(l)

			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				printParserErrors(out, p.Errors())
				continue
			}
			if r.Evaluate {
				evaluated := evaluator.Eval(program, env)
				if evaluated != nil {
					io.WriteString(out, evaluated.Inspect())
					io.WriteString(out, "\n")
				}
			} else {
				if r.Format {
					io.WriteString(out, program.String())
					io.WriteString(out, "\n")
					return
				}
				b, err := json.MarshalIndent(program, "", "  ")
				if err != nil {
					panic(err)
				}
				fmt.Fprintf(out, "%s\n", string(b))
			}
		} else {
			for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
				fmt.Fprintf(out, "%#v\n", tok)
			}
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
