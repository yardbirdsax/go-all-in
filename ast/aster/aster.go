package aster

// This should not show up
type Something int

// This line isn't matched to anything because it is off by itself.

// Astor is a struct used to showcase how to do type level introspection / parsing with `ast`.
// All of these lines are considered to be in the same Node / comment block, since they are on
// subsequent lines with no breaks.
type Astor struct { // This is another comment
  // Field is a field used to showcase how to access field level comments with `ast`.
  Field string
}

// Astor2 is just another struct for testing purposes.
type Astor2 struct {

}