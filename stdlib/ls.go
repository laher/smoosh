package stdlib

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "l", Help: "Use a long listing format"},
		object.Flag{Name: "R", Help: "List subdirectories recursively"},
		object.Flag{Name: "a", Help: "All files (do not ignore entries starting with .)"},
		object.Flag{Name: "h", Help: "Print human readable sizes (e.g., 1K 234M 2G)"},
	}

	RegisterBuiltin("ls", &object.Builtin{
		Fn:    ls,
		Flags: opts,
		Help: `List files
Usage: ls [OPTION]... [FILE]...                        
List information about the FILEs (the current directory by default).   
Sort entries alphabetically.
`,
	})

}

// Ls represents and performs a `ls` invocation
type Ls struct {
	LongList   bool
	Recursive  bool
	Human      bool
	AllFiles   bool
	OnePerLine bool

	Stdin bool

	Filenames []string

	counter int
}

/*
func (bi bi) builtinW(scope object.Scope, args ...object.Object) (object.Operation, error) {
	op, err := bi.builtinO(env, in, out, args...)
	if err != nil {
		return object.NewError(err.Error())
	}

	if out != nil {
		doAsync(op, out, nil)
	} else {
		err := op()
		if err != nil {
			return object.NewError(err.Error())
		}
	}
	return Null
}*/

func ls(scope object.Scope, args ...object.Object) (object.Operation, error) {
	//object.Object {
	ls := &Ls{}
	if scope.In != nil {
		ls.Stdin = true
	}
	var err error
	ls.Filenames, err = interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}

	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "l":
				ls.LongList = true
			case "a":
				ls.AllFiles = true
			case "h":
				ls.Human = true
			case "1":
				ls.OnePerLine = true
			case "r":
				ls.Recursive = true
			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}
		}
	}

	return func() object.Object {
		err := ls.Go(scope.Env.Streams)
		if err != nil {
			return object.NewError(err.Error())
		}
		return Null
	}, nil
}

// Go actually runs the ls ...
func (ls *Ls) Go(streams object.Streams) error {
	tout := tabwriter.NewWriter(streams.Stdout, 4, 4, 1, ' ', 0)

	args, err := getDirList(ls, streams.Stdin)
	if err != nil {
		return err
	}

	ls.counter = 0
	lastWasDir := false
	endswithNewline := false
	for i, arg := range args {
		if !strings.HasPrefix(arg, ".") || ls.AllFiles ||
			strings.HasPrefix(arg, "..") || "." == arg {
			argInfo, err := os.Stat(arg)
			if err != nil {
				fmt.Fprintln(streams.Stderr, "stat failed for ", arg)
				return err
			}
			if argInfo.IsDir() {
				if len(args) > 1 { //if more than one, print dir name before contents

					if i > 0 {
						fmt.Fprintf(tout, "\n")
					}
					if !lastWasDir {
						fmt.Fprintf(tout, "\n")
					}
					fmt.Fprintf(tout, "%s:\n", arg)
				}
				dir := arg

				//show . and ..
				if ls.AllFiles {
					df, err := os.Stat(filepath.Dir(dir))
					if err != nil {
						fmt.Fprintf(tout, "Error opening parent dir: %v", err)
					} else {
						printEntry("..", df, tout, ls)
					}
					df, err = os.Stat(dir)
					if err != nil {
						fmt.Fprintf(tout, "Error opening dir: %v", err)
					} else {
						printEntry(".", df, tout, ls)
					}
				}

				endswithNewline, err = list(tout, streams.Stderr, dir, "", ls)
				if err != nil {
					return err
				}
				if len(args) > 1 {
					fmt.Fprintf(tout, "\n")
					endswithNewline = true
				}
			} else {
				endswithNewline, err = listItem(argInfo, tout, streams.Stderr, filepath.Dir(arg), "", ls)
				if err != nil {
					return err
				}
			}
			lastWasDir = argInfo.IsDir()
		}
	}
	if !endswithNewline {
		fmt.Fprintf(tout, "\n")
	}
	tout.Flush()
	return nil
}

func list(out *tabwriter.Writer, errPipe io.Writer, dir, prefix string, ls *Ls) (bool, error) {
	endswithNewline := false
	if !strings.HasPrefix(dir, ".") || ls.AllFiles ||
		strings.HasPrefix(dir, "..") || "." == dir {

		entries, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Fprintf(errPipe, "Error reading dir '%s'", dir)
			return endswithNewline, err
		}
		//dirs first, then files
		for _, entry := range entries {
			if entry.IsDir() {
				endswithNewline, err = listItem(entry, out, errPipe, dir, prefix, ls)
				if err != nil {
					return endswithNewline, err
				}
			}
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				endswithNewline, err = listItem(entry, out, errPipe, dir, prefix, ls)
				if err != nil {
					return endswithNewline, err
				}
			}
		}
	}
	return endswithNewline, nil
}

