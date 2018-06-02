package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "P"},
		object.Flag{Name: "E"},
		object.Flag{Name: "i"},
		object.Flag{Name: "H"},
		object.Flag{Name: "n"},
		object.Flag{Name: "v"},
	}
	RegisterBuiltin("grep", &object.Builtin{
		Fn:    grep,
		Flags: opts,
	})
}

// Grep represents and performs a `grep` invocation
type Grep struct {
	IsPerl            bool
	IsExtended        bool // TODO extended is true by default
	IsIgnoreCase      bool
	IsInvertMatch     bool
	IsPrintFilename   bool // TODO filename is true by default
	IsPrintLineNumber bool
	IsRecurse         bool
	IsQuiet           bool // TODO
	LinesBefore       int  // TODO
	LinesAfter        int  // TODO
	LinesAround       int  // TODO

	pattern string
	globs   []string
}

func grep(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {

	grep := &Grep{}
	myArgs := []string{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "P":
				grep.IsPerl = true
			case "i":
				grep.IsIgnoreCase = true
			case "H":
				grep.IsPrintFilename = true
			case "n":
				grep.IsPrintLineNumber = true
			case "v":
				grep.IsInvertMatch = true
			case "E":
				grep.IsExtended = true
			default:
				return object.NewError("flag %s not supported", arg.Name)
			}
		case *object.String:
			d, err := Interpolate(env.Export(), arg.Value)
			if err != nil {
				return object.NewError(err.Error())
			}
			myArgs = append(myArgs, d)
		default:
			return object.NewError("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}
	if len(myArgs) < 1 {
		return object.NewError("Missing operand")
	}
	if len(myArgs) > 1 {
		grep.globs = myArgs[1:]
	}
	grep.pattern = myArgs[0]
	reg, err := compile(grep)
	if err != nil {
		return object.NewError(err.Error())
	}
	stdout, _ := getWriters(out)
	if len(grep.globs) > 0 {
		files := []string{}
		for _, glob := range grep.globs {
			results, err := filepath.Glob(glob)
			if err != nil {
				return object.NewError(err.Error())
			}
			if len(results) < 1 { //no match
				return object.NewError("grep: cannot access %s: No such file or directory", glob)
			}
			files = append(files, results...)
		}
		err = grepAll(reg, files, grep, stdout)
		if err != nil {
			return object.NewError(err.Error())
		}
	} else {
		if in != nil {
			err = grepReader(in.Out, "", reg, grep, stdout)
			if err != nil {
				return object.NewError(err.Error())
			}
		} else {
			//NOT piping.
			return object.NewError("Not enough args")
		}
	}
	return Null
}

func grepAll(reg *regexp.Regexp, files []string, grep *Grep, out io.Writer) error {
	for _, filename := range files {
		fi, err := os.Stat(filename)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			//recurse here
			if grep.IsRecurse {
				//
				fmt.Fprintf(out, "Recursion not implemented yet\n")
			}
		}
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		err = grepReader(file, filename, reg, grep, out)
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func grepReader(file io.Reader, filename string, reg *regexp.Regexp, grep *Grep, out io.Writer) error {
	scanner := bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			return err
		}
		line := scanner.Text()
		candidate := line
		if grep.IsIgnoreCase && !grep.IsPerl {
			candidate = strings.ToLower(line)
		}
		isMatch := reg.MatchString(candidate)
		if (isMatch && !grep.IsInvertMatch) || (!isMatch && grep.IsInvertMatch) {
			if grep.IsPrintFilename && filename != "" {
				fmt.Fprintf(out, "%s:", filename)
			}
			if grep.IsPrintLineNumber {
				fmt.Fprintf(out, "%d:", lineNumber)
			}
			fmt.Fprintln(out, line)
		}
		lineNumber++
	}
	return nil
}

func compile(grep *Grep) (*regexp.Regexp, error) {
	if grep.IsPerl {
		if grep.IsIgnoreCase && !strings.HasPrefix(grep.pattern, "(?") {
			grep.pattern = "(?i)" + grep.pattern
		}
		return regexp.Compile(grep.pattern)
	}
	if grep.IsIgnoreCase {
		grep.pattern = strings.ToLower(grep.pattern)
	}
	return regexp.CompilePOSIX(grep.pattern)
}
