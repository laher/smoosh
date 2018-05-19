package stdlib

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	RegisterFn("$", dollar)
	RegisterFn("w", write)
	RegisterFn("r", read)
}

func dollar(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) < 1 {
		return object.NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	inputs := []string{}
	envV := env.Export()
	for _, arg := range args {
		switch argT := arg.(type) {
		case *object.String:
			strings := parseArgv(argT.Value)
			for _, s := range strings {
				input, err := Interpolate(envV, s)
				if err != nil {
					return object.NewError("cannot parse arg for interpolation - %s",
						err)
				}
				inputs = append(inputs, input)
			}
		default:
			return object.NewError("argument to `$` not supported, got %s",
				argT.Type())
		}

	}
	cmd := exec.Command(inputs[0], inputs[1:]...)
	if in != nil {
		cmd.Stdin = in.Out
	}
	if out != nil {
		stdOut, err := cmd.StdoutPipe()
		if err != nil {
			return object.NewError(err.Error())
		}
		errOut, err := cmd.StderrPipe()
		if err != nil {
			return object.NewError(err.Error())
		}
		out.Out = stdOut
		out.Err = errOut
		out.Wait = cmd.Wait
		err = cmd.Start()
		if err != nil {
			return object.NewError(err.Error())
		}
		p := object.Pipes(*out)
		return &p
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return object.NewError(err.Error())
	}
	return Null
}
func write(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError("wrong number of arguments. got=%d, want=1 or 2",
			len(args))
	}
	inputs := []string{}
	envV := env.Export()
	for i, arg := range args {
		if arg.Type() != object.STRING_OBJ {
			return object.NewError("argument to `$` must be STRING, got %s",
				args[i].Type())
		}
		switch argT := arg.(type) {
		case *object.String:
			input, err := Interpolate(envV, argT.Value)
			if err != nil {
				return object.NewError("cannot parse arg for interpolation - %s",
					err)
			}
			inputs = append(inputs, input)
		default:
			return object.NewError("argument to `$` not supported, got %s",
				argT.Type())
		}

	}
	if in == nil {
		return Null
	}

	if inputs[0] != "" {
		f, err := os.Create(inputs[0])
		if err != nil {
			return object.NewError(err.Error())
		}
		if _, err := io.Copy(f, in.Out); err != nil {
			return object.NewError(err.Error())
		}
	}
	if len(inputs) > 1 && inputs[1] != "" && in.Err != nil {
		f, err := os.Create(inputs[1])
		if err != nil {
			return object.NewError(err.Error())
		}
		if _, err := io.Copy(f, in.Err); err != nil {
			return object.NewError(err.Error())
		}
	}

	return Null
}
func read(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return object.NewError("wrong number of arguments. got=%d, want=1 or 2",
			len(args))
	}
	inputs, err := InterpolateArgsAsStrings(env, args)
	if err != nil {
		return object.NewError(err.Error())
	}
	f, err := os.Open(inputs[0])
	if err != nil {
		return object.NewError(err.Error())
	}
	if out != nil {
		out.Out = f
		return Null
	}
	if _, err := io.Copy(os.Stdout, f); err != nil {
		return object.NewError(err.Error())
	}
	return Null
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
