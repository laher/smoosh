package evaluator

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/laher/smoosh/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1",
				len(args))
		}

		switch arg := args[0].(type) {
		case *object.Array:
			return &object.Integer{Value: int64(len(arg.Elements))}
		case *object.String:
			return &object.Integer{Value: int64(len(arg.Value))}
		default:
			return newError("argument to `len` not supported, got %s",
				args[0].Type())
		}
	},
	},
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	},
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		},
	},
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return NULL
		},
	},
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, arr.Elements[1:length])
				return &object.Array{Elements: newElements}
			}

			return NULL
		},
	},
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2",
					len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]

			return &object.Array{Elements: newElements}
		},
	},
	"pwd": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("wrong number of arguments. got=%d, want=0",
					len(args))
			}
			d, err := os.Getwd()
			if err != nil {
				return newError(err.Error())
			}
			//fmt.Println(d) //TODO make the repl print this somehow instead
			return &object.String{Value: d}
		},
	},
	"cd": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument to `cd` must be STRING, got %s",
					args[0].Type())
			}
			switch arg := args[0].(type) {
			case *object.String:
				err := os.Chdir(arg.Value)
				if err != nil {
					return newError(err.Error())
				}
				return NULL
			default:
				return newError("argument to `cd` not supported, got %s",
					args[0].Type())
			}

		},
	},
	"exit": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) > 1 {
				return newError("wrong number of arguments. got=%d, want=0/1",
					len(args))
			}
			if len(args) == 1 {
				if args[0].Type() != object.INTEGER_OBJ {
					return newError("argument to `exit` must be INTEGER, got %s",
						args[0].Type())
				}
				switch arg := args[0].(type) {
				case *object.Integer:
					os.Exit(int(arg.Value))
				default:
					return newError("argument to `exit` not supported, got %s",
						args[0].Type())
				}
			}
			os.Exit(0)
			return NULL
		},
	},

	"$": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			inputs := []string{}
			for i, arg := range args {
				if arg.Type() != object.STRING_OBJ {
					return newError("argument to `$` must be STRING, got %s",
						args[i].Type())
				}
				switch argT := arg.(type) {
				case *object.String:
					inputs = append(inputs, argT.Value)
				default:
					return newError("argument to `$` not supported, got %s",
						argT.Type())
				}

			}
			cmd := exec.Command(inputs[0], inputs[1:]...)

			out, err := cmd.StdoutPipe()
			if err != nil {
				return newError(err.Error())
			}
			errOut, err := cmd.StderrPipe()
			if err != nil {
				return newError(err.Error())
			}

			err = cmd.Start()
			//stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				return newError(err.Error())
			}
			return &object.Pipes{Out: out, Err: errOut, Wait: cmd.Wait}
			//return &object.String{Value: string(stdoutStderr)} // TODO new type 'pipe' with stdout/stderr as reader/writer
		},
	},
}
