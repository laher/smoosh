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
	inputs, err := interpolateArgs(scope.Env, args, false)
	if err != nil {
		return nil, err
	}
	if len(inputs) < 1 || len(inputs) > 2 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1 or 2",
			len(inputs))
	}
	return func() object.Object {
		w := strings.Join(inputs, " ")
		fmt.Fprintf(scope.Env.Streams.Stdout, "%s\n", w)
		return Null
	}, nil

}
