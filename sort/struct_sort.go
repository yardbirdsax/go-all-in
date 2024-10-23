package sort

import (
	"cmp"
	"slices"
)

// Struct is, well, a struct! It has two fields to sort on.
type Struct struct {
	Foo string
	Bar string
}

// SortStruct sorts a slice of Structs by Foo and Bar in ascending order.
func SortStruct(sl []Struct) {
	slices.SortFunc(sl, func(a Struct, b Struct) int {
		if c := cmp.Compare(a.Foo, b.Foo); c != 0 {
			return c
		}
		return cmp.Compare(a.Bar, b.Bar)
	})
}
