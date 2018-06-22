package checker

import (
	"bytes"
	"testing"

	"github.com/laher/smoosh/lexer"
	"github.com/laher/smoosh/object"
	"github.com/laher/smoosh/parser"
)

func TestCheckProgram(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5", "5"},
		{"5 + 5 + 5 + 5 - 10", "10"},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			evaluated, err := testCheck(tt.input)
			if err != nil {
				t.Error("Error checking program", err)
				return
			}
			if evaluated != object.ObjectType(tt.expected) {
				t.Error("Error checking program", err)
			}
		})
		//testIntegerObject(t, evaluated, tt.expected)
	}
}

func testCheck(input string) (object.ObjectType, error) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	streams := object.Streams{
		Stdin:  bytes.NewBuffer([]byte(input)),
		Stdout: bytes.NewBuffer([]byte{}),
		Stderr: bytes.NewBuffer([]byte{}),
	}
	env := object.NewEnvironment(streams)
	return Check(program, env)
}
