package html

import (
	"strings"

	"github.com/neuroplastio/hotty-go"
)

// diff appends the ops that make the host's element o into n (SPEC §6), as
// small as the ids allow: attr and unattr for attributes, text for an
// element whose content is one text, a recursion into children that keep
// their ids and order, append for children added after those, and inner,
// which morphs the children, for any other change below. o and n have the
// same id.
func diff(o, n *node, out []hotty.Op) []hotty.Op {
	if o.tag != n.tag || o.id != n.id {
		return append(out, hotty.OpMorph(o.id, n.html()))
	}
	for _, a := range o.attrs {
		if _, ok := n.attr(a.k); !ok {
			out = append(out, hotty.OpUnattr(n.id, a.k))
		}
	}
	for _, a := range n.attrs {
		if v, ok := o.attr(a.k); !ok || v != a.v {
			out = append(out, hotty.OpAttr(n.id, a.k, a.v))
		}
	}
	switch {
	case o.raw || n.raw:
		if o.raw != n.raw || o.text != n.text {
			out = append(out, inner(n))
		}
	case textOnly(o) && textOnly(n):
		if textOf(o) != textOf(n) {
			out = append(out, hotty.OpText(n.id, textOf(n)))
		}
	case sameShape(o, n):
		for i, k := range n.kids {
			if k.id != "" {
				out = diff(o.kids[i], k, out)
			}
		}
	case len(n.kids) > len(o.kids) && sameShape(o, &node{kids: n.kids[:len(o.kids)]}) && n.id != "":
		// Children added after the ones it had, as a log's lines are: they
		// are appended, and the rest is diffed as it is.
		for i, k := range n.kids[:len(o.kids)] {
			if k.id != "" {
				out = diff(o.kids[i], k, out)
			}
		}
		var added strings.Builder
		for _, k := range n.kids[len(o.kids):] {
			added.WriteString(k.html())
		}
		out = append(out, hotty.OpAppend(n.id, added.String()))
	default:
		out = append(out, inner(n))
	}
	return out
}

func inner(n *node) hotty.Op {
	return hotty.OpInner(n.id, n.innerHTML())
}

// textOnly: the element holds one text, or nothing.
func textOnly(n *node) bool {
	return len(n.kids) == 0 || len(n.kids) == 1 && n.kids[0].tag == ""
}

func textOf(n *node) string {
	if len(n.kids) == 0 {
		return ""
	}
	return n.kids[0].text
}

// sameShape: the children are the same elements in the same order, by
// tag and id, and those without an id are equal.
func sameShape(o, n *node) bool {
	if len(o.kids) != len(n.kids) {
		return false
	}
	for i, k := range n.kids {
		ok := o.kids[i]
		if k.id == "" || ok.id == "" {
			if !equal(ok, k) {
				return false
			}
			continue
		}
		if k.id != ok.id || k.tag != ok.tag {
			return false
		}
	}
	return true
}
