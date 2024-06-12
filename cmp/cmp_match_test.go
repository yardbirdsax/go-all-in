package cmp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {
	testCases := []struct {
		name string
		a    any
		b    any
		want string
	}{
		{
			name: "mulit-line string",
			a: `hello
josh`,
			b: `hello
there
josh`,
			want: `  "hello"
+ "there"
  "josh"
`,
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			got := Diff[string](tc.a, tc.b)

			assert.Equal(t, tc.want, got, "got and wanted diff are not the same")
		})
	}
}
