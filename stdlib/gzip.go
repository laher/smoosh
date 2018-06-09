package stdlib

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "t", Help: "Test archive data"},
		object.Flag{Name: "k", Help: "keep gzip file"},
		object.Flag{Name: "c", Help: "output will go to the standard output"},
	}
	RegisterBuiltin("gzip", &object.Builtin{
		Fn:    gz,
		Flags: opts,
	})
}

// Gzip represents and performs `gz` invocations
type Gzip struct {
	IsKeep    bool
	IsStdout  bool
	Filenames []string
	outFile   string
}

func gz(scope object.Scope, args ...object.Object) (object.Operation, error) {
	gz := &Gzip{}
	var err error
	gz.Filenames, err = interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "k":
				gz.IsKeep = true
			case "c":
				gz.IsStdout = true
			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}
		}
	}

	return func() object.Object {
		if len(gz.Filenames) < 1 {
			//pipe in?
			var writer io.Writer
			var outputFilename string
			if gz.outFile != "" {
				outputFilename = gz.outFile
				var err error
				w, err := os.Create(outputFilename)
				if err != nil {
					return object.NewError(err.Error())
				}
				defer w.Close()
				writer = w
			} else {
				//	fmt.Printf("scope.Env.Streams.Stdin to scope.Env.Streams.Stdout: %+v\n", gz)
				outputFilename = "S" //seems to be the default used by gzip
				writer = scope.Env.Streams.Stdout
			}
			err := gz.doGzip(scope.Env.Streams.Stdin, writer, filepath.Base(outputFilename))
			if err != nil {
				return object.NewError(err.Error())
			}
		} else {
			//todo make sure it closes saved file cleanly
			for _, inputFilename := range gz.Filenames {
				inputFile, err := os.Open(inputFilename)
				if err != nil {
					return object.NewError(err.Error())
				}
				defer inputFile.Close()

				var writer io.Writer
				if !gz.IsStdout {
					outputFilename := inputFilename + ".gz"
					gzf, err := os.Create(outputFilename)
					if err != nil {
						return object.NewError(err.Error())
					}
					defer gzf.Close()
					writer = gzf
				} else {
					writer = scope.Env.Streams.Stdout
				}
				err = gz.doGzip(inputFile, writer, filepath.Base(inputFilename))
				if err != nil {
					return object.NewError(err.Error())
				}

				err = inputFile.Close()
				if err != nil {
					return object.NewError(err.Error())
				}

				// only remove source if specified and possible
				if !gz.IsKeep && !gz.IsStdout {
					err = os.Remove(inputFilename)
					if err != nil {
						return object.NewError(err.Error())
					}
				}
			}
		}
		return Null
	}, nil
}

func (gz *Gzip) doGzip(reader io.Reader, writer io.Writer, filename string) error {

	inw := new(bytes.Buffer)
	_, err := io.Copy(inw, reader)
	if err != nil {
		return err
	}
	rdr := strings.NewReader(inw.String())

	outw := new(bytes.Buffer)

	gzw := gzip.NewWriter(outw)
	defer gzw.Close()
	gzw.Header.Comment = "Gzipped by smoosh"
	gzw.Header.Name = filename

	_, err = io.Copy(gzw, rdr)
	if err != nil {
		fmt.Println("Copied err", err)
		return err
	}
	//get error where possible
	err = gzw.Close()
	if err != nil {
		fmt.Println("Closed err", err)
		return err
	}

	_, err = io.Copy(writer, outw)

	return err
}
