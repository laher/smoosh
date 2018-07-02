package checker

import (
	"errors"
	"fmt"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
	"github.com/laher/smoosh/stdlib"
)

type typeErrors struct {
	errors []error
}

func (e typeErrors) Error() string {
	return ""
}

type TypeError struct {
	Type string
	Msg  string
	Line int
}

func (er TypeError) Error() string {
	return er.Msg
}

var (
	TypeMismatch      = "mismatched types"
	UnknownIdentifier = "unknown identifier"
	unknown           = object.ObjectType("unknown")
)

func newEnclosedEnvironment(outer *environment) *environment {
	env := NewEnvironment(outer.Streams)
	env.outer = outer
	return env
}

func NewEnvironment(streams object.Streams) *environment {
	s := make(map[string]object.ObjectType)
	return &environment{store: s, outer: nil, Streams: streams}
}

type environment struct {
	store   map[string]object.ObjectType
	outer   *environment
	Streams object.Streams
}

func (e *environment) Get(name string) (object.ObjectType, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *environment) Set(name string, val object.ObjectType) object.ObjectType {
	e.store[name] = val
	return val
}

// Check compares types against expected types
func Check(node ast.Node, env *environment) (object.ObjectType, error) {

	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return checkProgram(node, env)
	case *ast.BlockStatement:
		return checkBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return Check(node.Expression, env)
	case *ast.ReturnStatement:
		// TODO return types?
		r, err := Check(node.ReturnValue, env)
		if err != nil {
			return unknown, err
		}
		return r, nil
	case *ast.AssignStatement:
		t, err := Check(node.Value, env)
		if err != nil {
			return t, err
		}
		// check return types
		if v, ok := env.Get(node.Name.Value); ok {
			if t != v {
				return t, fmt.Errorf("type %s but expected %s", t, v)
			}
		}
		env.Set(node.Name.Value, t)
		return t, nil
		// Expressions
	case *ast.InfixExpression:

		left, err := Check(node.Left, env)
		if err != nil {
			return left, err
		}
		right, err := Check(node.Right, env)
		if err != nil {
			return right, err
		}
		// TODO special string / int cases?
		// TODO unknown operator?
		if left != right {
			return unknown, TypeError{Type: TypeMismatch, Msg: fmt.Sprintf("Infix [%s]: type [%s] does not equal [%s]. L: [%+v], R: [%+v]", node.Operator, left, right, node.Left, node.Right)}
		}
		switch node.Operator {
		case "==", "!=", ">", "<", "<=", ">=", "!":
			return object.BOOLEAN_OBJ, nil
		}
		// assume infix returns same type?
		return left, nil
	case *ast.PrefixExpression:
		right, err := Check(node.Right, env)
		if err != nil {
			return right, err
		}

		switch node.Operator {
		case "!":
			// check types
			if right != object.INTEGER_OBJ && right != object.BOOLEAN_OBJ {
				return unknown, TypeError{Type: TypeMismatch, Msg: fmt.Sprintf("Prefix [%s]: type [%s] is not an integer or bool. Value [%+v]", node.Operator, right, node.Right)}
			}
			return object.BOOLEAN_OBJ, nil
		case "-":
			// check types
			if right != object.INTEGER_OBJ {
				return unknown, TypeError{Type: TypeMismatch, Msg: fmt.Sprintf("Prefix [%s]: type [%s] is not an integer. Value [%+v]", node.Operator, right, node.Right)}
			}
			return object.INTEGER_OBJ, nil
		default:
			// TODO: what?
		}
		return right, nil

	case *ast.IfExpression:
		// TODO what to return?
		// IMO maybe VOID or a return value where appropriate
		r, err := Check(node.Condition, env)
		if err != nil {
			return unknown, err
		}
		r, err = Check(node.Consequence, env)
		if err != nil {
			return unknown, err
		}
		if node.Alternative != nil {
			r, err = Check(node.Alternative, env)
			if err != nil {
				return unknown, err
			}
		}
		return r, nil

	case *ast.Identifier:
		if val, ok := env.Get(node.Value); ok {
			return val, nil
		}

		if builtin, ok := stdlib.GetFn(node.Value); ok {
			return builtin.Type(), nil
		}

		return unknown, TypeError{Type: UnknownIdentifier, Msg: fmt.Sprintf("Unknown identifier [%s]", node.Value)}

	case *ast.CallExpression:
		if node.Function.TokenLiteral() == "quote" {
			// TODO macros
		}
		enclosedEnv := newEnclosedEnvironment(env)
		// TODO check args match signature ...
		for _, a := range node.Arguments {
			t, err := Check(a, env)
			if err != nil {
				return unknown, err
			}
			enclosedEnv.Set(a.String(), t)
		}
		r, err := Check(node.Function, enclosedEnv)
		if err != nil {
			return r, err
		}
		// apply extra flags during argument parsing
		switch r {
		case object.BUILTIN_OBJ:
			// TODO check arg types
			// node.
		case object.FUNCTION_OBJ:
			// TODO check arg types

		default:
			//			return unknown, TypeError{Type: TypeMismatch, Msg: fmt.Sprintf("[%s] is not a function", f)}
		}

		return r, nil
		// Literals
	case *ast.FunctionLiteral:
		env = newEnclosedEnvironment(env)
		for _, p := range node.Parameters {
			switch e := p.Right.(type) {
			case *ast.Boolean:
				env.Set(p.Left.(*ast.Identifier).Value, object.BOOLEAN_OBJ)
			case *ast.IntegerLiteral:
				env.Set(p.Left.(*ast.Identifier).Value, object.INTEGER_OBJ)
			case *ast.StringLiteral:
				env.Set(p.Left.(*ast.Identifier).Value, object.STRING_OBJ)
			case *ast.Identifier: // TODO detect type
				env.Set(p.Left.(*ast.Identifier).Value, object.INTEGER_OBJ)
			default:
				return unknown, fmt.Errorf("Unsupported 'right' type: [%T]", e)
			}
		}
		b, err := Check(node.Body, env)
		if err != nil {
			return unknown, err
		}
		return b, nil

	case *ast.IntegerLiteral:
		return object.INTEGER_OBJ, nil
	case *ast.StringLiteral:
		return object.STRING_OBJ, nil
	case *ast.BacktickLiteral:
		return object.BACKTICK_OBJ, nil
	case *ast.Boolean:
		return object.BOOLEAN_OBJ, nil
	default:
		return object.NULL_OBJ, fmt.Errorf("Not implemented: %T", node)
	}
}

