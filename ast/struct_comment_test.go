package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStructComments(t *testing.T) {
  expectedComments := []StructComment{
    {
      PackageName: "astor",
      File: "astor",
      Name: "Astor",
      Comment: "Astor is a struct used to showcase how to do type level introspection / parsing with `ast`.\n" +
               "All of these lines are considered to be in the same Node / comment block, since they are on\n" +
               "subsequent lines with no breaks.\n",
    },
    {
      PackageName: "astor",
      Name: "Astor2",
      File: "astor",
      Comment: "Astor2 is just another struct for testing purposes.\n",
    },
  }

  actualComment := GetStructComments("astor")

  assert.Equal(t, expectedComments, actualComment)
}