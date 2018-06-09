package stdlib

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

func which(scope object.Scope, args ...object.Object) (object.Operation, error) {
	which := Which{}
	var err error
	which.args, err = interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "a":
				which.all = true
			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}
		}
	}
	return func() object.Object {
		err := which.do(scope.Env.Streams.Stdout, scope.Env.Streams.Stdin)
		if err != nil {
			return object.NewError(err.Error())
		}
		return Null
	}, nil
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
