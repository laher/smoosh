package stdlib

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "t", Help: "Test archive data"},
		object.Flag{Name: "d", Help: "destination directory", ParamType: object.STRING_OBJ},
	}
	RegisterBuiltin("unzip", &object.Builtin{
		Fn:    unzip,
		Flags: opts,
	})
}

// Unzip represents and performs `unzip` invocations
type Unzip struct {
	isTest    bool
	destDir   string
	Filenames []string
	ZipFile   string
}

func unzip(scope object.Scope, args ...object.Object) (object.Operation, error) {
	unzip := &Unzip{destDir: "."}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "t":
				unzip.isTest = true
			case "d":
				s, ok := arg.Param.(*object.String)
				if !ok {
					return nil, fmt.Errorf("flag %s does not have a valid parameter", arg.Name)
				}
				unzip.destDir = s.Value
			default:
				return nil, fmt.Errorf("flag %s not supported", arg.Name)
			}

		case *object.String:
			//Filenames (globs):
			d, err := Interpolate(scope.Env.Export(), arg.Value)
			if err != nil {
				return nil, fmt.Errorf(err.Error())
			}
			if unzip.ZipFile == "" {
				unzip.ZipFile = d
			} else {
				unzip.Filenames = append(unzip.Filenames, d)
			}
		default:
			return nil, fmt.Errorf("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}

	stdout, stderr := getWriters(scope.Out)
	return func() object.Object {
		if unzip.isTest {
			err := testItems(unzip.ZipFile, unzip.Filenames, stdout, stderr)
			if err != nil {
				return object.NewError(err.Error())
			}
		} else {
			err := unzipItems(unzip.ZipFile, unzip.destDir, unzip.Filenames, stderr)
			if err != nil {
				return object.NewError(err.Error())
			}
		}
		return Null
	}, nil
}

func testItems(zipfile string, includeFiles []string, outPipe io.Writer, errPipe io.Writer) error {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		flags := f.FileHeader.Flags
		if len(includeFiles) == 0 || containsGlob(includeFiles, f.Name, errPipe) {
			if flags&1 == 1 {
				fmt.Fprintf(outPipe, "[Password Protected:] %s\n", f.Name)
			} else {
				fmt.Fprintf(outPipe, "%s\n", f.Name)
			}
		}
	}
	return nil
}

func containsGlob(haystack []string, needle string, errPipe io.Writer) bool {
	for _, item := range haystack {
		m, err := filepath.Match(item, needle)
		if err != nil {
			fmt.Fprintf(errPipe, "Glob error %v", err)
			return false
		}
		if m == true {
			return true
		}
	}
	return false
}

func unzipItems(zipfile, destDir string, includeFiles []string, errPipe io.Writer) error {

	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return err
	}
	defer r.Close()

	dinf, err := os.Stat(destDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		} else {
			//doesnt exist
			err = os.MkdirAll(destDir, 0777) //TODO review permissions
			if err != nil {
				return err
			}
		}
	} else {
		if !dinf.IsDir() {
			return errors.New("destination is an existing non-directory")
		}
	}

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		finf := f.FileHeader.FileInfo()
		flags := f.FileHeader.Flags
		if flags&1 == 1 {
			fmt.Fprintf(errPipe, "WARN: Skipping password protected file (flags %v, '%s')\n", flags, f.Name)
		} else {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			destFileName := filepath.Join(destDir, f.Name)
			if finf.IsDir() {
				//mkdir ...
				fdinf, err := os.Stat(destFileName)
				if err != nil {
					if !os.IsNotExist(err) {
						return err
					}
					//doesnt exist
					err = os.MkdirAll(destFileName, finf.Mode())
					if err != nil {
						return err
					}
				} else {
					if !fdinf.IsDir() {
						return errors.New("destination " + destFileName + " is an existing non-directory")
					}
				}
			} else {
				fileDestDir := filepath.Dir(destFileName)
				if fileDestDir != destDir {
					fdinf, err := os.Stat(fileDestDir)
					if err != nil {
						if !os.IsNotExist(err) {
							return err
						}
						//doesnt exist
						err = os.MkdirAll(fileDestDir, 0777) //TODO review dir permissions
						if err != nil {
							return err
						}
					} else {
						if !fdinf.IsDir() {
							return errors.New("destination " + fileDestDir + " is an existing non-directory")
						}
					}
				}
				//TODO remove on error
				destFile, err := os.OpenFile(destFileName, os.O_CREATE, finf.Mode())
				defer destFile.Close()
				if err != nil {
					return err
				}
				_, err = io.Copy(destFile, rc)
				if err != nil {
					return err
				}

			}
			rc.Close()
		}
	}
	return nil
}
