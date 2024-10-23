package sort

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortStruct(t *testing.T) {
	sl := []Struct{
		{
			Foo: "A",
			Bar: "A",
		},
		{
			Foo: "A",
			Bar: "B",
		},
		{
			Foo: "A",
			Bar: "C",
		},
		{
			Foo: "B",
			Bar: "A",
		},
		{
			Foo: "B",
			Bar: "B",
		},
		{
			Foo: "B",
			Bar: "C",
		},
	}
	randomSlice := make([]Struct, len(sl))
	copy(randomSlice, sl)
	randomizeSlice(randomSlice)

	SortStruct(randomSlice)

	assert.Equal(t, sl, randomSlice)
}

func randomizeSlice[S ~[]E, E any](sl S) {
	for i := range sl {
		j := rand.Intn(i + 1)
		sl[i], sl[j] = sl[j], sl[i]
	}
}