func listItem(entry os.FileInfo, out *tabwriter.Writer, errPipe io.Writer, dir, prefix string, ls *Ls) (bool, error) {
	endswithNewline := false
	if !strings.HasPrefix(entry.Name(), ".") || ls.AllFiles {
		printEntry(entry.Name(), entry, out, ls)
		if entry.IsDir() && ls.Recursive {
			folder := filepath.Join(prefix, entry.Name())
			if ls.counter%3 == 2 || ls.LongList || ls.OnePerLine {
				endswithNewline = true
				fmt.Fprintf(out, "%s:\n", folder)
			} else {
				fmt.Fprintf(out, "%s:\t", folder)
			}
			ls.counter++
			var err error
			endswithNewline, err = list(out, errPipe, filepath.Join(dir, entry.Name()), folder, ls)
			if err != nil {
				return endswithNewline, err
			}
		}
	}
	return endswithNewline, nil
}

func printEntry(name string, e os.FileInfo, out *tabwriter.Writer, ls *Ls) {
	if ls.LongList {
		fmt.Fprintf(out, "%s\t", getModeString(e))
		if !e.IsDir() {
			fmt.Fprintf(out, "%s\t", getSizeString(e.Size(), ls.Human))
		} else {
			fmt.Fprintf(out, "\t")
		}
		fmt.Fprintf(out, "%s\t", getModTimeString(e))
		//disabling due to native-only support
		//fmt.Fprintf(out, "%s\t", getUserString(e.Sys.(*syscall.Stat_t).Uid))
	}
	fmt.Fprintf(out, "%s%s\t", name, getEntryTypeString(e))
	if ls.counter%3 == 2 || ls.LongList || ls.OnePerLine {
		fmt.Fprintln(out, "")
	}
	ls.counter++
}

func getModTimeString(e os.FileInfo) (s string) {
	s = e.ModTime().Format("Jan 2 15:04")
	return
}

const accessSymbols = "xwr"

func getModeString(e os.FileInfo) (s string) {
	mode := e.Mode()
	if e.IsDir() {
		s = "d"
	} else {
		s = "-"
	}
	for i := 8; i >= 0; i-- {
		if mode&(1<<uint(i)) == 0 {
			s += "-"
		} else {
			char := i % 3
			s += accessSymbols[char : char+1]
		}
	}
	return
}

var sizeSymbols = "BkMGT"

func getSizeString(size int64, humanFlag bool) (s string) {
	if !humanFlag {
		return fmt.Sprintf("%9dB", size)
	}
	var power int
	if size == 0 {
		power = 0
	} else {
		power = int(math.Log(float64(size)) / math.Log(1024.0))
	}
	if power > len(sizeSymbols)-1 {
		power = len(sizeSymbols) - 1
	}
	rSize := float64(size) / math.Pow(1024, float64(power))
	return fmt.Sprintf("%7.1f%s", rSize, sizeSymbols[power:power+1])
}

func getEntryTypeString(e os.FileInfo) string {
	if e.IsDir() {
		return string(os.PathSeparator)
		/*	} else if e.IsBlock() {
				return "<>"
			} else if e.IsFifo() {
				return ">>"
			} else if e.IsSymlink() {
				return "@"
			} else if e.IsSocket() {
				return "&"
			} else if e.IsRegular() && (e.Mode&0001 == 0001) {
				return "*" */
	}
	return ""
}

func getUserString(id int) string {
	return fmt.Sprintf("%03d", id)
}

func getDirList(ls *Ls, inPipe io.Reader) ([]string, error) {
	if len(ls.Filenames) <= 0 {
		if ls.Stdin {
			globs := []string{}
			//check STDIN
			bio := bufio.NewReader(inPipe)
			//defer bio.Close()
			line, hasMoreInLine, err := bio.ReadLine()
			if err == nil {
				//adding from scope.Env.Streams.Stdin
				globs = append(globs, strings.TrimSpace(string(line)))
			} else {
				//ok
			}
			for hasMoreInLine {
				line, hasMoreInLine, err = bio.ReadLine()
				if err == nil {
					//adding from scope.Env.Streams.Stdin
					globs = append(globs, string(line))
				} else {
					//finish
				}
			}
			args := []string{}
			for _, glob := range globs {
				results, err := filepath.Glob(glob)
				if err != nil {
					return args, err
				}
				if len(results) < 1 { //no match
					return args, errors.New("ls: cannot access " + glob + ": No such file or directory")
				}
				args = append(args, results...)
			}

			return args, nil
		}
		//NOT piping. Just use cwd by default.
		cwd, err := os.Getwd()
		return []string{cwd}, err

	}
	return ls.Filenames, nil
}
