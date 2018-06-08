package stdlib

import (
	"fmt"
	"path"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("dirname", &object.Builtin{
		Fn: dirname,
	})

}

func dirname(scope object.Scope, args ...object.Object) (object.Operation, error) {
	myArgs := []string{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			return nil, fmt.Errorf("flag %s not supported", arg.Name)
		case *object.String:
			d, err := Interpolate(scope.Env.Export(), arg.Value)
			if err != nil {
				return nil, fmt.Errorf(err.Error())
			}
			myArgs = append(myArgs, d)
		default:
			return nil, fmt.Errorf("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}
	if len(myArgs) < 1 {
		return nil, fmt.Errorf("Missing operand")
	}
	return func() object.Object {
		stdout, _ := getWriters(scope.Out)
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
	}, nil
}

func dirnames(files []string) []string {
	ret := []string{}
	for _, f := range files {
		dir := path.Dir(f)
		ret = append(ret, dir)
	}
	return ret
}
