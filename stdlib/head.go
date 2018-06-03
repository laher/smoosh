package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/laher/smoosh/ast"
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
func head(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	head := &Head{lines: 10}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "n":
				l, ok := arg.Param.(*object.Integer)
				if !ok {
					return object.NewError("flag %s parse error", arg.Name)
				}
				head.lines = l.Value
			default:
				return object.NewError("flag %s not supported", arg.Name)
			}
		case *object.String:
			//Filenames (globs):
			d, err := Interpolate(env.Export(), arg.Value)
			if err != nil {
				return object.NewError(err.Error())
			}
			head.Filenames = append(head.Filenames, d)
		case *object.Integer:
			// oops
			head.lines = arg.Value
		default:
			return object.NewError("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}

	err := head.do(out, in)
	if err != nil {
		return object.NewError(err.Error())
	}
	return Null
}

func (head *Head) do(out, in *ast.Pipes) error {
	stdout, _ := getWriters(out)
	stdin := getReader(in)
	if len(head.Filenames) > 0 {
		for _, fileName := range head.Filenames {
			file, err := os.Open(fileName)
			if err != nil {
				return err
			}
			defer file.Close()
			err = head.head(stdout, file)
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
		err := head.head(stdout, stdin)
		if err != nil {
			return err
		}
	}
	return nil
}

func (head *Head) head(out io.WriteCloser, in io.ReadCloser) error {
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
