package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
		object.Flag{Name: "r"},
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
	paths   []string
}

func grep(scope object.Scope, args ...object.Object) (object.Operation, error) {

	grep := &Grep{}
	var err error
	myArgs, err := interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}
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
			case "r":
				grep.IsRecurse = true
			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}
		}
	}
	if len(myArgs) < 1 {
		return nil, fmt.Errorf("Missing operand")
	}
	grep.pattern = myArgs[0]
	if len(myArgs) > 1 {
		grep.paths = myArgs[1:]
	}
	if len(grep.paths) > 1 {
		grep.IsPrintFilename = true
	}

	reg, err := compile(grep)
	if err != nil {
		return nil, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return func() object.Object {
		if len(grep.paths) > 0 {
			err = grepAll(reg, cwd, grep.paths, grep, scope.Env.Streams.Stdout)
			if err != nil {
				return object.NewError(err.Error())
			}
		} else {
			if scope.In != nil {
				err = grepReader(scope.In.Main, cwd, "", reg, grep, scope.Env.Streams.Stdout)
				if err != nil {
					return object.NewError(err.Error())
				}
			} else {
				//NOT piping.
				return object.NewError("Not enough args")
			}
		}
		return Null
	}, nil
}

func grepAll(reg *regexp.Regexp, cwd string, files []string, grep *Grep, out io.Writer) error {
	for _, filename := range files {
		fi, err := os.Stat(filename)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			//recurse here
			if grep.IsRecurse {
				entries, err := ioutil.ReadDir(filename)
				if err != nil {
					return err
				}
				fs := []string{}
				for _, e := range entries {
					fs = append(fs, filepath.Join(filename, e.Name()))
				}
				err = grepAll(reg, filename, fs, grep, out)
				if err != nil {
					return err
				}
				continue
			}
			continue
		}
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		err = grepReader(file, cwd, filename, reg, grep, out)
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

func grepReader(file io.Reader, cwd, filename string, reg *regexp.Regexp, grep *Grep, out io.Writer) error {
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
