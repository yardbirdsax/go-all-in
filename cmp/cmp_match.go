package cmp

import (
	"fmt"
	"reflect"
	"strings"
)

func Diff[E comparable](a, b any) string {
	diff := ""
	if reflect.TypeOf(a).Kind().String() == "string" {
		a = strings.Split(a.(string), "\n")
	}
	if reflect.TypeOf(b).Kind().String() == "string" {
		b = strings.Split(b.(string), "\n")
	}
	if reflect.TypeOf(a).Kind().String() == "slice" && reflect.TypeOf(b).Kind().String() == "slice" {
		return diffSlice(a.([]E), b.([]E))
	}

	return diff
}

func diffSlice[S []E, E comparable](a S, b S) string {
	diff := ""
	offset := 0
	var (
		iA = 0
		iB = 0
	)
	greater := greaterOf(len(a), len(b))
	for iA = 0; iA < greater && iB < greater; {
		eA := a[iA]
		eB := b[iB]
		switch {
		case eA == eB:
			iA++
			iB++
			diff += fmt.Sprintf("  %+#v\n", eA)
		default:
			diff += fmt.Sprintf("+ %+#v\n", eB)
			offset++
		}
		iB = iA + offset
	}
	return diff
}

func greaterOf(a, b int) int {
	switch {
	case a >= b:
		return a
	default:
		return b
	}
}
