package stdlib

import (
	"os"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("cd", cd)
}

func cd(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.STRING_OBJ {
		return object.NewError("argument to `cd` must be STRING, got %s",
			args[0].Type())
	}
	switch arg := args[0].(type) {
	case *object.String:
		d, err := Interpolate(env.Export(), arg.Value)
		if err != nil {
			return object.NewError(err.Error())
		}
		err = os.Chdir(d)
		if err != nil {
			return object.NewError(err.Error())
		}
		return Null
	default:
		return object.NewError("argument to `cd` not supported, got %s",
			args[0].Type())
	}

}
