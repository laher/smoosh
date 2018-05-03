package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

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

// RunFile runs a file as a single program
func (r *Runner) RunFile(filename string, out io.Writer) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return r.Run(f, out)
}

// Run runs an io.Reader as a single program
func (r *Runner) Run(rdr io.Reader, out io.Writer) error {

	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return fmt.Errorf("could not read: %v", err)
	}
	env := object.NewEnvironment()
	return r.runData(string(data), out, env)
}

func (r *Runner) runData(data string, out io.Writer, env *object.Environment) error {
	l := lexer.New(data)
	if r.Parse {
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			return errors.New(p.Errors()[0])
		}
		if r.Evaluate {
			result := evaluator.Eval(program, env)
			if _, ok := result.(*object.Null); ok {
				return nil
			}
			if result == nil {
				return nil
			}
			if pipes, ok := result.(*object.Pipes); ok {
				cmdOut, err := ioutil.ReadAll(pipes.Out)
				if err != nil {
					return err
				}
				err = pipes.Wait()
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(out, "%s", cmdOut)
				return err
			}
			_, err := io.WriteString(out, result.Inspect()+"\n")
			return err
		}
		if r.Format {
			_, err := io.WriteString(out, program.String())
			return err
		}
		b, err := json.MarshalIndent(program, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "%s\n", string(b))
		return err

	}
	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		_, err := fmt.Fprintf(out, "%#v\n", tok)
		if err != nil {
			return err
		}
	}
	return nil
}
