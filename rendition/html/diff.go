package html

import "github.com/neuroplastio/hotty-go"

// diff appends the deltas that make the host's element o into n, in the
// surface named name (SPEC §6), as small as the ids allow: attr and unattr for attributes, text for
// an element whose content is one text, a recursion into children that
// keep their ids and order, and inner, which morphs the children, for any
// other change below. o and n have the same id.
func diff(name string, o, n *node, out []string) []string {
	if o.tag != n.tag || o.id != n.id {
		return append(out, hotty.MorphTo(name, o.id, n.html()))
	}
	for _, a := range o.attrs {
		if _, ok := n.attr(a.k); !ok {
			out = append(out, hotty.RemoveAttr(name, n.id, a.k))
		}
	}
	for _, a := range n.attrs {
		if v, ok := o.attr(a.k); !ok || v != a.v {
			out = append(out, hotty.SetAttr(name, n.id, a.k, a.v))
		}
	}
	switch {
	case o.raw || n.raw:
		if o.raw != n.raw || o.text != n.text {
			out = append(out, inner(name, n))
		}
	case textOnly(o) && textOnly(n):
		if textOf(o) != textOf(n) {
			out = append(out, hotty.SetText(name, n.id, textOf(n)))
		}
	case sameShape(o, n):
		for i, k := range n.kids {
			if k.id != "" {
				out = diff(name, o.kids[i], k, out)
			}
		}
	default:
		out = append(out, inner(name, n))
	}
	return out
}

func inner(name string, n *node) string {
	return hotty.Delta(name, hotty.OpInner, n.id, "", []byte(n.innerHTML()))
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
