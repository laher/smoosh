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

// Start the repl
func Start(in io.Reader, out io.Writer) {
	start(in, out, true, true)
}

// Tokenize only
func Tokenize(in io.Reader, out io.Writer) {
	start(in, out, false, false)
}

// Tokenize+Parse only
func Parse(in io.Reader, out io.Writer) {
	start(in, out, true, false)
}

func isPipedInput() bool {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return true
	}
	return false
}

func start(in io.Reader, out io.Writer, parse, evaluate bool) {
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
		if parse {
			p := parser.New(l)

			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				printParserErrors(out, p.Errors())
				continue
			}
			if evaluate {
				evaluated := evaluator.Eval(program, env)
				if evaluated != nil {
					io.WriteString(out, evaluated.Inspect())
					io.WriteString(out, "\n")
				}
			} else {
				b, err := json.MarshalIndent(program, "", "  ")
				if err != nil {
					panic(err)
				}
				fmt.Fprintf(out, "%s\n", string(b))
			}
		} else {
			for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
				fmt.Fprintf(out, "%+v\n", tok)
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
