package stdlib

import (
	"reflect"
	"testing"
)

func TestDirname(t *testing.T) {
	tests := []struct {
		F        []string
		Expected []string
	}{
		{
			F:        []string{"a/b", "/b"},
			Expected: []string{"a", "/"},
		},
	}
	for _, test := range tests {
		res := dirnames(test.F)
		if !reflect.DeepEqual(test.Expected, res) {
			t.Errorf("%v doenst match expected %v", res, test.Expected)
		}
	}
}
