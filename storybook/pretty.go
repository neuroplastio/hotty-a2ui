package storybook

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode/utf8"
)

// prettyWidth is how wide a line of pretty JSON may be before an object or
// array on it is laid out a member a line.
const prettyWidth = 80

// pretty is JSON as panel shows it: indented by two spaces, but an object
// or array that fits on its line, with its indent, stays on it, with a
// space after each colon and comma. Keys keep their order and strings
// their escapes. What isn't JSON comes back as it was.
func pretty(raw []byte) string {
	var b strings.Builder
	if !writePretty(&b, raw, 0, 0) {
		return string(raw)
	}
	return b.String()
}

// writePretty writes raw to b at an indent of depth levels, starting at
// column col of a line already begun.
func writePretty(b *strings.Builder, raw []byte, depth, col int) bool {
	var c bytes.Buffer
	if json.Compact(&c, raw) != nil {
		return false
	}
	one := spaced(c.Bytes())
	open := one[0]
	if open != '{' && open != '[' || col+utf8.RuneCountInString(one) <= prettyWidth {
		b.WriteString(one)
		return true
	}
	d := json.NewDecoder(bytes.NewReader(c.Bytes()))
	d.UseNumber()
	if _, err := d.Token(); err != nil {
		return false
	}
	pad := strings.Repeat("  ", depth+1)
	b.WriteByte(open)
	for first := true; d.More(); first = false {
		if !first {
			b.WriteByte(',')
		}
		b.WriteString("\n" + pad)
		at := len(pad)
		if open == '{' {
			k, err := d.Token()
			if err != nil {
				return false
			}
			key, _ := k.(string)
			q := quote(key) + ": "
			b.WriteString(q)
			at += utf8.RuneCountInString(q)
		}
		var v json.RawMessage
		if d.Decode(&v) != nil || !writePretty(b, v, depth+1, at) {
			return false
		}
	}
	b.WriteString("\n" + strings.Repeat("  ", depth))
	if open == '{' {
		b.WriteByte('}')
	} else {
		b.WriteByte(']')
	}
	return true
}

// spaced is compact JSON with a space after each colon and comma outside
// its strings.
func spaced(compact []byte) string {
	var b strings.Builder
	in, esc := false, false
	for _, c := range compact {
		b.WriteByte(c)
		switch {
		case esc:
			esc = false
		case in && c == '\\':
			esc = true
		case c == '"':
			in = !in
		case !in && (c == ':' || c == ','):
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// quote is s as a JSON string, <, > and & as they are.
func quote(s string) string {
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	e.SetEscapeHTML(false)
	_ = e.Encode(s)
	return strings.TrimSuffix(b.String(), "\n")
}
