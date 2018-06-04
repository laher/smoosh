package stdlib

import (
	"io"
	"os"
	"sync"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("w", &object.Builtin{
		Fn:    write,
		Flags: []object.Flag{{Name: "a"}}})
	RegisterFn("r", read)
}

func write(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	inputs, err := interpolateArgs(env, args, false)
	if err != nil {
		return object.NewError(err.Error())
	}
	if len(inputs) < 1 || len(inputs) > 2 {
		return object.NewError("wrong number of arguments. got=%d, want=1 or 2",
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
				return object.NewError("flag %s not supported", argT.Name)
			}
		}
	}
	if in == nil {
		return object.NewError("Nothing to write. 'w' expects an input stream")
	}
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
			if in.Wait != nil {
				in.Wait()
			}
			f.Close()
		}()
		if _, err := io.Copy(f, in.Out); err != nil {
			if err != io.ErrClosedPipe {
				return object.NewError(err.Error())
			}
		}
	}
	// stderr
	if len(inputs) > 1 && inputs[1] != "" && in.Err != nil {
		f, err := os.OpenFile(inputs[1], opts, 0666)
		if err != nil {
			return object.NewError(err.Error())
		}
		defer func() {
			if in.Wait != nil {
				in.Wait()
			}
			f.Close()
		}()
		if _, err := io.Copy(f, in.Err); err != nil {
			//return object.NewError(err.Error())
			if err != io.ErrClosedPipe {
				return object.NewError(err.Error())
			}
		}
	}
	return Null
}

func read(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError("wrong number of arguments. got=%d, want=1 or 2",
			len(args))
	}
	inputs, err := interpolateArgs(env, args, false)
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
