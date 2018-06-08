package stdlib

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("cp", &object.Builtin{
		Fn: cp,
		Flags: []object.Flag{
			object.Flag{Name: "r"},
		},
	})

}

func cp(scope object.Scope, args ...object.Object) (object.Operation, error) {
	var (
		srces     []string
		dest      string
		recursive bool
	)
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "r":
				recursive = true
			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}
		case *object.String:
			d, err := Interpolate(scope.Env.Export(), arg.Value)
			if err != nil {
				return nil, fmt.Errorf(err.Error())
			}
			if i+1 < len(args) {
				ss, err := filepath.Glob(d)
				if err != nil {
					return nil, fmt.Errorf(err.Error())
				}
				srces = append(srces, ss...)
			} else {
				dest = d
			}
		default:
			return nil, fmt.Errorf("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}
	return func() object.Object {
		for _, src := range srces {
			err := copyFile(src, dest, recursive)
			if err != nil {
				return object.NewError(err.Error())
			}
		}
		return Null
	}, nil
}

func copyFile(src, dest string, recursive bool) error {
	//println("copy "+src+" to "+dest)

	srcFile, err := os.Open(src)
	defer srcFile.Close()
	if err != nil {
		return err
	}
	sinf, err := srcFile.Stat()
	if err != nil {
		return err
	}
	if sinf.IsDir() && !recursive {
		return errors.New("Omitting directory " + src)
	}

	//check if destination given is full filename or its (existing) parent dir
	var destFull string
	dinf, err := os.Stat(dest)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		//doesnt exist yet. New file/dir
		destFull = dest
	} else {
		if dinf.IsDir() {
			//copy file name
			destFull = filepath.Join(dest, sinf.Name())
		} else {
			destFull = dest
		}
	}
	//println("copy "+src+" to "+destFull)

	var destExists bool
	dinf, err = os.Stat(destFull)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		//doesnt exist. New file/dir
		destExists = false
	} else {
		destExists = true
		if sinf.IsDir() && !dinf.IsDir() {
			return errors.New("destination is an existing non-directory")
		}
	}

	if sinf.IsDir() {
		//println("copying dir")
		if !destExists {
			//println("mkdir")
			err = os.Mkdir(destFull, sinf.Mode())
			if err != nil {
				return err
			}
		} else {
			//continue
		}
		contents, err := srcFile.Readdir(0)
		if err != nil {
			return err
		}
		err = srcFile.Close()
		if err != nil {
			return err
		}
		for _, fi := range contents {
			copyFile(filepath.Join(src, fi.Name()), destFull, recursive)
		}
	} else {
		flags := os.O_WRONLY
		if !destExists {
			flags = flags + os.O_CREATE
		} else {
			flags = flags + os.O_TRUNC
		}
		destFile, err := os.OpenFile(destFull, flags, sinf.Mode())
		defer destFile.Close()
		if err != nil {
			return err
		}
		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return err
		}
		err = destFile.Close()
		if err != nil {
			return err
		}
		err = srcFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
