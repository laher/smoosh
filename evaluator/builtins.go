package evaluator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/alecthomas/template"
	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
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
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return NULL
		},
	},
	"first": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
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
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
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
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
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
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
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
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
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
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
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
				d, err := interpolate(env.Export(), arg.Value)
				if err != nil {
					return newError(err.Error())
				}
				err = os.Chdir(d)
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
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
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
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			inputs := []string{}
			envV := env.Export()
			for i, arg := range args {
				if arg.Type() != object.STRING_OBJ {
					return newError("argument to `$` must be STRING, got %s",
						args[i].Type())
				}
				switch argT := arg.(type) {
				case *object.String:
					input, err := interpolate(envV, argT.Value)
					if err != nil {
						return newError("cannot parse arg for interpolation - %s",
							err)
					}
					inputs = append(inputs, input)
				default:
					return newError("argument to `$` not supported, got %s",
						argT.Type())
				}

			}
			cmd := exec.Command(inputs[0], inputs[1:]...)
			if in != nil {
				cmd.Stdin = in.Out
			}
			if out != nil {
				stdOut, err := cmd.StdoutPipe()
				if err != nil {
					return newError(err.Error())
				}
				errOut, err := cmd.StderrPipe()
				if err != nil {
					return newError(err.Error())
				}
				out.Out = stdOut
				out.Err = errOut
				out.Wait = cmd.Wait
				err = cmd.Start()
				if err != nil {
					return newError(err.Error())
				}
				p := object.Pipes(*out)
				return &p
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				return newError(err.Error())
			}
			return NULL
		},
	},
	"echo": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2",
					len(args))
			}
			inputs, err := interpolateArgsAsStrings(env, args)
			if err != nil {
				return newError(err.Error())
			}
			f := os.Stdout
			if out != nil {
				// ?
			}
			for _, w := range inputs {
				fmt.Fprintf(f, "%s", w)
			}
			return NULL
		},
	},
	"w": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2",
					len(args))
			}
			inputs := []string{}
			envV := env.Export()
			for i, arg := range args {
				if arg.Type() != object.STRING_OBJ {
					return newError("argument to `$` must be STRING, got %s",
						args[i].Type())
				}
				switch argT := arg.(type) {
				case *object.String:
					input, err := interpolate(envV, argT.Value)
					if err != nil {
						return newError("cannot parse arg for interpolation - %s",
							err)
					}
					inputs = append(inputs, input)
				default:
					return newError("argument to `$` not supported, got %s",
						argT.Type())
				}

			}
			if in == nil {
				return NULL
			}

			if inputs[0] != "" {
				f, err := os.Create(inputs[0])
				if err != nil {
					return newError(err.Error())
				}
				if _, err := io.Copy(f, in.Out); err != nil {
					return newError(err.Error())
				}
			}
			if len(inputs) > 1 && inputs[1] != "" && in.Err != nil {
				f, err := os.Create(inputs[1])
				if err != nil {
					return newError(err.Error())
				}
				if _, err := io.Copy(f, in.Err); err != nil {
					return newError(err.Error())
				}
			}

			return NULL
		},
	},
	"r": &object.Builtin{
		Fn: func(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2",
					len(args))
			}
			inputs, err := interpolateArgsAsStrings(env, args)
			if err != nil {
				return newError(err.Error())
			}
			f, err := os.Open(inputs[0])
			if err != nil {
				return newError(err.Error())
			}
			if out != nil {
				out.Out = f
				return NULL
			}
			if _, err := io.Copy(os.Stdout, f); err != nil {
				return newError(err.Error())
			}
			return NULL
		},
	},
}

func interpolateArgsAsStrings(env *object.Environment, args []object.Object) ([]string, error) {
	inputs := []string{}
	envV := env.Export()
	for i, arg := range args {
		if arg.Type() != object.STRING_OBJ {
			return nil, fmt.Errorf("argument must be STRING, got %s",
				args[i].Type())
		}
		switch argT := arg.(type) {
		case *object.String:
			input, err := interpolate(envV, argT.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse arg for interpolation - %s",
					err)
			}
			inputs = append(inputs, input)
		default:
			return nil, fmt.Errorf("argument not supported, got %s",
				argT.Type())
		}
	}
	return inputs, nil
}

func interpolate(envV map[string]interface{}, value string) (string, error) {
	tmpl, err := template.New("test").Parse(value)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buf, envV)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
