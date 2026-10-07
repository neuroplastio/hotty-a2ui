package a2ui

import (
	"fmt"
	"strings"
)

// Pointer is a JSON Pointer (RFC 6901) taken apart into its reference
// tokens, unescaped. The empty Pointer is the whole document.
type Pointer []string

// ParsePointer reads an absolute JSON Pointer as A2UI's data model does:
// "" and "/" are the whole model, and empty segments are dropped, so
// "//a///b/" is "/a/b" (RFC 6901 would read empty keys).
func ParsePointer(s string) (Pointer, error) {
	if s != "" && s[0] != '/' {
		return nil, fmt.Errorf("a2ui: %q is not an absolute JSON Pointer", s)
	}
	p := Pointer{}
	for _, t := range strings.Split(s, "/") {
		if t == "" {
			continue
		}
		u, err := unescapeToken(t)
		if err != nil {
			return nil, fmt.Errorf("a2ui: %q: %w", s, err)
		}
		p = append(p, u)
	}
	return p, nil
}

func unescapeToken(t string) (string, error) {
	if !strings.Contains(t, "~") {
		return t, nil
	}
	var b strings.Builder
	for i := 0; i < len(t); i++ {
		if t[i] != '~' {
			b.WriteByte(t[i])
			continue
		}
		if i+1 == len(t) {
			return "", fmt.Errorf("~ at the end of %q", t)
		}
		switch t[i+1] {
		case '0':
			b.WriteByte('~')
		case '1':
			b.WriteByte('/')
		default:
			return "", fmt.Errorf("~%c in %q", t[i+1], t)
		}
		i++
	}
	return b.String(), nil
}

// EscapeToken escapes one reference token for a JSON Pointer.
func EscapeToken(t string) string {
	return strings.ReplaceAll(strings.ReplaceAll(t, "~", "~0"), "/", "~1")
}

// String writes the Pointer back: "/" for the whole document, as A2UI
// writes it.
func (p Pointer) String() string {
	if len(p) == 0 {
		return "/"
	}
	var b strings.Builder
	for _, t := range p {
		b.WriteByte('/')
		b.WriteString(EscapeToken(t))
	}
	return b.String()
}

// Child is p with one more token.
func (p Pointer) Child(token string) Pointer {
	c := make(Pointer, len(p), len(p)+1)
	copy(c, p)
	return append(c, token)
}

// HasPrefix reports whether p is q or lies under it.
func (p Pointer) HasPrefix(q Pointer) bool {
	if len(q) > len(p) {
		return false
	}
	for i := range q {
		if p[i] != q[i] {
			return false
		}
	}
	return true
}

// Join resolves path in a scope (SPEC §"Path resolution & scope"): an
// absolute path stands alone, and a relative one continues the scope, an
// absolute path itself ("" or "/" for the root scope).
func Join(scope, path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	if path == "" {
		if scope == "" {
			return "/"
		}
		return scope
	}
	if scope == "" || scope == "/" {
		return "/" + path
	}
	return strings.TrimSuffix(scope, "/") + "/" + path
}
