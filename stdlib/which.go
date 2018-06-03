package stdlib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "a", Help: "All"},
	}
	RegisterBuiltin("which", &object.Builtin{
		Fn:    which,
		Flags: opts,
	})
}

func which(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	which := Which{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "a":
				which.all = true
			default:
				return object.NewError("flag %s not supported", arg.Name)
			}

		case *object.String:
			//Filenames (globs):
			d, err := Interpolate(env.Export(), arg.Value)
			if err != nil {
				return object.NewError(err.Error())
			}
			which.args = append(which.args, d)
		default:
			return object.NewError("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}
	stdin := getReader(in)
	stdout, _ := getWriters(out)
	err := which.do(stdout, stdin)
	if err != nil {
		return object.NewError(err.Error())
	}
	return Null
}

// Which represents and performs a `which` invocation
type Which struct {
	all  bool
	args []string
}

// Exec actually performs the which
func (which *Which) do(stdout io.Writer, stdin io.Reader) error {
	path := os.Getenv("PATH")
	if runtime.GOOS == "windows" {
		path = ".;" + path
	}
	pl := filepath.SplitList(path)
	for _, arg := range which.args {
		checkPathParts(arg, pl, which, stdout)
	}
	return nil

}

func checkPathParts(arg string, pathParts []string, which *Which, outPipe io.Writer) {
	for _, pathPart := range pathParts {
		fi, err := os.Stat(pathPart)
		if err == nil {
			if fi.IsDir() {
				possibleExe := filepath.Join(pathPart, arg)
				if runtime.GOOS == "windows" {
					if !strings.HasSuffix(possibleExe, ".exe") {
						possibleExe += ".exe"
					}
				}
				_, err := os.Stat(possibleExe)
				if err != nil {
					//skip
				} else {
					abs, err := filepath.Abs(possibleExe)
					if err == nil {
						fmt.Fprintln(outPipe, abs)
					} else {
						//skip
					}
					if !which.all {
						return
					}
				}
			} else {
				//skip
			}
		} else {
			//skip
		}
	}
}
