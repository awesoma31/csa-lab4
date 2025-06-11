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
		{"hello", "golden/hello"},
		{"cat", "golden/cat"},
		{"hello_username", "golden/hello_user"},
		{"sort", "golden/sort"},
		// {"variant", "golden/alg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testingutil.RunGolden(t, tt.dir, 20_000)
		})
	}
}
