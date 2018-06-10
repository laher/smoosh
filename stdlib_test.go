package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/laher/smoosh/run"
)

func TestStdLibNonDestructive(t *testing.T) {
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
		{
			name:   "wc",
			input:  `wc(l, "LICENSE")`,
			expOut: "24 LICENSE\n",
		},
		{
			name:   "basename(pwd())",
			input:  `var x=pwd(); basename(x)`,
			expOut: "smoosh\n",
		},
		{
			name:   "pwd|basename",
			input:  `pwd()|basename()`,
			expOut: "smoosh\n",
		},
		{
			name:   "grep",
			input:  `grep("hello", "testdata/hello.txt")`,
			expOut: "hello\n",
		},
		{
			name:   "grep",
			input:  `grep(H, "hello", "testdata/hello.txt")`,
			expOut: "testdata/hello.txt:hello\n",
		},
		{
			name:   "head",
			input:  `head(n(5), "testdata/100.txt")`,
			expOut: "1\n2\n3\n4\n5\n",
		},
		{
			name:   "tail",
			input:  `tail(n(5), "testdata/100.txt")`,
			expOut: "96\n97\n98\n99\n100\n",
		},
	}
	createFile(t, "testdata/hello.txt", "hello\n")
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			r := run.NewRunner()
			rbuf := bytes.NewBuffer([]byte(test.input))
			wbuf := bytes.NewBuffer([]byte{})
			ebuf := bytes.NewBuffer([]byte{})
			err := r.Run(rbuf, wbuf, ebuf)
			if test.expErr {
				if err == nil {
					t.Errorf("Expected error but none triggered")
				}
			} else if err != nil {
				t.Errorf("Unexpected error: [%s]", err.Error())
			}
			out := string(wbuf.Bytes())
			if out != test.expOut {
				t.Errorf("Unexpected output: [%s](len %d) (expected [%s], len %d)", out, len(out), test.expOut, len(test.expOut))
			}
		})
	}
}

func createFile(t *testing.T, name string, content string) {
	f, err := os.Create(name)
	if err != nil {
		t.Errorf("Error creating file [%s]", err)
		t.FailNow()
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	if err != nil {
		t.Errorf("Error writing file [%s]", err)
		t.FailNow()
	}
}

func checkFileExists(t *testing.T, name string) {
	_, err := os.Stat(name)
	if err != nil {
		t.Errorf("Couldnt stat file [%v]", err)
		return
	}
}

func checkFileNotExists(t *testing.T, name string) {
	_, err := os.Stat(name)
	if !os.IsNotExist(err) {
		t.Errorf("File exists/other-error [%v]", err)
		return
	}
}

func checkFile(t *testing.T, name string, expected string) {
	f2, err := os.Open(name)
	if err != nil {
		t.Errorf("Couldn't stat file [%v]", err)
		return
	}
	defer f2.Close()
	b, err := ioutil.ReadAll(f2)
	if string(b) != expected {
		t.Errorf("Content [%s] did not match. Expected: [%s]", string(b), expected)
	}
}

func deleteFileIfExists(t *testing.T, name string) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return
	}
	deleteFile(t, name)
}

func deleteFile(t *testing.T, name string) {
	err := os.Remove(name)
	if err != nil {
		t.Errorf("Error deleting file [%s]", err)
		t.FailNow()
	}
}

func TestStdLibDestructive(t *testing.T) {
	tests := []struct {
		name  string
		input string
		setup func()
		check func(mbuf, ebuf io.Reader, runErr error)
	}{
		{
			name:  "mv",
			input: `mv("testdata/tmp.txt", "testdata/tmp2.txt")`,
			setup: func() {
				createFile(t, "testdata/tmp.txt", "abcabcabc")
			},
			check: func(mbuf io.Reader, ebuf io.Reader, runErr error) {
				if _, err := os.Stat("testdata/tmp.txt"); !os.IsNotExist(err) {
					t.Errorf("tmp.txt should not exist [%v]", err)
				}
				checkFile(t, "testdata/tmp2.txt", "abcabcabc")
				deleteFile(t, "testdata/tmp2.txt")
			},
		},
		{
			name:  "cp",
			input: `cp("testdata/tmp.txt", "testdata/tmp2.txt")`,
			setup: func() {
				createFile(t, "testdata/tmp.txt", "abcabcabc")
			},
			check: func(mbuf io.Reader, ebuf io.Reader, runErr error) {
				checkFile(t, "testdata/tmp.txt", "abcabcabc")
				checkFile(t, "testdata/tmp2.txt", "abcabcabc")
				deleteFile(t, "testdata/tmp.txt")
				deleteFile(t, "testdata/tmp2.txt")
			},
		},
		{
			name:  "gzip",
			input: `cd("testdata"); gzip("tmp.txt"); cd("..")`,
			setup: func() {
				createFile(t, "testdata/tmp.txt", "abcabcabc")
			},
			check: func(mbuf io.Reader, ebuf io.Reader, runErr error) {
				checkFileExists(t, "testdata/tmp.txt.gz")
				checkFileNotExists(t, "testdata/tmp.txt")
				deleteFile(t, "testdata/tmp.txt.gz")
			},
		},
		{
			name:  "gzip;gunzip",
			input: `cd("testdata"); gzip("tmp.txt"); gunzip("tmp.txt.gz"); cd("..")`,
			setup: func() {
				createFile(t, "testdata/tmp.txt", "abcabcabc")
			},
			check: func(mbuf io.Reader, ebuf io.Reader, runErr error) {
				checkFile(t, "testdata/tmp.txt", "abcabcabc")
				checkFileNotExists(t, "testdata/tmp.txt.gz")
				deleteFile(t, "testdata/tmp.txt")
			},
		},
	}
	for i := range tests {
		test := tests[i]
		t.Logf("Running: [%s]", test.name)
		test.setup()
		wbuf := bytes.NewBuffer([]byte{})
		ebuf := bytes.NewBuffer([]byte{})
		var err error
		r := run.NewRunner()
		rbuf := bytes.NewBuffer([]byte(test.input))
		err = r.Run(rbuf, wbuf, ebuf)
		if err != nil {
			t.Errorf("Unexpected error: [%s]", err.Error())
		}
		test.check(wbuf, ebuf, err)
	}
	//result := evaluator.Eval(program, env)
	/*
		if result.Type() != "" {
			t.Errorf("%v", p.Errors())
		}
	*/
}
