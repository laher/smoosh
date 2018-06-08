package stdlib

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/template"
	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

var (
	// Null can be a single instance
	Null = &object.Null{}
)

func interpolateArgs(env *object.Environment, args []object.Object, glob bool) ([]string, error) {
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
			if glob {
				ss, err := filepath.Glob(input)
				if err != nil {
					return nil, err
				}
				inputs = append(inputs, ss...)
			} else {
				inputs = append(inputs, input)
			}
		case *object.Integer:
			input := fmt.Sprintf("%d", argT.Value)
			inputs = append(inputs, input)
		case *object.Null:
			// ignore nulls
		case *object.Flag:
			// ignore flags here. Parse them separately
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
		out.Out = r // this will be closed by the evaluator

		r, w = io.Pipe()
		stderr = w
		out.Err = r // this will be closed by the evaluator
	}
	return stdout, stderr
}
