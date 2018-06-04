package stdlib

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/laher/smoosh/ast"
	"github.com/laher/smoosh/object"
)

func init() {
	var opts = []object.Flag{
		object.Flag{Name: "t", Help: "Test archive data"},
	}
	RegisterBuiltin("zip", &object.Builtin{
		Fn:    z,
		Flags: opts,
	})
}

// Zip represents and performs `gz` invocations
type Zip struct {
	Filenames []string
	test      bool
	outFile   string
}

func z(env *object.Environment, in, out *ast.Pipes, args ...object.Object) object.Object {
	gz := &Zip{}
	for i := range args {
		switch arg := args[i].(type) {
		case *object.Flag:
			switch arg.Name {
			case "t":
				gz.test = true
			default:
				return object.NewError("flag %s not supported", arg.Name)
			}

		case *object.String:
			//Filenames (globs):
			d, err := Interpolate(env.Export(), arg.Value)
			if err != nil {
				return object.NewError(err.Error())
			}
			gz.Filenames = append(gz.Filenames, d)
		default:
			return object.NewError("argument %d not supported, got %s", i,
				args[0].Type())
		}
	}

	if len(gz.Filenames) < 2 {
		return object.NewError("Fewer than 2 filenames given")
	}
	zipFilename := gz.Filenames[0]
	itemsToArchive := gz.Filenames[1:]
	zipItems(zipFilename, itemsToArchive)
	return Null
}

func zipItems(zipFilename string, itemsToArchive []string) error {
	_, err := os.Stat(zipFilename)
	var zf *os.File
	if err != nil {
		if os.IsNotExist(err) {
			zf, err = os.Create(zipFilename)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		zf, err = os.Create(zipFilename)
		if err != nil {
			return err
		}
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	defer zw.Close()

	//resources
	for _, itemS := range itemsToArchive {
		//todo: relative/full path checking
		item := archiveItem{itemS, itemS}
		err = addFileToZIP(zw, item)
		if err != nil {
			return err
		}
	}
	//get error where possible
	err = zw.Close()
	return err
}

type archiveItem struct {
	fileSystemPath string
	archivePath    string
}

func addFileToZIP(zw *zip.Writer, item archiveItem) error {
	//fmt.Printf("Adding %s\n", item.FileSystemPath)
	binfo, err := os.Stat(item.fileSystemPath)
	if err != nil {
		return err
	}
	if binfo.IsDir() {
		header, err := zip.FileInfoHeader(binfo)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		header.Name = item.archivePath
		_, err = zw.CreateHeader(header)
		if err != nil {
			return err
		}
		file, err := os.Open(item.fileSystemPath)
		if err != nil {
			return err
		}
		fis, err := file.Readdir(0)
		for _, fi := range fis {
			err = addFileToZIP(zw, archiveItem{filepath.Join(item.fileSystemPath, fi.Name()), filepath.Join(item.archivePath, fi.Name())})
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(binfo)
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		header.Name = item.archivePath
		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		bf, err := os.Open(item.fileSystemPath)
		if err != nil {
			return err
		}
		defer bf.Close()
		_, err = io.Copy(w, bf)
		if err != nil {
			return err
		}
	}
	return err
}
