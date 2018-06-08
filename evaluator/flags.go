package evaluator

import (
	"fmt"

	"github.com/laher/smoosh/object"
)

func flagFn(flag *object.Flag) object.Object {
	return &object.Builtin{
		Fn: func(scope object.Scope, args ...object.Object) (object.Operation, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("Value not supplied for flag [%s]", flag.Name)
			}
			if flag.ParamType != args[0].Type() {
				return nil, fmt.Errorf("Unexpected value type [%v] for flag [%s]. Expected [%v]", args[0].Type(), flag.Name, flag.ParamType)
			}
			return func() object.Object {
				// takes in an arg and adds it into the Flag.Param
				flag.Param = args[0]
				return flag
			}, nil
		},
	}
}
