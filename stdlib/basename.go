package stdlib

import (
	"fmt"
	"path"
	"strings"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("basename", &object.Builtin{
		Fn: basename,
	})

}

func basename(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	var relativeTo, inputPath string
	myArgs := []string{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			return object.NewError("flag %s not supported", arg.Name)
		case *object.String:
			d, err := Interpolate(env.Export(), arg.Value)
			if err != nil {
				return object.NewError(err.Error())
			}
			myArgs = append(myArgs, d)
		default:
			return object.NewError("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}
	if len(myArgs) < 1 {
		return object.NewError("Missing operand")
	}
	if len(myArgs) > 1 {
		relativeTo = myArgs[0]
		inputPath = myArgs[1]
	} else {
		inputPath = myArgs[0]
	}

	base := basenameFile(inputPath, relativeTo)
	stdout, _ := getWriters(out)
	_, err := fmt.Fprintln(stdout, base)
	if err != nil {
		return object.NewError(err.Error())
	}
	return &object.String{Value: base}
}

func basenameFile(inputPath, relativeTo string) string {
	if relativeTo != "" {
		last := strings.LastIndex(relativeTo, inputPath)
		return inputPath[:last]

	}
	return path.Base(inputPath)
}
