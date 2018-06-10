package stdlib

import (
	"fmt"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("help", &object.Builtin{
		Fn: help,
	})
}

func help(scope object.Scope, args ...object.Object) (object.Operation, error) {
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Builtin:
			h := arg.Help
			for _, flag := range arg.Flags {
				if flag.ParamType == "" {
					flag.ParamType = object.BOOLEAN_OBJ
				}
				h = fmt.Sprintf("%s\n%s (%s):\t%s", h, flag.Name, flag.ParamType, flag.Help)
			}
			return func() object.Object {
				return &object.String{
					Value: h,
				}
			}, nil
		}
	}
	return func() object.Object {
		return &object.String{Value: "use help(fn) to find out more about a particular function fn"}
	}, nil
}
