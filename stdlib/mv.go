package stdlib

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("mv", &object.Builtin{
		Fn: mv,
	})

}

func mv(scope object.Scope, args ...object.Object) (object.Operation, error) {
	var (
		srcs []string
		dest string
		err  error
	)
	srcs, err = interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			return nil, fmt.Errorf("flag %s not supported", arg.Name)
		}
	}
	return func() object.Object {
		for _, src := range srcs {
			err := moveFile(src, dest)
			if err != nil {
				return object.NewError(err.Error())
			}
		}
		return Null
	}, nil
}

func moveFile(src, dest string) error {
	//wd, err := os.Getwd()
	//fmt.Printf("%s: %s -> %s\n", wd, src, dest)
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	sinf, err := srcFile.Stat()
	if err != nil {
		return err
	}
	err = srcFile.Close()
	if err != nil {
		return err
	}

	//check if destination given is full filename or its (existing) parent dir
	var destFull string
	dinf, err := os.Stat(dest)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		//doesnt exist
		destFull = dest
	} else {
		if dinf.IsDir() {
			//copy file name
			destFull = filepath.Join(dest, sinf.Name())
		} else {
			destFull = dest
		}
	}
	err = os.Rename(src, destFull)
	return err
}
