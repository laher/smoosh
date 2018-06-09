package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "n", ParamType: object.INTEGER_OBJ},
	}
	RegisterBuiltin("head", &object.Builtin{
		Fn:    head,
		Flags: opts,
	})
}

// Head represents and performs a `head` invocation
type Head struct {
	lines     int64
	ch        byte
	Filenames []string
}

// Name() returns the name of the util
func (head *Head) Name() string {
	return "head"
}

// Exec actually performs the head
func head(scope object.Scope, args ...object.Object) (object.Operation, error) {
	head := &Head{lines: 10}
	var err error
	head.Filenames, err = interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "n":
				l, ok := arg.Param.(*object.Integer)
				if !ok {
					return nil, fmt.Errorf("flag %s parse error", arg.Name)
				}
				head.lines = l.Value
			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}
		}
	}

	return func() object.Object {
		err := head.do(scope.Env.Streams)
		if err != nil {
			return object.NewError(err.Error())
		}
		return Null
	}, nil
}

func (head *Head) do(streams object.Streams) error {
	if len(head.Filenames) > 0 {
		for _, fileName := range head.Filenames {
			file, err := os.Open(fileName)
			if err != nil {
				return err
			}
			defer file.Close()
			err = head.head(streams.Stdout, file)
			if err != nil {
				return err
			}
			err = file.Close()
			if err != nil {
				return err
			}
		}
	} else {
		//stdin ..
		err := head.head(streams.Stdout, streams.Stdin)
		if err != nil {
			return err
		}
	}
	return nil
}

func (head *Head) head(out io.Writer, in io.Reader) error {
	reader := bufio.NewReader(in)
	lineNo := int64(1)
	ch := '\n' //should this be an option?
	for lineNo <= head.lines {
		text, err := reader.ReadBytes(byte(ch))
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		//text := scanner.Text()
		fmt.Fprintf(out, "%s", text) //, string(ch))
		lineNo++
	}
	return nil
}
