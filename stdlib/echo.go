package stdlib

import (
	"fmt"
	"strings"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("echo", echo)
}

func echo(scope object.Scope, args ...object.Object) (object.Operation, error) {
	envV := scope.Env.Export()
	inputs := []string{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.String:
			input, err := Interpolate(envV, arg.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse arg for interpolation - %s",
					err)
			}
			inputs = append(inputs, input)
		default:
			inputs = append(inputs, arg.Inspect())
		}
	}
	if len(inputs) < 1 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1 or more",
			len(inputs))
	}
	return func() object.Object {
		w := strings.Join(inputs, " ")
		fmt.Fprintf(scope.Env.Streams.Stdout, "%s\n", w)
		return Null
	}, nil

}
