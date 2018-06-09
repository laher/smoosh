package stdlib

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("w", &object.Builtin{
		Fn:    write,
		Flags: []object.Flag{{Name: "a"}}})
	RegisterFn("r", read)
}

func write(scope object.Scope, args ...object.Object) (object.Operation, error) {
	inputs, err := interpolateArgs(scope.Env, args, false)
	if err != nil {
		return nil, err
	}
	if len(inputs) < 1 || len(inputs) > 2 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1 or 2",
			len(args))
	}

	app := false
	for i := range args {
		switch argT := args[i].(type) {
		case *object.Flag:
			switch argT.Name {
			case "a":
				app = true
			default:
				return nil, fmt.Errorf("flag %s not supported", argT.Name)
			}
		}
	}
	if scope.In == nil {
		return nil, fmt.Errorf("Nothing to write. 'w' expects an input stream")
	}
	return func() object.Object {
		opts := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		if app {
			opts = os.O_APPEND | os.O_WRONLY
		}
		// stdout
		if inputs[0] != "" {
			f, err := os.OpenFile(inputs[0], opts, 0666)
			if err != nil {
				return object.NewError(err.Error())
			}
			defer func() {
				if scope.In.Wait != nil {
					scope.In.Wait()
				}
				f.Close()
			}()
			if _, err := io.Copy(f, scope.In.Main); err != nil {
				if err != io.ErrClosedPipe {
					return object.NewError(err.Error())
				}
			}
		}
		// stderr
		if len(inputs) > 1 && inputs[1] != "" && scope.In.Err != nil {
			f, err := os.OpenFile(inputs[1], opts, 0666)
			if err != nil {
				return object.NewError(err.Error())
			}
			defer func() {
				if scope.In.Wait != nil {
					scope.In.Wait()
				}
				f.Close()
			}()
			if _, err := io.Copy(f, scope.In.Err); err != nil {
				//return object.NewError(err.Error())
				if err != io.ErrClosedPipe {
					return object.NewError(err.Error())
				}
			}
		}
		return Null
	}, nil
}

func read(scope object.Scope, args ...object.Object) (object.Operation, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1 or 2",
			len(args))
	}
	inputs, err := interpolateArgs(scope.Env, args, false)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(inputs[0])
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	if scope.Out != nil {
		scope.Out.Wait = func() error {
			wg.Wait()
			return f.Close()
		}
	}
	return func() object.Object {
		if _, err := io.Copy(scope.Env.Streams.Stdout, f); err != nil {
			//return object.NewError(err.Error())
			panic(err.Error())
		}
		wg.Done()
		return Null
	}, nil
}
