package stdlib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("http.Get", get)
}
func get(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	switch arg := args[0].(type) {
	case *object.String:
		a, err := Interpolate(env.Export(), arg.Value)
		if err != nil {
			return object.NewError(err.Error())
		}
		resp, err := http.Get(a)
		if err != nil {
			return object.NewError(err.Error())
		}
		if out != nil {
			out.Out = resp.Body
			errStr := resp.Status + "\n\n"
			for k, v := range resp.Header {
				for _, h := range v {
					errStr += fmt.Sprintf("%s: %s\n", k, h)
				}
			}
			out.Err = ioutil.NopCloser(bytes.NewBufferString(errStr))
			p := object.Pipes(*out)
			return &p
		}
		rb, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return object.NewError(err.Error())
		}
		return &object.String{Value: string(rb)}
	default:
		return object.NewError("argument to `len` not supported, got %s",
			args[0].Type())
	}
}
