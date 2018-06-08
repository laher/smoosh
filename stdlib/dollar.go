package stdlib

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("$", dollar)
}

func dollar(scope object.Scope, args ...object.Object) (object.Operation, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	inputs := []string{}
	envV := scope.Env.Export()
	for i := range args {
		switch arg := args[i].(type) {
		case *object.String:
			strings := parseArgv(arg.Value)
			for _, s := range strings {
				input, err := Interpolate(envV, s)
				if err != nil {
					return nil, fmt.Errorf("cannot parse arg for interpolation - %s",
						err)
				}
				inputs = append(inputs, input)
			}
		default:
			return nil, fmt.Errorf("argument to `$` not supported, got %s",
				args[i].Type())
		}
	}
	cmd := exec.Command(inputs[0], inputs[1:]...)
	return func() object.Object {
		if scope.In != nil {
			cmd.Stdin = scope.In.Out
		}
		if scope.Out != nil {
			stdOut, err := cmd.StdoutPipe()
			if err != nil {
				return object.NewError(err.Error())
			}
			errOut, err := cmd.StderrPipe()
			if err != nil {
				return object.NewError(err.Error())
			}
			scope.Out.Out = stdOut
			scope.Out.Err = errOut
			scope.Out.Wait = cmd.Wait
			err = cmd.Start()
			if err != nil {
				return object.NewError(err.Error())
			}
			p := object.Pipes(*scope.Out)
			return &p
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return object.NewError(err.Error())
		}
		return Null
	}, nil
}

func parseArgv(p string) []string {
	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)

		}
	}
	m := strings.FieldsFunc(p, f)
	return m
}
