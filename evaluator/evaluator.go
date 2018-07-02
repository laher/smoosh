package evaluator

import (
	"fmt"
	"io"
	"sync"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
	"github.com/laher/smoosh/stdlib"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {

	switch node := node.(type) {

	// Statements
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.AssignStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		if v, ok := env.Get(node.Name.Value); ok {
			if val.Type() != v.Type() {
				return newError("type %s but expected %s", val.Type(), v.Type())
			}
		}
		env.Set(node.Name.Value, val)

	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.BacktickLiteral:
		return &object.BacktickExpression{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

	case *ast.CallExpression:
		if node.Function.TokenLiteral() == "quote" {
			return quote(node.Arguments[0], env)
		}
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		enclosedEnv := object.NewEnclosedEnvironment(env)
		// apply extra flags during argument parsing
		switch fn := function.(type) {
		case *object.Builtin:
			if len(fn.Flags) > 0 {
				for i := range fn.Flags {
					switch fn.Flags[i].ParamType {
					case object.INTEGER_OBJ, object.STRING_OBJ:
						enclosedEnv.Set(fn.Flags[i].Name, flagFn(&fn.Flags[i]))
					case object.BOOLEAN_OBJ, object.ObjectType(""):
						// TODO maybe allow flags to be passed as 'false' (where default is true)... OR just force them to always default to false
						// enclosedEnv.Set(fn.Flags[i].Name, flagFn(&fn.Flags[i]))
						enclosedEnv.Set(fn.Flags[i].Name, &fn.Flags[i])
					default:
						return object.NewError("Unexpected flag ParamType [%v]", fn.Flags[i].ParamType)
					}
				}
			}
		}
		args := evalExpressions(node.Arguments, enclosedEnv)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args, node.In, node.Out, env, node.Function.TokenLiteral())

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.PipeExpression:
		return Eval(node.Destination, env)

	case *ast.RangeExpression:
		return evalRangeExpression(node, env)

	case *ast.ForExpression:
		return evalForExpression(node, env)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	connectPipes(program.Statements)
	for _, statement := range program.Statements {
		result = Eval(statement, env)
		if shouldBePiping(statement) && !isPiping(statement) {
			//verify that the in.Out is non-nil
			panic("Call is not piping when it should be")
		}

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(
	block *ast.BlockStatement,
	env *object.Environment,
) object.Object {
	var result object.Object

	connectPipes(block.Statements)
	for _, statement := range block.Statements {
		result = Eval(statement, env)
		if shouldBePiping(statement) && !isPiping(statement) {
			//verify that the in.Out is non-nil
			panic("Call is not piping when it should be")
		}

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
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

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)

	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

}

func evalIfExpression(
	ie *ast.IfExpression,
	env *object.Environment,
) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalRangeExpression(
	fe *ast.RangeExpression, env *object.Environment,
) object.Object {

	it := Eval(fe.Iterator, env)

	if it.Type() != object.ARRAY_OBJ {
		return newError("only arrays supported: " + string(it.Type()))
	}

	var ret object.Object
	for i, v := range it.(*object.Array).Elements {
		extendedEnv := object.NewEnclosedEnvironment(env)
		extendedEnv.Set(fe.Identifier.String(), &object.Integer{Value: int64(i)})
		if fe.Iteree != nil {
			extendedEnv.Set(fe.Iteree.String(), v)
		}
		ret = Eval(fe.Body, extendedEnv)
	}
	return ret
}

func evalForExpression(
	fe *ast.ForExpression, env *object.Environment,
) object.Object {

	init := Eval(fe.Init, env)
	if init != nil && init.Type() == object.ERROR_OBJ {
		return newError("Error returned from FOR init: " + string(init.Type()))
	}

	var ret object.Object
	i := 0
	for {
		i++
		c := Eval(fe.Condition, env)
		if c.Type() != object.BOOLEAN_OBJ {
			return newError("Error returned from FOR condition: " + string(c.Type()))
		}
		if !c.(*object.Boolean).Value {
			break
		}
		extendedEnv := object.NewEnclosedEnvironment(env)
		//TODO FOR varables
		ret = Eval(fe.Body, extendedEnv)
		Eval(fe.After, env)
	}
	return ret
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := stdlib.GetFn(node.Value); ok {
		return builtin
	}

	return newError("identifier not found: %v", node.Value)
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object, in, out *ast.Pipes, env *object.Environment, tokenLiteral string) object.Object {
	defer func() {
		if in != nil {
			// defer guarantees this runs AFTER applyFunction.
			// go guarantees that it doesn't block the next call
			//go in.WaitAndClose() // ensure cleanup always happens
		}
	}()
	switch fn := fn.(type) {

	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		myEnv := env
		if in != nil || out != nil {
			myEnv = object.NewEnclosedEnvironment(env)
			if in != nil {
				if in.Main == nil {
					panic("in.Main is nil")
				}
				myEnv.Streams.Stdin = in.Main
			}
			if out != nil {
				r, w := io.Pipe()
				myEnv.Streams.Stdout = w // this will be closed by the evaluator
				out.Main = r
				r, w = io.Pipe()
				myEnv.Streams.Stderr = w // this will be closed by the evaluator
				out.Err = r
			}
		}
		op, err := fn.Fn(object.Scope{
			Env: myEnv,
			In:  in,
			Out: out,
		}, args...)
		if err != nil {
			return object.NewError(err.Error())
		}
		if out != nil {
			doAsync(op, out, myEnv.Streams.Stderr)
			return NULL
		}

		return op()
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func doAsync(op object.Operation, out *ast.Pipes, stderr io.Writer) {
	wg := sync.WaitGroup{}
	out.Wait = func() error {
		wg.Wait()
		return nil
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		o := op()
		if oe, ok := o.(*object.Error); ok {
			fmt.Fprintf(stderr, "Error returned from piped func: [%s]\n", oe.Message)
		}
		err := out.Main.Close()
		if err != nil {
			fmt.Fprintf(stderr, "Error closing pipe: [%s]\n", err)
		}
	}()

}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		n := param.Left.(*ast.Identifier)
		if len(args) > paramIdx {
			env.Set(n.Value, args[paramIdx])
			continue
		}
		if param.Operator == "=" {
			val := Eval(param.Right, fn.Env)
			env.Set(n.Value, val)
		}
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}
