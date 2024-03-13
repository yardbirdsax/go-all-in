package hcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func ModifyLocal(input string) (b []byte, err error) {
	f, diags := hclwrite.ParseConfig([]byte(input), "file", hcl.InitialPos)
	if diags.HasErrors() {
		return b, diags
	}
	for _, bod := range f.Body().Blocks() {
		if bod.Type() == "module" && bod.Labels()[0] == "something" {
			bod.Body().SetAttributeTraversal("source", hcl.Traversal{
				hcl.TraverseRoot{Name: "local"},
				hcl.TraverseAttr{Name: "something"},
			})
		}
	}
	b = f.Bytes()
	return b, err
}
