package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "n", ParamType: object.INTEGER_OBJ},
		object.Flag{Name: "s", ParamType: object.INTEGER_OBJ},
		object.Flag{Name: "F"},
	}
	RegisterBuiltin("tail", &object.Builtin{
		Fn:    tail,
		Flags: opts,
	})
}

// Tail represents and performs a `tail` invocation
type Tail struct {
	Lines              int64
	FollowByDescriptor bool //TODO
	FollowByName       bool
	SleepInterval      float64
	Filenames          []string
}

func tail(scope object.Scope, args ...object.Object) (object.Operation, error) {
	tail := &Tail{SleepInterval: 1.0}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "n":
				l, ok := arg.Param.(*object.Integer)
				if !ok {
					return nil, fmt.Errorf("flag %s parse error", arg.Name)
				}
				tail.Lines = l.Value
			case "F": //follow by name
				tail.FollowByName = true
			case "s": //sleep
				//TODO float
				l, ok := arg.Param.(*object.Integer)
				if !ok {
					return nil, fmt.Errorf("flag %s parse error", arg.Name)
				}
				tail.SleepInterval = float64(l.Value)

			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}
		case *object.String:
			//Filenames (globs):
			d, err := Interpolate(scope.Env.Export(), arg.Value)
			if err != nil {
				return nil, fmt.Errorf(err.Error())
			}
			tail.Filenames = append(tail.Filenames, d)
		default:
			return nil, fmt.Errorf("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}
	stdout, _ := getWriters(scope.Out)
	stdin := getReader(scope.In)
	return func() object.Object {
		err := tail.do(stdout, stdin)
		if err != nil {
			return object.NewError(err.Error())
		}
		return Null
	}, nil
}

func (tail *Tail) do(stdout io.Writer, stdin io.Reader) error {
	if len(tail.Filenames) > 0 {

		for _, fileName := range tail.Filenames {
			finf, err := os.Stat(fileName)
			if err != nil {
				return err
			}
			file, err := os.Open(fileName)
			if err != nil {
				return err
			}
			seek := int64(0)
			if finf.Size() > 10000 {
				//just get last 10K (good enough for now)
				seek = finf.Size() - 10000
				_, err = file.Seek(seek, 0)
				if err != nil {
					return err
				}
			}
			end, err := tailReader(file, seek, tail, stdout)
			if err != nil {
				file.Close()
				return err
			}
			err = file.Close()
			if err != nil {
				return err
			}
			if tail.FollowByName {
				sleepIntervalMs := time.Duration(tail.SleepInterval * 1000)
				for {
					//sleep n.x seconds
					//use milliseconds to get some accuracy with the int64
					time.Sleep(sleepIntervalMs * time.Millisecond)
					finf, err := os.Stat(fileName)
					if err != nil {
						return err
					}
					file, err := os.Open(fileName)
					if err != nil {
						return err
					}
					_, err = file.Seek(end, 0)
					if err != nil {
						return err
					}
					if finf.Size() > end {
						end, err = tailReader(file, end, tail, stdout)
						if err != nil {
							file.Close()
							return err
						}
					} else {
						//TODO start again
					}
					err = file.Close()
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		_, err := tailReader(stdin, 0, tail, stdout)
		if err != nil {
			return err
		}
	}
	return nil
}

func tailReader(file io.Reader, start int64, tail *Tail, out io.Writer) (int64, error) {
	var buffer []string
	end := start
	scanner := bufio.NewScanner(file)
	lastLine := tail.Lines - 1

	for scanner.Scan() {
		text := scanner.Text()
		end += int64(len(text) + 1) //for the \n character
		lastLine++
		if lastLine == tail.Lines {
			lastLine = 0
		}
		if lastLine >= int64(len(buffer)) {
			buffer = append(buffer, text)
		} else {
			buffer[lastLine] = text
		}
	}
	err := scanner.Err()
	if err != nil {
		return end, err
	}

	if lastLine == tail.Lines-1 {
		for _, text := range buffer {
			fmt.Fprintf(out, "%s\n", text)
		}
	} else {
		for _, text := range buffer[lastLine+1:] {
			fmt.Fprintf(out, "%s\n", text)
		}
		//if lastLine > 0 {
		for _, text := range buffer[:lastLine+1] {
			fmt.Fprintf(out, "%s\n", text)
		}
		//}
	}
	return end, nil
}
