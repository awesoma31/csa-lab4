package golden_test

import (
	"testing"

	"github.com/awesoma31/csa-lab4/internal/testingutil"
)

func TestGolden(t *testing.T) {
	tests := []struct {
		name string
		dir  string
	}{
		{"hello", "hello"},
		{"cat", "cat"},
		{"hello_username", "hello_user"},
		{"sort", "sort"},
		{"alg", "alg"},
		{"math", "math"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testingutil.RunGolden(t, tt.dir)
		})
	}
}
