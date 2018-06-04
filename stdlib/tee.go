package stdlib

import (
	"io"
	"os"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "a"},
	}
	RegisterBuiltin("tee", &object.Builtin{
		Fn:    tee,
		Flags: opts,
	})
}

// Tee represents and performs a `tee` invocation
type Tee struct {
	isAppend bool
	flag     int
	args     []string
}

func tee(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	tee := &Tee{}
	inputs, err := interpolateArgs(env, args, false)
	if err != nil {
		return object.NewError(err.Error())
	}
	tee.args = inputs
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "a": //follow by name
				tee.isAppend = true
			default:
				return object.NewError("flag %s not supported", arg.Name)
			}
		}
	}
	stdout, _ := getWriters(out)
	stdin := getReader(in)
	err = tee.do(stdout, stdin)
	if err != nil {
		return object.NewError(err.Error())
	}
	return Null
}

func (tee *Tee) do(stdout io.WriteCloser, stdin io.ReadCloser) error {
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if tee.isAppend {
		flag = os.O_APPEND | os.O_WRONLY
	}
	closers := []io.WriteCloser{}
	writers := []io.Writer{stdout}
	for _, file := range tee.args {
		f, err := os.OpenFile(file, flag, 0666)
		if err != nil {
			return err
		}
		defer f.Close()
		writers = append(writers, f)
		closers = append(closers, f)
	}
	multiwriter := io.MultiWriter(writers...)
	_, err := io.Copy(multiwriter, stdin)
	if err != nil {
		return err
	}
	for _, file := range closers {
		err = file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
