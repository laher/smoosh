package checker

import (
	"bytes"
	"testing"

	"github.com/laher/smoosh/lexer"
	"github.com/laher/smoosh/object"
	"github.com/laher/smoosh/parser"
)

type test struct {
	input             string
	expected          object.ObjectType
	expectedErrorType string
}

func TestCheckProgram(t *testing.T) {
	tests := []test{
		{"5", object.INTEGER_OBJ, ""},
		{"5 + 6", object.INTEGER_OBJ, ""},
		{"5 + 5 + 5 + 5 - 10", object.INTEGER_OBJ, ""},
		{"\"helo\" + \" you\"", object.STRING_OBJ, ""},
		{"\"helo\" == 1", object.STRING_OBJ, TypeMismatch},
	}
	testThings(t, tests)
}

func TestBooleans(t *testing.T) {
	tests := []test{
		{input: "true"},
		{input: "false"},
		{input: "1 < 2"},
		{input: "1 > 2"},
		{input: "1 < 1"},
		{input: "1 > 1"},
		{input: "1 == 1"},
		{input: "1 != 1"},
		{input: "\"1\" == \"1\""},
		{input: "\"1\" != \"1\""},
		{input: "\"1\" != \"2\""},
		{input: "1 == 2"},
		{input: "1 != 2"},
		{input: "true == true"},
		{input: "false == false"},
		{input: "true == false"},
		{input: "true != false"},
		{input: "false != true"},
		{input: "(1 < 2) == true"},
		{input: "(1 < 2) == false"},
		{input: "(1 > 2) == true"},
		{input: "(1 > 2) == false"},
	}
	for i := range tests {
		tests[i].expected = object.BOOLEAN_OBJ
	}
	testThings(t, tests)
}

func TestIntegers(t *testing.T) {
	tests := []test{
		{input: "5"},
		{input: "10"},
		{input: "-5"},
		{input: "-10"},
		{input: "5 + 5 + 5 + 5 - 10"},
		{input: "2 * 2 * 2 * 2 * 2"},
		{input: "-50 + 100 + -50"},
		{input: "5 * 2 + 10"},
		{input: "5 + 2 * 10"},
		{input: "20 + 2 * -10"},
		{input: "50 / 2 * 2 + 10"},
		{input: "2 * (5 + 10)"},
		{input: "3 * 3 * 3 + 10"},
		{input: "3 * (3 * 3) + 10"},
		{input: "(5 + 10 * 2 + 15 / 3) * 2 + -10"},
	}
	for i := range tests {
		tests[i].expected = object.INTEGER_OBJ
	}
	testThings(t, tests)
}
func TestBangs(t *testing.T) {
	tests := []test{
		{input: "!true"},
		{input: "!false"},
		{input: "!5"},
		{input: "!!true"},
		{input: "!!false"},
		{input: "!!5"},
	}
	for i := range tests {
		tests[i].expected = object.BOOLEAN_OBJ
	}
	testThings(t, tests)
}

func TestIfElse(t *testing.T) {
	// TODO sum type?
	tests := []test{
		{input: "if (true) { 10 }"},
		{input: "if (false) { 10 }"},
		{input: "if (1) { 10 }"},
		{input: "if (1 < 2) { 10 }"},
		{input: "if (1 > 2) { 10 }"},
		{input: "if (1 > 2) { 10 } else { 20 }"},
		{input: "if (1 < 2) { 10 } else { 20 }"},
	}
	for i := range tests {
		tests[i].expected = object.INTEGER_OBJ
	}
	testThings(t, tests)

}

func TestReturnStatements(t *testing.T) {
	tests := []test{
		{input: "return 10;"},
		{input: "return 10; 9;"},
		{input: "return 2 * 5; 9;"},
		{input: "9; return 2 * 5; 9;"},
		{input: "if (10 > 1) { return 10; }"},
		{input: `
if (10 > 1) {
  if (10 > 1) {
    return 10;
  }

  return 1;
}
`,
		},
		{input: `
var f = fn(x) {
  return x;
  x + 10;
};
f(10);`,
		},
		{input: `
var f = fn(x) {
   var result = x + 10;
   return result;
   return 10;
};
f(11);`,
		},
	}
	for i := range tests {
		tests[i].expected = object.RETURN_VALUE_OBJ
	}
	testThings(t, tests)
}

func testThings(t *testing.T, tests []test) {
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
