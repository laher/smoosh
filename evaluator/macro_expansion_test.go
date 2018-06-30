package evaluator

import (
	"testing"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/lexer"
	"github.com/laher/smoosh/object"
	"github.com/laher/smoosh/parser"
)

func TestDefineMacros(t *testing.T) {
	input := `
	var number = 1;
	var function = fn(x = 1, y = 0) { x + y };
	var mymacro = macro(x = 1, y = 0) { x + y; };
	var mymacroTwo = macro(x = 1, y = 0) { x + y; };
	`

	streams := object.Streams{}
	env := object.NewEnvironment(streams)
	program := testParseProgram(input)

	DefineMacros(program, env)

	if len(program.Statements) != 2 {
		t.Fatalf("Wrong number of statements. got=%d",
			len(program.Statements))
	}

	_, ok := env.Get("number")
	if ok {
		t.Fatalf("number should not be defined")
	}
	_, ok = env.Get("function")
	if ok {
		t.Fatalf("function should not be defined")
	}

	obj, ok := env.Get("mymacro")
	if !ok {
		t.Fatalf("macro not in environment.")
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		t.Fatalf("object is not Macro. got=%T (%+v)", obj, obj)
	}

	if len(macro.Parameters) != 2 {
		t.Fatalf("Wrong number of macro parameters. got=%d",
			len(macro.Parameters))
	}

	if macro.Parameters[0].String() != "(x = 1)" {
		t.Fatalf("parameter is not '(x = 1)'. got=%q", macro.Parameters[0])
	}
	if macro.Parameters[1].String() != "(y = 0)" {
		t.Fatalf("parameter is not '(y = 0)'. got=%q", macro.Parameters[1])
	}

	expectedBody := "  (x + y)\n"

	if macro.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, macro.Body.String())
	}
}

func TestExpandMacros(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
		        var infixExpression = macro() { quote(1 + 2); };

			infixExpression();
			`,
			`(1 + 2)`,
		},
		{
			`
			var reverse = macro(a, b) { quote(unquote(b) - unquote(a)); };

			reverse(2 + 2, 10 - 5);
			`,
			`(10 - 5) - (2 + 2)`,
		},
		{
			`
			var unless = macro(condition, consequence, alternative) {
				quote(if (!(unquote(condition))) {
					unquote(consequence);
				} else {
					unquote(alternative);
				});
			};

			unless(10 > 5, puts("not greater"), puts("greater"));
			`,
			`if (!(10 > 5)) { puts("not greater") } else { puts("greater") }`,
		},
	}

	streams := object.Streams{}
	for _, tt := range tests {
		expected := testParseProgram(tt.expected)
		program := testParseProgram(tt.input)

		env := object.NewEnvironment(streams)
		DefineMacros(program, env)
		expanded := ExpandMacros(program, env)

		if expanded.String() != expected.String() {
			t.Errorf("not equal. want=%q, got=%q",
				expected.String(), expanded.String())
		}
	}
}

func testParseProgram(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}
