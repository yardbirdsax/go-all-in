// This is a package comment.
package ast

import (
	"go/parser"
	"go/token"
	"strings"
)

// GetPackageComment returns all package level comments for files in the path specified.
func GetPackageComments(paths ...string) map[string]string {
  fileset := token.NewFileSet()
  allPackageComments := map[string]string{}
  
  for _, path := range paths {
    packages, err := parser.ParseDir(fileset, path, nil, parser.ParseComments)
    if err != nil {
      return nil
    }
    for _, p := range packages {
      packageComments := []string{}
      for _, f := range p.Files {
        if f.Doc != nil {
          packageComments = append(packageComments, f.Doc.Text())
        }
      }
      allPackageComments[p.Name] = strings.Join(packageComments, "\n")
    }
  }
  return allPackageComments
}