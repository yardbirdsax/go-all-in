package replace

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeek(t *testing.T) {
	testCases := []struct {
		in     string
		start  int
		length int
		want   string
	}{
		{
			"hello",
			0,
			4,
			"hell",
		},
		{
			"hello",
			1,
			4,
			"ello",
		},
		{
			"hello",
			1,
			5,
			"ello",
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(fmt.Sprintf("%s-%d-%d", tc.in, tc.start, tc.length), func(t *testing.T) {
			got := peek(tc.in, tc.start, tc.length)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReplace(t *testing.T) {
	testCases := []struct {
		in          string
		replace     string
		replacement string
		want        string
	}{
		{
			"hello joe",
			"joe",
			"jane",
			"hello jane",
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(fmt.Sprintf("%s-%s-%s", tc.in, tc.replace, tc.replacement), func(t *testing.T) {
			got := replace(tc.in, tc.replace, tc.replacement)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReplaceV2(t *testing.T) {
	testCases := []struct {
		in          string
		replace     string
		replacement string
		want        string
	}{
		{
			"hello joe",
			"joe",
			"jane",
			"hello jane",
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(fmt.Sprintf("%s-%s-%s", tc.in, tc.replace, tc.replacement), func(t *testing.T) {
			got := replaceV2(tc.in, tc.replace, tc.replacement)

			assert.Equal(t, tc.want, got)
		})
	}
}

func BenchmarkReplace(b *testing.B) {
	for i := 0; i < 100; i += 10 {
		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			str := replicate("aaa", "bbb", i)
			for range b.N {
				replace(str, "aaa", "")
			}
		})
	}
}

func BenchmarkReplaceV2(b *testing.B) {
	for i := 0; i < 100; i += 10 {
		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			str := replicate("aaa", "bbb", i)
			for range b.N {
				replaceV2(str, "aaa", "")
			}
		})
	}
}

func replicate(in, pad string, count int) string {
	b := strings.Builder{}
	b.Grow((len(in) + len(pad)) * count)
	for range count {
		b.WriteString(in)
		b.WriteString(pad)
	}
	return b.String()
}
