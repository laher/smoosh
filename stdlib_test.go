package main

import (
	"bytes"
	"testing"

	"github.com/laher/smoosh/run"
)

func TestStdLib(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expErr    bool
		expOut    string
		expStderr string
	}{
		{
			name:   "cat",
			input:  `cat("testdata/hello.txt")`,
			expOut: "hello\n",
		},
		{
			name:   "cat-file-not-exist",
			input:  `cat("testdata/hello.txtx")`,
			expErr: true,
		},
		{
			name:   "basename",
			input:  `basename("testdata/hello.txt")`,
			expOut: "hello.txt\n",
		},
		{
			name:   "basename",
			input:  `basename("testdata/")`,
			expOut: "testdata\n",
		},
		{
			name:   "dirname",
			input:  `dirname("testdata/hello.txt")`,
			expOut: "testdata\n",
		},
		{
			name:   "echo",
			input:  `echo("hello")`,
			expOut: "hello\n",
		},
		{
			name:   "ls",
			input:  `ls("testdata/hello.txt")`,
			expOut: "hello.txt \n",
		},
		{
			name:   "echo|ls",
			input:  `echo("testdata/hello.txt")|ls()`,
			expOut: "hello.txt \n",
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			r := run.NewRunner()
			rbuf := bytes.NewBuffer([]byte(test.input))
			wbuf := bytes.NewBuffer([]byte{})
			ebuf := bytes.NewBuffer([]byte{})
			err := r.Run(rbuf, wbuf, ebuf)
			if test.expErr && err == nil {
				t.Errorf("Expected error but none triggered")
			} else if !test.expErr && err != nil {
				t.Errorf("Unexpected error: [%s]", err.Error())
			}
			out := string(wbuf.Bytes())
			if out != test.expOut {
				t.Errorf("Unexpected output: [%s](len %d) (expected [%s], len %d)", out, len(out), test.expOut, len(test.expOut))
			}
		})
	}
	//result := evaluator.Eval(program, env)
	/*
		if result.Type() != "" {
			t.Errorf("%v", p.Errors())
		}
	*/
}
