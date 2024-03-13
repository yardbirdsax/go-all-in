package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkCustom(b *testing.B) {
	want := &sampleWithUnmarshal{
		Count: 10,
		Items: []itemWithUnmarshal{
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
	cust := &custom{}

	var (
		got *sampleWithUnmarshal
		err error
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err = cust.Unmarshal("testdata/sample.json")
	}

	assert.Equal(b, want, got, "want and got differ")
	assert.NoError(b, err, "err is not nil")
}

func BenchmarkCustomV2(b *testing.B) {
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
	cust := &custom{}

	var (
		got *sample
		err error
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got, err = cust.UnmarshalV2("testdata/sample.json")
	}

	assert.Equal(b, want, got, "want and got differ")
	assert.NoError(b, err, "err is not nil")
}
