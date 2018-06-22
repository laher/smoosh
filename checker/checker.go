package checker

import (
	"errors"
	"fmt"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

type typeErrors struct {
	errors []error
}

func (e typeErrors) Error() string {
	return ""
}

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// Check compares types against expected types
func Check(node ast.Node, env *object.Environment) (object.ObjectType, error) {

	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return checkProgram(node, env)

	case *ast.BlockStatement:
		return checkBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return Check(node.Expression, env)

		// Expressions
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

func checkProgram(program *ast.Program, env *object.Environment) (object.ObjectType, error) {
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
	env *object.Environment,
) (object.ObjectType, error) {
	var (
		result object.ObjectType
		err    error
	)

	connectPipes(block.Statements)
	for _, statement := range block.Statements {
		result, err = Check(statement, env)
		if err != nil {
			return object.ObjectType(""), err
		}
		if shouldBePiping(statement) && !isPiping(statement) {
			//verify that the in.Out is non-nil
			err := errors.New("Call is not piping when it should be")
			return object.ObjectType(""), err
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
			return c.Out != nil && c.Out.Main != nil
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
