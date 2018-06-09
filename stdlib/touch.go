package stdlib

import (
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

func touch(scope object.Scope, args ...object.Object) (object.Operation, error) {
	files, err := interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}
	return func() object.Object {
		for _, f := range files {
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
