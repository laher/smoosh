package ast

import (
	"testing"

	"github.com/laher/smoosh/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&AssignStatement{
				Token: token.Token{Type: token.VAR, Literal: "var"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}

	if program.String() != "var myVar = anotherVar \n" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
