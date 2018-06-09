package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/laher/smoosh/run"
)

func TestGood(t *testing.T) {
	runner := run.NewRunner()
	files, err := filepath.Glob("testdata/*.smoosh")
	if err != nil {
		t.Errorf("failed: %s", err)
		t.FailNow()
	}
	pwd, _ := os.Getwd()
	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			//in case of directory changes in-script
			_ = os.Chdir(pwd)
			err := runner.RunFile(f, os.Stdout, os.Stderr)
			if err != nil {
				t.Errorf("Failed to run file: %v", err)
			}
		})
	}
}

func TestBad(t *testing.T) {
	runner := run.NewRunner()
	files, err := filepath.Glob("testdata/bad/*.smoosh")
	if err != nil {
		t.Errorf("failed: %s", err)
		t.FailNow()
	}
	pwd, _ := os.Getwd()
	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			//in case of directory changes in-script
			_ = os.Chdir(pwd)
			err := runner.RunFile(f, os.Stdout, os.Stderr)
			if err != nil {
				t.Logf("Error as expected ... '%s'", err)
				return
			}
			t.Errorf("File should have errored")
		})
	}
}
