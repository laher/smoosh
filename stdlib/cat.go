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
	stdout, _ := getWriters(out)
	catIt(stdout, stdin, fileNames, showEnds, number, squeezeBlank)
	return Null
}

func catIt(stdout io.Writer, in io.Reader, fileNames []string, showEnds, number, squeezeBlank bool) error {
	if len(fileNames) > 0 {
		for _, fileName := range fileNames {
			file, err := os.Open(fileName)
			if err != nil {
				return err
			}
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
			return file.Close()
		}
	} else {
		_, err := io.Copy(stdout, in)
		if err != nil {
			return err
		}
	}

	return nil
}
