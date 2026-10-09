package view

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/diff"
)

// HottyDiff (profile §6.13) in the view: a change as files of hunks
// (package diff), from a patch or from two texts, shown unified or split.
// It is a list of its hunks, as a Table is of its rows: one is selected,
// and onActivate acts on it. It has no height: a HottyScrollView scrolls
// a long one.

// mapDiff makes a HottyDiff's element: its patch parsed, else its old and
// new compared, as its language says, or each file's name; Variant, its
// view (unified, or split); Numbers, its lineNumbers (true by default);
// Wrap (true by default); RowIDs, its hunks' ids (HunkID); Value, the
// selected one's.
func mapDiff(b *Builder, n *a2ui.Node) *Element {
	lang := b.String(n, "language")
	e := &Element{Kind: DiffView, Variant: "unified", Value: a2ui.ToString(b.Value(n, "selected")),
		Numbers: b.Raw(n, "lineNumbers") == nil || b.Bool(n, "lineNumbers"),
		Wrap:    b.Raw(n, "wrap") == nil || b.Bool(n, "wrap")}
	if b.String(n, "view") == "split" {
		e.Variant = "split"
	}
	if p := b.String(n, "patch"); p != "" {
		e.Diff = diff.Parse(p, lang)
	} else {
		context := -1
		if v := b.Raw(n, "context"); v != nil {
			context = max(int(a2ui.ToNumber(v)), 0)
		}
		e.Diff = diff.Compare(b.String(n, "file"), b.String(n, "old"), b.String(n, "new"), lang, context)
	}
	for _, f := range e.Diff.Files {
		for _, h := range f.Hunks {
			e.RowIDs = append(e.RowIDs, HunkID(f, h))
		}
	}
	return e
}

// HunkID is what identifies a hunk to the agent, as its selected value:
// its file's name and the first line of its new side, as
// "api/handler.go:12"; the line alone for texts with no name.
func HunkID(f diff.File, h diff.Hunk) string {
	line := strconv.Itoa(h.NewStart)
	if f.Name() == "" {
		return line
	}
	return f.Name() + ":" + line
}

// DiffKey works a HottyDiff by a key, named as SPEC §10.4 has it: the
// arrows (k and j) select the hunk before or after, Home and End (g and G)
// the first and the last; from no selection, each of them selects the
// first. Enter acts on the selected hunk (Activate). ok reports whether
// the key is one of those.
func (c *Controller) DiffKey(id, key string) (ok bool, err error) {
	e := c.V.Find(id)
	if e == nil || e.Kind != DiffView || len(e.RowIDs) == 0 {
		return false, nil
	}
	cur := e.SelectedRow()
	to := cur
	switch key {
	case "Enter":
		return true, c.Activate(id)
	case "ArrowUp", "k":
		to = cur - 1
	case "ArrowDown", "j":
		to = cur + 1
	case "Home", "g":
		to = 0
	case "End", "G":
		to = len(e.RowIDs) - 1
	default:
		return false, nil
	}
	if cur < 0 {
		to = 0
	}
	return true, c.SelectRow(id, to)
}
