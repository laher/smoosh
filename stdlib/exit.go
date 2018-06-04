package stdlib

import (
	"os"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("exit", exit)
}

func exit(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) > 1 {
		return object.NewError("wrong number of arguments. got=%d, want=0/1",
			len(args))
	}
	if len(args) == 1 {
		if args[0].Type() != object.INTEGER_OBJ {
			return object.NewError("argument to `exit` must be INTEGER, got %s",
				args[0].Type())
		}
		switch arg := args[0].(type) {
		case *object.Integer:
			os.Exit(int(arg.Value))
		default:
			return object.NewError("argument to `exit` not supported, got %s",
				args[0].Type())
		}
	}
	os.Exit(0)
	return Null
}
