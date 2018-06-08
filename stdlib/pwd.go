package stdlib

import (
	"fmt"
	"os"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("pwd", pwd)
}

func pwd(scope object.Scope, args ...object.Object) (object.Operation, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=0",
			len(args))
	}
	return func() object.Object {
		d, err := os.Getwd()
		if err != nil {
			return object.NewError(err.Error())
		}
		//fmt.Println(d) //TODO make the repl print this somehow instead
		return &object.String{Value: d}
	}, nil
}
