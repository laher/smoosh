package stdlib

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/laher/smoosh/object"
)

var (
	// Null can be a single instance
	Null = &object.Null{}
)

func interpolateArgs(env *object.Environment, args []object.Object, glob bool) ([]string, error) {
	inputs := []string{}
	envV := env.Export()
	for i := range args {
		switch arg := args[i].(type) {
		case *object.String:
			input, err := Interpolate(envV, arg.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse arg for interpolation - %s",
					err)
			}
			if glob {
				ss, err := filepath.Glob(input)
				if err != nil {
					return nil, err
				}
				if len(ss) == 0 {
					inputs = append(inputs, input)
					break
				}
				inputs = append(inputs, ss...)
			} else {
				inputs = append(inputs, input)
			}
		case *object.Integer:
			input := fmt.Sprintf("%d", arg.Value)
			inputs = append(inputs, input)
		case *object.Null:
			// ignore nulls
		case *object.Flag:
			// ignore flags here. Parse them separately
		default:
			return nil, fmt.Errorf("argument %d not supported, got %s",
				i, arg.Type())
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
