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
	fileNames, err := interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}

	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			return nil, fmt.Errorf("flag %s not supported", arg.Name)
		}
	}
	if len(fileNames) < 1 {
		return nil, fmt.Errorf("Missing operand")
	}
	return func() object.Object {
		ret := &object.Array{}
		myDirs := dirnames(fileNames)
		for _, dir := range myDirs {
			_, err := fmt.Fprintln(scope.Env.Streams.Stdout, dir)
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
