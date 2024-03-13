package hcl

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModify(t *testing.T) {
	testCases := []struct {
		name    string
		genFunc func(input string) (b []byte, err error)
		input   string
	}{
		{
			name:    "local",
			genFunc: ModifyLocal,
			input:   "module \"something\" {\n  source = local.somethingelse \n}\n\nmodule \"somethingelse\" {\n  source = local.somethingelse \n}\n",
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantBytes, err := afero.ReadFile(afero.NewOsFs(), fmt.Sprintf("testdata/modify/%s.hcl", tc.name))
			require.NoError(t, err, "error reading wanted file")

			gotBytes, err := tc.genFunc(tc.input)

			assert.NoError(t, err, "error running genFunc")
			assert.Equal(t, string(wantBytes), string(gotBytes), "got and wanted differed")
		})
	}
}
