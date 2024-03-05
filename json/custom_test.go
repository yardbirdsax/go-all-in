package json

import (
	"runtime"
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

func BenchmarkCustomItem(b *testing.B) {
	var err error
	runtime.MemProfileRate = 1
	data := []byte(`{"foo": "hello", "bar": 1}`)
	it := &itemWithUnmarshal{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = it.UnmarshalJSON(data)
	}
	if err != nil {
		b.Errorf("error unmarshaling: %s", err.Error())
	}
}
