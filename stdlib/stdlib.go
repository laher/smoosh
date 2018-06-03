package stdlib

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/alecthomas/template"
	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

var (
	Null = &object.Null{}
)

func init() {
	RegisterFn("cd", cd)
	RegisterFn("pwd", pwd)
	RegisterFn("exit", exit)
	RegisterFn("echo", echo)
}

func cd(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) != 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	if args[0].Type() != object.STRING_OBJ {
		return object.NewError("argument to `cd` must be STRING, got %s",
			args[0].Type())
	}
	switch arg := args[0].(type) {
	case *object.String:
		d, err := Interpolate(env.Export(), arg.Value)
		if err != nil {
			return object.NewError(err.Error())
		}
		err = os.Chdir(d)
		if err != nil {
			return object.NewError(err.Error())
		}
		return Null
	default:
		return object.NewError("argument to `cd` not supported, got %s",
			args[0].Type())
	}

}

func pwd(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) != 0 {
		return object.NewError("wrong number of arguments. got=%d, want=0",
			len(args))
	}
	d, err := os.Getwd()
	if err != nil {
		return object.NewError(err.Error())
	}
	//fmt.Println(d) //TODO make the repl print this somehow instead
	return &object.String{Value: d}
}
func exit(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) > 1 {
		return object.NewError("wrong number of arguments. got=%d, want=0/1",
			len(args))
	}
	if len(args) == 1 {
		if args[0].Type() != object.INTEGER_OBJ {
			return object.NewError("argument to `exit` must be INTEGER, got %s",
				args[0].Type())
		}
		switch arg := args[0].(type) {
		case *object.Integer:
			os.Exit(int(arg.Value))
		default:
			return object.NewError("argument to `exit` not supported, got %s",
				args[0].Type())
		}
	}
	os.Exit(0)
	return Null

}

func echo(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError("wrong number of arguments. got=%d, want=1 or 2",
			len(args))
	}
	inputs, err := InterpolateArgsAsStrings(env, args)
	if err != nil {
		return object.NewError(err.Error())
	}
	o, _ := getWriters(out)
	for i, w := range inputs {
		s := " "
		if len(inputs) == i+1 {
			s = "\n"
		}

		fmt.Fprintf(o, "%s%s", w, s)
	}
	return Null
}

func InterpolateArgsAsStrings(env *object.Environment, args []object.Object) ([]string, error) {
	inputs := []string{}
	envV := env.Export()
	for i, arg := range args {
		switch argT := arg.(type) {
		case *object.String:
			input, err := Interpolate(envV, argT.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse arg for interpolation - %s",
					err)
			}
			inputs = append(inputs, input)
		case *object.Integer:
			input := fmt.Sprintf("%d", argT.Value)
			inputs = append(inputs, input)
		case *object.Null:
			inputs = append(inputs, "<NULL>")

		default:
			return nil, fmt.Errorf("argument %d not supported, got %s",
				i, argT.Type())
		}
	}
	return inputs, nil
}

// Interpolate replaces strings using a template
func Interpolate(envV map[string]interface{}, value string) (string, error) {
	tmpl, err := template.New("test").Parse(value)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buf, envV)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getReader(in *ast.Pipes) io.ReadCloser {
	if in != nil {
		return in.Out
	}
	return nil
}

func getWriters(out *ast.Pipes) (io.WriteCloser, io.WriteCloser) {
	var (
		stdout io.WriteCloser = os.Stdout
		stderr io.WriteCloser = os.Stderr
	)
	if out != nil {
		r, w := io.Pipe()
		stdout = w
		out.Out = ioutil.NopCloser(r) // this will be closed by the evaluator

		r, w = io.Pipe()
		stderr = w
		out.Err = ioutil.NopCloser(r) // this will be closed by the evaluator
	}
	return stdout, stderr
}
