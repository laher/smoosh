package run

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/lexer"
	"github.com/laher/smoosh/parser"
	"github.com/laher/smoosh/stdlib"
	"github.com/laher/smoosh/token"
)

func (r *Runner) promptExecutor(data string) {
	err := r.runData(data)
	if err != nil {
		fmt.Println(err)
	}
}

func (r *Runner) promptCompleter(t prompt.Document) []prompt.Suggest {

	var (
		text       = t.Text
		incomplete = ""
		program    *ast.Program
		ret        = []prompt.Suggest{}
	)
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

	if program != nil && len(program.Statements) > 0 {
		lastStatement := program.Statements[len(program.Statements)-1]
		if incomplete != "" {
			var lastToken token.Token
			l := lexer.New(incomplete)
		labl:
			for {
				tok := l.NextToken()
				switch tok.Type {
				case token.EOF:
					break labl
				default:
					//		fmt.Printf("Token: %+v. Lit: [%s]\n", tok, tok.Literal)
				}
				lastToken = tok
			}
			if lastToken.Type == token.STRING {
				// NOTE: it would be nice to use a parser here which knows about string-quotes
				if !strings.HasSuffix(incomplete, `"`) {
					ret = append(ret, prompt.Suggest{Text: t.Text + `"`})
					return ret
				}
			}

			if strings.Contains(incomplete, "(") { // incomplete
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
		} else {
			for _, b := range stdlib.ListBuiltins() {
				if strings.Contains(b, lastStatement.TokenLiteral()) {
					bi, _ := stdlib.GetFn(b)
					h := strings.Split(bi.Help, "\n")
					ret = append(ret, prompt.Suggest{Text: b + "(", Description: h[0]})
				}
			}

			for _, k := range token.ListKeywords() {
				if strings.Contains(k, lastStatement.TokenLiteral()) {
					ret = append(ret, prompt.Suggest{Text: k})
				}
			}
		}
	}
	return ret
}
