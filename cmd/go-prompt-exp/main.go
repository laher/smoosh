package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/evaluator"
	"github.com/laher/smoosh/lexer"
	"github.com/laher/smoosh/object"
	"github.com/laher/smoosh/parser"
	"github.com/laher/smoosh/stdlib"
)

type runner struct {
	env      *object.Environment
	macroEnv *object.Environment
	streams  object.Streams
}

func (r *runner) executor(data string) {
	err := r.runData(data)
	if err != nil {
		fmt.Println(err)
	}
}

func (rnr *runner) runData(data string) error {
	l := lexer.New(data)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return errors.New(p.Errors()[0])
	}
	evaluator.DefineMacros(program, rnr.macroEnv)
	expanded := evaluator.ExpandMacros(program, rnr.macroEnv)
	result := evaluator.Eval(expanded, rnr.env)
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
		cmdOut, err := ioutil.ReadAll(pipes.Main)
		if err != nil {
			return err
		}
		err = pipes.Wait()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(rnr.streams.Stdout, "%s", cmdOut)
		return err
	}
	_, err := io.WriteString(rnr.streams.Stdout, result.Inspect()+"\n")
	return err
}

var keywords = []prompt.Suggest{
	{Text: "if"},
	{Text: "else"},
	{Text: "for"},
}

func (r *runner) completer(t prompt.Document) []prompt.Suggest {
	text := t.Text
	incomplete := ""
	program := &ast.Program{}
	for len(text) > 0 { // keep trying until the shortest parseable string
		l := lexer.New(text)
		p := parser.New(l)
		program = p.ParseProgram()
		if len(p.Errors()) == 0 {
			//	fmt.Println("text: ", text, ". incomplete: ", incomplete)
			break
		}
		incomplete = text[len(text)-1:] + incomplete
		text = text[:len(text)-1]
	}

	ret := []prompt.Suggest{}
	if len(program.Statements) > 0 {
		lastStatement := program.Statements[len(program.Statements)-1]
		if strings.Contains(incomplete, "(") {
			b, ok := stdlib.GetFn(lastStatement.TokenLiteral())
			if ok {
				//fmt.Println("text: ", text, ". incomplete: ", incomplete)
				// offer to close brackets at any time
				if !strings.HasSuffix(incomplete, ",") {
					ret = append(ret, prompt.Suggest{Text: t.Text + ")"})
				}
				if strings.HasSuffix(incomplete, "(") || strings.HasSuffix(incomplete, ",") {
					// an arg:
					for _, f := range b.Flags {
						if !strings.Contains(incomplete, f.Name) {
							ret = append(ret, prompt.Suggest{Text: t.Text + f.Name, Description: f.Help})
						}
					}
				} else {
					// after an arg:
					ret = append(ret, prompt.Suggest{Text: t.Text + ","})
				}
			}
		}
		for _, b := range stdlib.ListBuiltins() {
			if strings.Contains(b, lastStatement.TokenLiteral()) {
				bi, _ := stdlib.GetFn(b)

				ret = append(ret, prompt.Suggest{Text: b + "(", Description: bi.Help})
			}
		}
		for _, k := range keywords {
			if strings.Contains(k.Text, lastStatement.TokenLiteral()) {
				ret = append(ret, k)
			}
		}
	}
	/*
		words := strings.Split(t.Text, " ")
		lastWord := words[len(words)-1]
		if strings.Contains(lastWord, "(") {
			ret = append(ret, prompt.Suggest{Text: t.Text + ")"})
		}
		for _, k := range keywords {
			if strings.Contains(k.Text, lastWord) {
				ret = append(ret, k)
			}
		}*/
	return ret
}

func main() {

	// REPL:
	streams := object.Streams{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	env := object.NewEnvironment(streams)
	macroEnv := object.NewEnvironment(streams)

	r := &runner{
		env:      env,
		macroEnv: macroEnv,
		streams:  streams,
	}
	p := prompt.New(
		r.executor,
		r.completer,
	)
	p.Run()
}
