package stdlib

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/laher/smoosh/object"
)

func init() {
	RegisterBuiltin("zip", &object.Builtin{
		Fn: z,
	})
}

func z(scope object.Scope, args ...object.Object) (object.Operation, error) {
	filenames, err := interpolateArgs(scope.Env, args, true)
	if err != nil {
		return nil, err
	}
	if len(filenames) < 2 {
		return nil, fmt.Errorf("Fewer than 2 filenames given")
	}
	zipFilename := filenames[0]
	itemsToArchive := filenames[1:]
	return func() object.Object {
		err := zipItems(zipFilename, itemsToArchive)
		if err != nil {
			return object.NewError(err.Error())
		}
		return Null
	}, nil
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
