package stdlib

import (
	"bufio"
	"fmt"
	"path"
	"strings"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("basename", &object.Builtin{
		Fn: basename,
		Help: `Usage: basename NAME [SUFFIX]
  or:  basename OPTION... NAME...
Print NAME with any leading directory components removed.
If specified, also remove a trailing SUFFIX.`,
	})

}

func basename(scope object.Scope, args ...object.Object) (object.Operation, error) {
	var relativeTo, inputPath string
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
		if scope.In == nil {
			return nil, fmt.Errorf("Missing operand")
		}
	}
	/*
		if scope.Out != nil {
			r, w := io.Pipe()
			scope.Env.Streams.Stdout = w // this will be closed by the evaluator
			scope.Out.Main = r
		}
	*/

	return func() object.Object {
		if scope.In != nil {
			bio := bufio.NewReader(scope.Env.Streams.Stdin)
			//defer bio.Close()
			line, _, err := bio.ReadLine()
			if err != nil {
				return object.NewError(err.Error())
			}
			myArgs = strings.Split(string(line), " ")
		}

		if len(myArgs) > 1 {
			relativeTo = myArgs[0]
			inputPath = myArgs[1]
		} else {
			inputPath = myArgs[0]
		}

		base := basenameFile(inputPath, relativeTo)
		//_, err := fmt.Fprintln(scope.Env.Streams.Stdout, base)
		_, err := fmt.Fprintln(scope.Env.Streams.Stdout, base)
		if err != nil {
			return object.NewError(err.Error())
		}
		//return &object.String{Value: base}
		return Null
	}, nil
}

func basenameFile(inputPath, relativeTo string) string {
	if relativeTo != "" {
		last := strings.LastIndex(relativeTo, inputPath)
		return inputPath[:last]

	}
	return path.Base(inputPath)
}
