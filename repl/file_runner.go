package repl

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

func (r *Runner) RunAll(rdr io.Reader) error {
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return fmt.Errorf("could not read: %v", err)
	}

	return r.runData(data)
}

func (r *Runner) RunFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read %s: %v", filename, err)
	}

	return r.runData(data)
}

func (r *Runner) runData(data []byte) error {
	l := lexer.New(string(data))
	out := os.Stdout
	if r.Parse {
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			return errors.New(p.Errors()[0])
		}
		if r.Evaluate {
			env := object.NewEnvironment()
			result := evaluator.Eval(program, env)
			if _, ok := result.(*object.Null); ok {
				return nil
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
