package stdlib

import (
	"bytes"
	"reflect"
	"testing"
)

func TestCat(t *testing.T) {
	tests := []struct {
		f        []string
		ends     bool
		n        bool
		sq       bool
		stdin    string
		expected string
	}{
		{
			stdin:    "hello",
			expected: "hello",
		},
	}
	for _, test := range tests {
		in := bytes.NewBuffer([]byte(test.stdin))
		out := bytes.NewBuffer([]byte{})
		err := catIt(out, in, test.f, test.ends, test.n, test.sq)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}
		res := string(out.Bytes())
		if !reflect.DeepEqual(test.expected, res) {
			t.Errorf("%v doenst match expected %v", res, test.expected)
		}
	}
}
