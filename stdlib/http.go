package stdlib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("http.Get", get)
}
func get(scope object.Scope, args ...object.Object) (object.Operation, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.String:
		a, err := Interpolate(scope.Env.Export(), arg.Value)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		return func() object.Object {
			resp, err := http.Get(a)
			if err != nil {
				return object.NewError(err.Error())
			}
			if scope.Out != nil {
				scope.Out.Main = resp.Body
				errStr := resp.Status + "\n\n"
				for k, v := range resp.Header {
					for _, h := range v {
						errStr += fmt.Sprintf("%s: %s\n", k, h)
					}
				}
				scope.Out.Err = ioutil.NopCloser(bytes.NewBufferString(errStr))
				p := object.Pipes(*scope.Out)
				return &p
			}
			rb, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return object.NewError(err.Error())
			}
			return &object.String{Value: string(rb)}
		}, nil
	default:
		return nil, fmt.Errorf("argument to `len` not supported, got %s",
			args[0].Type())
	}
}
