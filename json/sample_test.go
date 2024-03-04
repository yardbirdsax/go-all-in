package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkStdLib(b *testing.B) {
	want := &sample{
		Count: 10,
		Items: []item{
			{
				Foo: "blah",
				Bar: 1,
			},
			{
				Foo: "blah",
				Bar: 2,
			},
			{
				Foo: "blah",
				Bar: 3,
			},
			{
				Foo: "blah",
				Bar: 4,
			},
			{
				Foo: "blah",
				Bar: 5,
			},
			{
				Foo: "blah",
				Bar: 6,
			},
			{
				Foo: "blah",
				Bar: 7,
			},
			{
				Foo: "blah",
				Bar: 8,
			},
			{
				Foo: "blah",
				Bar: 9,
			},
			{
				Foo: "blah",
				Bar: 10,
			},
		},
	}
	std := &stdlib{}

	var (
		got *sample
		err error
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err = std.Unmarshal("testdata/sample.json")
	}

	assert.Equal(b, want, got, "want and got differ")
	assert.NoError(b, err, "err is not nil")
}
