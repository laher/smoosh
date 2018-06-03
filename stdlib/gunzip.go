package stdlib

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "t", Help: "Test archive data"},
		object.Flag{Name: "k", Help: "keep gzip file"},
		object.Flag{Name: "c", Help: "output will go to the standard output"},
	}
	RegisterBuiltin("gunzip", &object.Builtin{
		Fn:    gunzip,
		Flags: opts,
	})
}

// Gunzip represents and performs `gunzip` invocations
type Gunzip struct {
	IsTest    bool
	IsKeep    bool
	IsPipeOut bool
	Filenames []string
}

func gunzip(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	gunzip := &Gunzip{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "t":
				gunzip.IsTest = true
			case "k":
				gunzip.IsKeep = true
			case "c":
				gunzip.IsPipeOut = true
			default:
				return object.NewError("flag %s not supported", arg.Name)
			}

		case *object.String:
			//Filenames (globs):
			d, err := Interpolate(env.Export(), arg.Value)
			if err != nil {
				return object.NewError(err.Error())
			}
			gunzip.Filenames = append(gunzip.Filenames, d)
		default:
			return object.NewError("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}

	stdout, stderr := getWriters(out)
	stdin := getReader(in)
	if gunzip.IsTest {
		err := TestGzipItems(gunzip.Filenames)
		if err != nil {
			return object.NewError(err.Error())
		}
	} else {
		err := gunzip.gunzipItems(stdin, stdout, stderr)
		if err != nil {
			return object.NewError(err.Error())
		}
	}

	return Null
}

func TestGzipItems(items []string) error {
	for _, item := range items {
		fh, err := os.Open(item)
		if err != nil {
			return err
		}
		err = TestGzipItem(fh)
		if err != nil {
			return err
		}
	}
	return nil
}

//TODO: proper file checking (how to check validity?)
func TestGzipItem(item io.Reader) error {
	r, err := gzip.NewReader(item)
	if err != nil {
		return err
	}
	defer r.Close()
	return nil
}

func (gunzip *Gunzip) gunzipItems(inPipe io.Reader, outPipe io.Writer, errPipe io.Writer) error {
	if len(gunzip.Filenames) == 0 {
		//in to out
		err := gunzip.gunzipItem(inPipe, outPipe, errPipe, true)
		if err != nil {
			return err
		}
	} else {
		for _, item := range gunzip.Filenames {
			fh, err := os.Open(item)
			if err != nil {
				return err
			}
			err = gunzip.gunzipItem(fh, outPipe, errPipe, gunzip.IsPipeOut)
			if err != nil {
				return err
			}
			err = fh.Close()
			if err != nil {
				return err
			}
			if !gunzip.IsKeep {
				err = os.Remove(item)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (gunzip *Gunzip) gunzipItem(item io.Reader, outPipe io.Writer, errPipe io.Writer, toOut bool) error {
	r, err := gzip.NewReader(item)
	if err != nil {
		return err
	}
	defer r.Close()
	if toOut {
		_, err = io.Copy(outPipe, r)
		if err != nil {
			return err
		}
	} else {
		destFileName := r.Header.Name
		fmt.Fprintln(errPipe, "Filename", destFileName)
		destFile, err := os.Create(destFileName)
		defer destFile.Close()
		if err != nil {
			return err
		}
		_, err = io.Copy(destFile, r)
		if err != nil {
			return err
		}

		err = destFile.Close()
		if err != nil {
			return err
		}
	}
	err = r.Close()
	return err
}
