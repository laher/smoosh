// Package run provides a hook into the smoosh interpreter.
package run

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/laher/smoosh/object"
)

func isPipedInput(in io.Reader) bool {
	if stdin, ok := in.(*os.File); ok {
		stat, _ := stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return true
		}
	}
	return false
}

// Start starts a line-by-line processor
func (r *Runner) Start(in io.Reader, out io.Writer, stderr io.Writer) {
	streams := object.GlobalStreams{in, out, stderr}
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment(streams)
	macroEnv := object.NewEnvironment(streams)
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	if isPipedInput(in) {
		all, err := ioutil.ReadAll(in)
		if err != nil {
			panic(err)
		}
		err = r.runData(string(all), out, env, macroEnv)
		if err != nil {
			panic(err)
		}
		return
	}
	for {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		prompt := fmt.Sprintf("[%s]/[%s]> ", user.Username, path.Base(pwd))
		fmt.Printf(prompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		err = r.runData(line, out, env, macroEnv)
		if err != nil {
			panic(err)
		}
	}
}

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
