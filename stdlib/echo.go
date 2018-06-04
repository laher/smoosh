package stdlib

import (
	"fmt"
	"strings"
	"sync"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("echo", echo)
}

func echo(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError("wrong number of arguments. got=%d, want=1 or 2",
			len(args))
	}
	inputs, err := interpolateArgs(env, args, false)
	if err != nil {
		return object.NewError(err.Error())
	}
	o, _ := getWriters(out)
	wg := sync.WaitGroup{}
	if out != nil {
		out.Wait = func() error {
			wg.Wait()
			return nil
		}
	}
	w := strings.Join(inputs, " ")
	if out != nil {
		wg.Add(1)
		go func() {
			fmt.Fprintf(o, "%s\n", w)
			o.Close()
			wg.Done()
		}()
	} else {
		fmt.Fprintf(o, "%s\n", w)
	}

	return Null
}
