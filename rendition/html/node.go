package html

import (
	"slices"
	"strings"
)

// node is an element of the document the rendition makes, or a text: a
// small tree it can write out as HTML and compare with the one it sent.
type node struct {
	tag   string
	id    string
	attrs []attr
	kids  []*node
	// text is a text node's data (tag ""), or, with raw set, HTML written
	// as it is (goldmark's, for a Text): the element's whole content.
	text string
	raw  bool
}

type attr struct{ k, v string }

func el(tag string, attrs ...string) *node {
	n := &node{tag: tag}
	for i := 0; i+1 < len(attrs); i += 2 {
		n.set(attrs[i], attrs[i+1])
	}
	return n
}

func txt(s string) *node { return &node{text: s} }

// set sets an attribute; "id" is the node's id.
func (n *node) set(k, v string) *node {
	if k == "id" {
		n.id = v
		return n
	}
	for i := range n.attrs {
		if n.attrs[i].k == k {
			n.attrs[i].v = v
			return n
		}
	}
	n.attrs = append(n.attrs, attr{k, v})
	return n
}

// unset removes an attribute.
func (n *node) unset(k string) {
	n.attrs = slices.DeleteFunc(n.attrs, func(a attr) bool { return a.k == k })
}

// flag sets a boolean attribute when on.
func (n *node) flag(k string, on bool) *node {
	if on {
		n.set(k, "")
	}
	return n
}

func (n *node) add(kids ...*node) *node {
	for _, k := range kids {
		if k != nil {
			n.kids = append(n.kids, k)
		}
	}
	return n
}

// find is the node with that id in n's tree, n included; nil if none is.
func (n *node) find(id string) *node {
	if n.id == id {
		return n
	}
	for _, k := range n.kids {
		if f := k.find(id); f != nil {
			return f
		}
	}
	return nil
}

func (n *node) attr(k string) (string, bool) {
	for _, a := range n.attrs {
		if a.k == k {
			return a.v, true
		}
	}
	return "", false
}

// void are the elements with no end tag.
var void = map[string]bool{"input": true, "img": true, "hr": true, "br": true, "meta": true}

// html writes the node out.
func (n *node) html() string {
	var b strings.Builder
	n.write(&b)
	return b.String()
}

func (n *node) write(b *strings.Builder) {
	if n.tag == "" {
		b.WriteString(escape(n.text, false))
		return
	}
	b.WriteString("<" + n.tag)
	if n.id != "" {
		b.WriteString(` id="` + escape(n.id, true) + `"`)
	}
	for _, a := range n.attrs {
		b.WriteString(" " + a.k)
		if a.v != "" {
			b.WriteString(`="` + escape(a.v, true) + `"`)
		}
	}
	b.WriteString(">")
	if void[n.tag] {
		return
	}
	n.writeInner(b)
	b.WriteString("</" + n.tag + ">")
}

// innerHTML writes the node's content out.
func (n *node) innerHTML() string {
	var b strings.Builder
	n.writeInner(&b)
	return b.String()
}

func (n *node) writeInner(b *strings.Builder) {
	if n.raw {
		b.WriteString(n.text)
	}
	// A parser drops a newline right after <textarea>.
	if n.tag == "textarea" && len(n.kids) > 0 && strings.HasPrefix(n.kids[0].text, "\n") {
		b.WriteString("\n")
	}
	for _, k := range n.kids {
		k.write(b)
	}
}

var (
	textEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	attrEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
)

func escape(s string, attr bool) string {
	if attr {
		return attrEscaper.Replace(s)
	}
	return textEscaper.Replace(s)
}

// equal compares two subtrees exactly.
func equal(a, b *node) bool {
	if a.tag != b.tag || a.id != b.id || a.text != b.text || a.raw != b.raw || !slices.Equal(a.attrs, b.attrs) || len(a.kids) != len(b.kids) {
		return false
	}
	for i := range a.kids {
		if !equal(a.kids[i], b.kids[i]) {
			return false
		}
	}
	return true
}
