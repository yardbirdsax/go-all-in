package hcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func Simple() (b []byte, err error) {
	f := hclwrite.NewFile()
	tfBlock := f.Body().AppendNewBlock("module", []string{"something"})
	tfBlock.Body().SetAttributeValue("source", cty.StringVal("something"))
	return f.Bytes(), err
}

func Local() (b []byte, err error) {
	f := hclwrite.NewFile()
	tfBlock := f.Body().AppendNewBlock("module", []string{"something"})
	tfBlock.Body().SetAttributeTraversal("source", hcl.Traversal{
		hcl.TraverseRoot{Name: "local"},
		hcl.TraverseAttr{Name: "something"},
	})
	return f.Bytes(), err
}
