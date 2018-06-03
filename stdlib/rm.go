package stdlib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/laher/smoosh/ast"
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

// Rm represents and performs a `rm` invocation
type Rm struct {
	IsRecursive bool
	fileGlobs   []string
}

func rm(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	rm := Rm{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "r": //follow by name
				rm.IsRecursive = true
			default:
				return object.NewError("flag %s not supported", arg.Name)
			}
		case *object.String:
			//Filenames (globs):
			d, err := Interpolate(env.Export(), arg.Value)
			if err != nil {
				return object.NewError(err.Error())
			}
			rm.fileGlobs = append(rm.fileGlobs, d)
		default:
			return object.NewError("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}

	for _, fileGlob := range rm.fileGlobs {
		files, err := filepath.Glob(fileGlob)
		if err != nil {
			return object.NewError(err.Error())
		}
		for _, file := range files {
			err := deleteFile(file, rm.IsRecursive)
			if err != nil {
				return object.NewError(err.Error())
			}
		}
	}

	return Null
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
