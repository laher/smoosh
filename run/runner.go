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
	_ "github.com/laher/smoosh/stdlib" //stdlib should always be loaded along with the evaluator ... how to do packages ... ?
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
func (r *Runner) RunFile(filename string, out io.Writer, stderr io.Writer) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return r.Run(f, out, stderr)
}

// Run runs an io.Reader as a single program
func (r *Runner) Run(rdr io.Reader, out io.Writer, stderr io.Writer) error {
	streams := object.Streams{
		Stdin:  rdr,
		Stdout: out,
		Stderr: stderr,
	}
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return fmt.Errorf("could not read: %v", err)
	}
	env := object.NewEnvironment(streams)
	macroEnv := object.NewEnvironment(streams)
	return r.runData(string(data), out, env, macroEnv)
}

func (r *Runner) runData(data string, out io.Writer, env, macroEnv *object.Environment) error {
	l := lexer.New(data)
	if r.Parse {
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			return errors.New(p.Errors()[0])
		}

		if r.Evaluate {
			evaluator.DefineMacros(program, macroEnv)
			expanded := evaluator.ExpandMacros(program, macroEnv)
			result := evaluator.Eval(expanded, env)
			if result == nil {
				return nil
			}

			switch r := result.(type) {
			case *object.Null:
				return nil
			case *object.Error:
				return fmt.Errorf("%s", r.Message)

			case *object.Pipes:
				pipes := r
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
		} else {
			// TODO detect type-checking errors
			// TODO detact macro errors
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
