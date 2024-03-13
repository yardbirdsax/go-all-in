package hcl

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	testCases := []struct {
		name    string
		genFunc func() (b []byte, err error)
	}{
		{
			name:    "simple",
			genFunc: Simple,
		},
		{
			name:    "local",
			genFunc: Local,
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantBytes, err := afero.ReadFile(afero.NewOsFs(), fmt.Sprintf("testdata/%s.hcl", tc.name))
			require.NoError(t, err, "error reading wanted file")

			gotBytes, err := tc.genFunc()

			assert.NoError(t, err, "error running genFunc")
			assert.Equal(t, string(wantBytes), string(gotBytes), "got and wanted differed")
		})
	}
}
