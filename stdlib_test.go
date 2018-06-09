package main

import (
	"bytes"
	"testing"

	"github.com/laher/smoosh/run"
)

func TestCatInt(t *testing.T) {
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
			name:   "cat",
			input:  `cat("testdata/hello.txt")`,
			expOut: "hello\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := run.NewRunner()
			rbuf := bytes.NewBuffer([]byte(test.input))
			wbuf := bytes.NewBuffer([]byte{})
			ebuf := bytes.NewBuffer([]byte{})
			err := r.Run(rbuf, wbuf, ebuf)
			if err != nil {
				t.Errorf(err.Error())
			}
			out := string(wbuf.Bytes())
			if out != test.expOut {
				t.Errorf("Incorrect output for cat: [%s] (expected [%s])", out, test.expOut)
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
