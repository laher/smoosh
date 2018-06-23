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
		input             string
		expected          object.ObjectType
		expectedErrorType string
	}{
		{"5", object.INTEGER_OBJ, ""},
		{"5 + 6", object.INTEGER_OBJ, ""},
		{"5 + 5 + 5 + 5 - 10", object.INTEGER_OBJ, ""},
		{"\"helo\" + \" you\"", object.STRING_OBJ, ""},
		{"\"helo\" == 1", object.STRING_OBJ, TypeMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			evaluated, err := testCheck(tt.input)
			if tt.expectedErrorType != "" {
				if err == nil {
					t.Error("checkProgram should have errored. Returned", evaluated)
					return
				}
				if typeErr, ok := err.(TypeError); ok {
					if typeErr.Type != tt.expectedErrorType {
						t.Error("checkProgram returned wrong error. Returned", typeErr, "expected:", tt.expectedErrorType)
						return
					}
					return
				} else {
					t.Error("checkProgram returned wrong error. Returned", err, "expected:", tt.expectedErrorType)
				}
			} else {
				if err != nil {
					t.Error("unexpected error checking program", err)
					return
				}
			}
			if evaluated != object.ObjectType(tt.expected) {
				t.Errorf("Error checking program: [%s]. Expected [%s]", evaluated, object.ObjectType(tt.expected))
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
	env := newEnvironment(streams)
	return Check(program, env)
}
