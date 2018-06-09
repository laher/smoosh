package stdlib

import (
	"fmt"
	"io"
	"os"

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

func tee(scope object.Scope, args ...object.Object) (object.Operation, error) {
	tee := &Tee{}
	inputs, err := interpolateArgs(scope.Env, args, false)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	tee.args = inputs
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "a": //follow by name
				tee.isAppend = true
			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}
		}
	}
	return func() object.Object {
		err = tee.do(scope.Env.Streams)
		if err != nil {
			return object.NewError(err.Error())
		}
		return Null
	}, nil
}

func (tee *Tee) do(streams object.Streams) error {
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if tee.isAppend {
		flag = os.O_APPEND | os.O_WRONLY
	}
	closers := []io.WriteCloser{}
	writers := []io.Writer{streams.Stdout}
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
	_, err := io.Copy(multiwriter, streams.Stdin)
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
