package main

import (
	"os"
	"testing"

	"github.com/laher/smoosh/run"
)

func TestExamples(t *testing.T) {
	runner := run.NewRunner()
	files := []string{
		"examples/sm1.smoosh",
		"examples/sm2.smoosh",
		"examples/comment.smoosh",
	}
	pwd, _ := os.Getwd()
	for _, f := range files {
		_ = os.Chdir(pwd)
		err := runner.RunFile(f, os.Stdout)
		if err != nil {
			t.Errorf("Failed to run file: %v", err)
		}
	}
}
