package stdlib

import (
	"bufio"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("ret", ret)
}

func ret(scope object.Scope, args ...object.Object) (object.Operation, error) {
	return func() object.Object {
		if scope.Env.Streams.Stdin == nil {
			if scope.In != nil {
				if scope.In.Main != nil {
					return object.NewError("Stdin not available (piping) - scope.In.Main is non-nil")
				}
				return object.NewError("Stdin not available (piping) - scope.In.Main is nil")
			}
			return object.NewError("Stdin not available (not piping)")
		}
		scanner := bufio.NewScanner(scope.Env.Streams.Stdin)
		scanner.Scan()
		val := scanner.Text()
		if err := scanner.Err(); err != nil {
			return object.NewError(err.Error())
		}
		return &object.String{Value: val}
	}, nil
}
