package stdlib

import (
	"io"
	"os"
	"sync"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func write(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError("wrong number of arguments. got=%d, want=1 or 2",
			len(args))
	}
	inputs := []string{}
	envV := env.Export()
	for i, arg := range args {
		if arg.Type() != object.STRING_OBJ {
			return object.NewError("argument to `$` must be STRING, got %s",
				args[i].Type())
		}
		switch argT := arg.(type) {
		case *object.String:
			input, err := Interpolate(envV, argT.Value)
			if err != nil {
				return object.NewError("cannot parse arg for interpolation - %s",
					err)
			}
			inputs = append(inputs, input)
		default:
			return object.NewError("argument to `$` not supported, got %s",
				argT.Type())
		}
	}
	if in == nil {
		return Null
	}
	if inputs[0] != "" {
		f, err := os.Create(inputs[0])
		if err != nil {
			return object.NewError(err.Error())
		}
		defer func() {
			if in.Wait != nil {
				in.Wait()
			}
			f.Close()
		}()
		go func() {
			if _, err := io.Copy(f, in.Out); err != nil {
				//return object.NewError(err.Error())
				panic(err.Error())
			}
		}()
	}
	if len(inputs) > 1 && inputs[1] != "" && in.Err != nil {
		f, err := os.Create(inputs[1])
		if err != nil {
			return object.NewError(err.Error())
		}
		defer func() {
			if in.Wait != nil {
				in.Wait()
			}
			f.Close()
		}()
		go func() {
			if _, err := io.Copy(f, in.Err); err != nil {
				//return object.NewError(err.Error())
				panic(err.Error())
			}
		}()
	}
	return Null
}

func read(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError("wrong number of arguments. got=%d, want=1 or 2",
			len(args))
	}
	inputs, err := InterpolateArgsAsStrings(env, args)
	if err != nil {
		return object.NewError(err.Error())
	}
	f, err := os.Open(inputs[0])
	if err != nil {
		return object.NewError(err.Error())
	}
	stdout, _ := getWriters(out)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		if _, err := io.Copy(stdout, f); err != nil {
			//return object.NewError(err.Error())
			panic(err.Error())
		}
		wg.Done()
	}()
	if out != nil {
		out.Wait = func() error {
			wg.Wait()
			return f.Close()
		}
	}
	return Null
}
