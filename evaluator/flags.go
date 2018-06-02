package evaluator

import (
	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func flagFn(flag *object.Flag) object.Object {
	return &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) < 1 {
				return object.NewError("Value not supplied for flag [%s]", flag.Name)
			}
			if flag.ParamType != args[0].Type() {
				return object.NewError("Unexpected value type [%v] for flag [%s]. Expected [%v]", args[0].Type(), flag.Name, flag.ParamType)
			}
			// takes in an arg and adds it into the Flag.Param
			flag.Param = args[0]
			return flag
		},
	}
}
