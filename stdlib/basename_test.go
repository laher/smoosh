package stdlib

import (
	"reflect"
	"testing"
)

func TestBasename(t *testing.T) {
	tests := []struct {
		path       string
		relativeTo string
		expected   string
	}{
		{"", "", "."},
		{"a/b", "", "b"},
		{"/a/b/c/", "", "c"},
	}
	for _, test := range tests {
		res := basenameFile(test.path, test.relativeTo)
		if !reflect.DeepEqual(test.expected, res) {
			t.Errorf("[%v] doenst match expected [%v]", res, test.expected)
		}
	}
}
