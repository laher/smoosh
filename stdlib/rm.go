package stdlib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "r"},
	}
	RegisterBuiltin("rm", &object.Builtin{
		Fn:    rm,
		Flags: opts,
	})
}

func rm(scope object.Scope, args ...object.Object) (object.Operation, error) {
	var (
		allFiles    []string
		err         error
		isRecursive bool
	)
	allFiles, err = interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "r": //follow by name
				isRecursive = true
			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}
		}
	}

	return func() object.Object {
		for _, file := range allFiles {
			err := deleteFile(file, isRecursive)
			if err != nil {
				return object.NewError(err.Error())
			}
		}
		return Null
	}, nil
}

func deleteFile(file string, recursive bool) error {
	fi, e := os.Stat(file)
	if e != nil {
		return e
	}
	if fi.IsDir() && recursive {
		e := deleteDir(file)
		if e != nil {
			return e
		}
	} else if fi.IsDir() {
		//do nothing
		return fmt.Errorf("'%s' is a directory. Use -r", file)
	}
	return os.Remove(file)
}

func deleteDir(dir string) error {
	files, e := ioutil.ReadDir(dir)
	if e != nil {
		return e
	}
	for _, file := range files {
		e = deleteFile(filepath.Join(dir, file.Name()), true)
		if e != nil {
			return e
		}
	}
	return nil
}
