package replace

import "strings"

func replace(in, replace, replacement string) string {
	for i := 0; i < len(in); i++ {
		if peek(in, i, len(replace)) == replace {
			in = in[0:i] + replacement + in[i+len(replace):]
		}
	}
	return in
}

func replaceV2(in, replace, replacement string) string {
	b := strings.Builder{}
	b.Grow(len(in))
	for i := 0; i < len(in); i++ {
		peeked := peek(in, i, len(replace))
		if peeked == replace {
			b.WriteString(replacement)
			i += len(replace)
			continue
		}
		b.WriteByte(in[i])
	}
	return b.String()
}

func peek(in string, start, length int) string {
	if length > len(in[start:]) {
		return in[start:]
	}
	return in[start : start+length]
}
