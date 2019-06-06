package run

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/lexer"
	"github.com/laher/smoosh/parser"
	"github.com/laher/smoosh/stdlib"
)

func (r *Runner) promptExecutor(data string) {
	err := r.runData(data)
	if err != nil {
		fmt.Println(err)
	}
}

func (r *Runner) promptCompleter(t prompt.Document) []prompt.Suggest {
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
		/*
			for _, k := range keywords {
				if strings.Contains(k.Text, lastStatement.TokenLiteral()) {
					ret = append(ret, k)
				}
			}*/
	}
	return ret
}
