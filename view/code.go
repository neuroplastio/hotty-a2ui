package view

import (
	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/highlight"
)

// HottyCode (profile §6.12) in the view: source code as lines of tokens
// (package highlight), which the renditions colour by role; line numbers
// and marked lines in a gutter. It has no height: a HottyScrollView
// scrolls a long one.

// Mark is what a marked line of code is: its sign and its tint.
type Mark string

// The marks a line takes.
const (
	MarkHighlight Mark = "highlight" // a line to look at
	MarkAdded     Mark = "added"     // a line a change adds
	MarkRemoved   Mark = "removed"   // a line a change removes
	MarkError     Mark = "error"     // a line with an error
	MarkWarning   Mark = "warning"   // a line with a warning
)

// mapCode makes a HottyCode's element: its code lexed as its language
// says; Numbers, its lineNumbers prop; FirstLine, its startLine (1 by
// default); Marks, each of its marks' lines (from line to end) with its
// kind, a later mark over an earlier; Wrap, its wrap prop, true by
// default.
func mapCode(b *Builder, n *a2ui.Node) *Element {
	lang := b.String(n, "language")
	e := &Element{Kind: Listing, Lang: lang, Code: highlight.Lines(b.String(n, "code"), lang), Numbers: b.Bool(n, "lineNumbers"),
		FirstLine: 1, Wrap: b.Raw(n, "wrap") == nil || b.Bool(n, "wrap")}
	if v := b.Raw(n, "startLine"); v != nil {
		e.FirstLine = int(a2ui.ToNumber(v))
	}
	marks, _ := b.Raw(n, "marks").([]any)
	for _, m := range marks {
		o, _ := m.(map[string]any)
		kind := Mark(a2ui.ToString(o["kind"]))
		switch kind {
		case MarkHighlight, MarkAdded, MarkRemoved, MarkError, MarkWarning:
		default:
			continue
		}
		from := int(a2ui.ToNumber(o["line"]))
		to := from
		if v, ok := o["end"]; ok {
			to = max(int(a2ui.ToNumber(v)), from)
		}
		// A mark past the code marks nothing; a range is cut to it.
		last := e.FirstLine + len(e.Code) - 1
		for l := max(from, e.FirstLine); l <= min(to, last); l++ {
			if e.Marks == nil {
				e.Marks = map[int]Mark{}
			}
			e.Marks[l] = kind
		}
	}
	return e
}
