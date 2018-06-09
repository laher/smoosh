package stdlib

import (
	"fmt"
	"os"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("cd", cd)
}

func cd(scope object.Scope, args ...object.Object) (object.Operation, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.String:
		d, err := Interpolate(scope.Env.Export(), arg.Value)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		return func() object.Object {
			err = os.Chdir(d)
			if err != nil {
				return object.NewError(err.Error())
			}
			return Null
		}, nil
	default:
		return nil, fmt.Errorf("argument to `cd` not supported, got %s",
			args[0].Type())
	}

}