func checkProgram(program *ast.Program, env *environment) (object.ObjectType, error) {
	var (
		result object.ObjectType
		err    error
	)

	connectPipes(program.Statements)
	for _, statement := range program.Statements {
		result, err = Check(statement, env)
		if err != nil {
			return result, err
		}
		if shouldBePiping(statement) && !isPiping(statement) {
			//verify that the in.Out is non-nil
			panic("Call is not piping when it should be")
		}

		switch result {
		case object.RETURN_VALUE_OBJ:
			// check in Call
			return result, nil
		case object.ERROR_OBJ:
			return result, nil
		default:

		}
	}

	return result, nil
}

func checkBlockStatement(
	block *ast.BlockStatement,
	env *environment,
) (object.ObjectType, error) {
	var (
		result object.ObjectType
		err    error
	)

	connectPipes(block.Statements)
	for _, statement := range block.Statements {
		result, err = Check(statement, env)
		if err != nil {
			return unknown, err
		}
		if shouldBePiping(statement) && !isPiping(statement) {
			//verify that the in.Out is non-nil
			err := errors.New("Call is not piping when it should be")
			return unknown, err
		}

		return result, nil
	}

	return result, nil
}

func shouldBePiping(statement ast.Statement) bool {
	if expS, ok := statement.(*ast.ExpressionStatement); ok {
		if c, ok := expS.Expression.(*ast.CallExpression); ok {
			return c.Out != nil
		}
	}
	return false
}

func isPiping(statement ast.Statement) bool {
	if expS, ok := statement.(*ast.ExpressionStatement); ok {
		if c, ok := expS.Expression.(*ast.CallExpression); ok {
			return c.Out != nil
			// TODO: c.Out.Main is not managed in the checker
			// What should we do here?
			// && c.Out.Main != nil
		}
	}
	return false
}

func connectPipes(statements []ast.Statement) {
	for i, this := range statements {
		if i > 0 {
			prev := statements[i-1]
			if expS, ok := this.(*ast.ExpressionStatement); ok {
				if p, ok := expS.Expression.(*ast.PipeExpression); ok {
					//this is a pipe ... hook up the outs and ins
					pipes := &ast.Pipes{}

					if expS, ok := prev.(*ast.ExpressionStatement); ok {
						if callS, ok := expS.Expression.(*ast.CallExpression); ok {
							callS.Out = pipes
						}
					}
					p.Destination.In = pipes
				}
			}
		}
	}

}
