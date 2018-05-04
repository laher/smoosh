package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/laher/smoosh/run"
)

func TestExamples(t *testing.T) {
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
			err := runner.RunFile(f, os.Stdout)
			if err != nil {
				t.Errorf("Failed to run file: %v", err)
			}
		})
	}
}
