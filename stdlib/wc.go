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
		object.Flag{Name: "l", Help: "Count lines"},
		object.Flag{Name: "w", Help: "Count words"},
		object.Flag{Name: "c", Help: "Count bytes"},
	}
	RegisterBuiltin("wc", &object.Builtin{
		Fn:    wc,
		Flags: opts,
	})
}

// Wc represents and performs a `wc` invocation
type Wc struct {
	IsBytes bool
	IsWords bool
	IsLines bool
	args    []string
}

func wc(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	wc := Wc{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "b":
				wc.IsBytes = true
			case "w":
				wc.IsWords = true
			case "l":
				wc.IsLines = true
			default:
				return object.NewError("flag %s not supported", arg.Name)
			}

		case *object.String:
			//Filenames (globs):
			d, err := Interpolate(env.Export(), arg.Value)
			if err != nil {
				return object.NewError(err.Error())
			}
			wc.args = append(wc.args, d)
		default:
			return object.NewError("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}
	stdin := getReader(in)
	stdout, _ := getWriters(out)
	err := wc.do(stdout, stdin)
	if err != nil {
		return object.NewError(err.Error())
	}
	return Null
}

// Invoke actually performs the wc
func (wc *Wc) do(stdout io.Writer, stdin io.Reader) error {
	if len(wc.args) > 0 {
		//treat no args as all args
		if !wc.IsWords && !wc.IsLines && !wc.IsBytes {
			wc.IsWords = true
			wc.IsLines = true
			wc.IsBytes = true
		}
		for _, fileName := range wc.args {
			bytes := int64(0)
			words := int64(0)
			lines := int64(0)
			//get byte count
			file, err := os.Open(fileName)
			if err != nil {
				return err
			}
			err = countWords(file, wc, &bytes, &words, &lines)
			if err != nil {
				file.Close()
				return err
			}
			err = file.Close()
			if err != nil {
				return err
			}
			if wc.IsWords && !wc.IsLines && !wc.IsBytes {
				fmt.Fprintf(stdout, "%d %s\n", words, fileName)
			} else if !wc.IsWords && wc.IsLines && !wc.IsBytes {
				fmt.Fprintf(stdout, "%d %s\n", lines, fileName)
			} else if !wc.IsWords && !wc.IsLines && wc.IsBytes {
				fmt.Fprintf(stdout, "%d %s\n", bytes, fileName)
			} else {
				fmt.Fprintf(stdout, "%d %d %d %s\n", lines, words, bytes, fileName)
			}
		}
	} else {
		//stdin ..
		if !wc.IsWords && !wc.IsLines && !wc.IsBytes {
			wc.IsWords = true
		}
		bytes := int64(0)
		words := int64(0)
		lines := int64(0)
		err := countWords(stdin, wc, &bytes, &words, &lines)
		if err != nil {
			return err
		}
		if wc.IsWords && !wc.IsLines && !wc.IsBytes {
			fmt.Fprintf(stdout, "%d\n", words)
		} else if !wc.IsWords && wc.IsLines && !wc.IsBytes {
			fmt.Fprintf(stdout, "%d\n", lines)
		} else if !wc.IsWords && !wc.IsLines && wc.IsBytes {
			fmt.Fprintf(stdout, "%d\n", bytes)
		} else {
			fmt.Fprintf(stdout, "%d %d %d\n", lines, words, bytes)
		}
	}
	return nil

}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func countWords(file io.Reader, wc *Wc, bytes *int64, words *int64, lines *int64) (err error) {
	lastWasSpace := false
	bio := bufio.NewReader(file)
	for err == nil {
		c, err := bio.ReadByte()
		if err != nil {
			if io.EOF == err {
				return nil
			}
			return err
		}
		*bytes++
		if isSpace(c) {
			if !lastWasSpace {
				*words++
			}
			lastWasSpace = true
		} else {
			lastWasSpace = false
		}
		if c == '\n' {
			*lines++
		}

	}
	return err
}
