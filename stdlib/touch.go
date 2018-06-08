package stdlib

import (
	"fmt"
	"os"
	"time"

	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "a"},
	}
	RegisterBuiltin("touch", &object.Builtin{
		Fn:    touch,
		Flags: opts,
	})
}

// Touch represents and performs a `touch` invocation
type Touch struct {
	args []string
}

func touch(scope object.Scope, args ...object.Object) (object.Operation, error) {
	touch := &Touch{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.String:
			//Filenames (globs):
			d, err := Interpolate(scope.Env.Export(), arg.Value)
			if err != nil {
				return nil, fmt.Errorf(err.Error())
			}
			touch.args = append(touch.args, d)
		default:
			return nil, fmt.Errorf("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}
	return func() object.Object {
		for _, f := range touch.args {
			err := touchFile(f)
			if err != nil {
				return object.NewError(err.Error())
			}
		}
		return Null
	}, nil
}

func touchFile(filename string) error {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.Create(filename)
			if err != nil {
				return err
			}
			return file.Close()
		}
		return err
	}
	//set access times
	os.Chtimes(filename, time.Now(), time.Now())
	return nil
}
