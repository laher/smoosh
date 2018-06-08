package stdlib

import (
	"fmt"
	"strings"
	"sync"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("echo", echo)
}

func echo(scope object.Scope, args ...object.Object) (object.Operation, error) {
	inputs, err := interpolateArgs(scope.Env, args, false)
	if err != nil {
		return nil, err
	}
	if len(inputs) < 1 || len(inputs) > 2 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1 or 2",
			len(inputs))
	}

	wg := sync.WaitGroup{}
	if scope.Out != nil {
		wg.Add(1)
		scope.Out.Wait = func() error {
			wg.Wait()
			return nil
		}
	}
	return func() object.Object {
		o, _ := getWriters(scope.Out)
		w := strings.Join(inputs, " ")
		if scope.Out != nil {
			go func() {
				fmt.Fprintf(o, "%s\n", w)
				o.Close()
				wg.Done()
			}()
		} else {
			fmt.Fprintf(o, "%s\n", w)
		}
		return Null
	}, nil

}
