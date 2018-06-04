package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "E"},
		object.Flag{Name: "n"},
		object.Flag{Name: "s"},
	}
	RegisterBuiltin("cat", &object.Builtin{
		Fn:    cat,
		Flags: opts,
	})

}

func cat(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	var showEnds, number, squeezeBlank bool
	var fileNames = []string{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "E":
				showEnds = true
			case "n":
				number = true
			case "s":
				squeezeBlank = true
			default:
				return object.NewError("flag %s not supported", arg.Name)
			}
		case *object.String:
			d, err := Interpolate(env.Export(), arg.Value)
			if err != nil {
				return object.NewError(err.Error())
			}
			fileNames = append(fileNames, d)
		default:
			return object.NewError("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}

	stdin := getReader(in)
	stdout, stderr := getWriters(out)
	op := catIt(stdin, stdout, fileNames, showEnds, number, squeezeBlank)
	if out != nil {
		doAsync(op, out, stderr)
	} else {
		err := op()
		if err != nil {
			return object.NewError(err.Error())
		}
	}
	return Null
}

func catIt(stdin io.Reader, stdout io.Writer, fileNames []string, showEnds, number, squeezeBlank bool) op {
	var op op
	if len(fileNames) > 0 {
		op = func() error {
			for _, fileName := range fileNames {
				file, err := os.Open(fileName)
				if err != nil {
					return err
				}
				defer file.Close()
				if !showEnds && !number && !squeezeBlank {
					_, err = io.Copy(stdout, file)
					if err != nil {
						return err
					}
				} else {
					scanner := bufio.NewScanner(file)
					line := 1
					var prefix string
					var suffix string
					for scanner.Scan() {
						text := scanner.Text()
						if !squeezeBlank || len(strings.TrimSpace(text)) > 0 {
							if number {
								prefix = fmt.Sprintf("%d ", line)
							} else {
								prefix = ""
							}
							if showEnds {
								suffix = "$"
							} else {
								suffix = ""
							}
							fmt.Fprintf(stdout, "%s%s%s\n", prefix, text, suffix)
						}
						line++
					}
					err := scanner.Err()
					if err != nil {
						return err
					}
				}
				err = file.Close()
				if err != nil {
					return err
				}
			}
			return nil
		}
	} else {
		op = func() error {
			_, err := io.Copy(stdout, stdin)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return op
}
