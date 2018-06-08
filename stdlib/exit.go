package stdlib

import (
	"fmt"
	"os"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("exit", exit)
}

func exit(scope object.Scope, args ...object.Object) (object.Operation, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=0/1",
			len(args))
	}
	code := 0
	if len(args) == 1 {
		if args[0].Type() != object.INTEGER_OBJ {
			return nil, fmt.Errorf("argument to `exit` must be INTEGER, got %s",
				args[0].Type())
		}
		switch arg := args[0].(type) {
		case *object.Integer:
			code = int(arg.Value)
		default:
			return nil, fmt.Errorf("argument to `exit` not supported, got %s",
				args[0].Type())
		}
	}

	return func() object.Object {
		os.Exit(code)
		return Null
	}, nil
}
