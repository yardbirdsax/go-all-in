package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
)

type StructComment struct {
  PackageName string
  File string
  Name string
  Comment string
  FieldComments []FieldComment
}

type FieldComment struct {
  Name string
  Type string
  Comment string
}

// GetStructComments retrieves all comments for structs in the files in a given list of paths.
func GetStructComments(paths ...string) []StructComment {
  fileset := token.NewFileSet()
  structComments := []StructComment{}

  for _, path := range paths {
    packages, err := parser.ParseDir(fileset, path, nil, parser.ParseComments)
    if err != nil {
      return structComments
    }
    for _, pkg := range packages {
      for _, f := range pkg.Files {
        ast.Inspect(f, func(n ast.Node) bool {
          typeSpec, ok := n.(*ast.TypeSpec)
          if ok && reflect.TypeOf(typeSpec.Type).Elem().Name() == "StructType" {
            structComment := StructComment{
              PackageName: pkg.Name,
              File: f.Name.Name,
              Name: typeSpec.Name.Name,
            }

            structLine := fileset.Position(typeSpec.Pos()).Line
            for _, cGroup := range f.Comments {
              if fileset.Position(cGroup.End()).Line == structLine - 1{
                structComment.Comment = cGroup.Text()
              }
            }

            structType := n.(*ast.StructType)
            for _, fld := range structType.Fields.List {
              fldLine := fileset.Position(fld.Pos()).Line
              fldComment := FieldComment{
                Name: fld.Names[0].Name,
                Type: reflect.TypeOf(fld.Type).Elem().Name(),
              }
              for _, cGroup := range f.Comments {
                if fileset.Position(cGroup.End()).Line == fldLine - 1 {
                  fldComment.Comment = cGroup.Text()
                }
              }
            }

            structComments = append(structComments, structComment)
          }
          return true
        })
      }
    }
  }
  return structComments
}