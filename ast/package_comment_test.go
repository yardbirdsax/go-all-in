package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageComment(t *testing.T) {
  expectedComments := map[string]string{
    "aster": "Package aster is an example used to help figure out using `ast` to parse and consume comments in Go code.\n",
  }

  actualComment := GetPackageComments("astor")

  assert.Equal(t, expectedComments, actualComment)
}