package stdlib

import (
	"fmt"
	"path"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("dirname", &object.Builtin{
		Fn: dirname,
	})

}

func dirname(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
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
	stdout, _ := getWriters(out)
	ret := &object.Array{}
	myDirs := dirnames(myArgs)
	for _, dir := range myDirs {
		_, err := fmt.Fprintln(stdout, dir)
		if err != nil {
			return object.NewError(err.Error())
		}
		ret.Elements = append(ret.Elements, &object.String{Value: dir})
	}
	return ret
}

func dirnames(files []string) []string {
	ret := []string{}
	for _, f := range files {
		dir := path.Dir(f)
		ret = append(ret, dir)
	}
	return ret
}
