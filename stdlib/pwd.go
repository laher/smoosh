package stdlib

import (
	"os"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("pwd", pwd)
}

func pwd(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) != 0 {
		return object.NewError("wrong number of arguments. got=%d, want=0",
			len(args))
	}
	d, err := os.Getwd()
	if err != nil {
		return object.NewError(err.Error())
	}
	//fmt.Println(d) //TODO make the repl print this somehow instead
	return &object.String{Value: d}
}
