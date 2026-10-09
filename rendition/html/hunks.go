package html

import (
	"strconv"

	change "github.com/neuroplastio/hotty-a2ui/diff" // diff is the deltas (diff.go)
	"github.com/neuroplastio/hotty-a2ui/view"
)

// diffKeys are the keys a HottyDiff gives the program, which moves its
// selection (Rendition.Key): the arrows, Home and End, which a host would
// scroll with, and k, j, g and G, which a HottyScrollView around it binds
// to scroll actions (SPEC §10.2). Enter reaches the program anyway.
const diffKeys = "ArrowUp=program ArrowDown=program Home=program End=program k=program j=program g=program G=program"

// diffView is a HottyDiff on a host (profile §2): a focusable box in the
// terminal's font whose keys reach the program (diffKeys), holding for
// each file its name, and for each hunk a block that reports a click,
// which selects the hunk, or acts on it once selected (Event). A hunk is
// its header and its lines: unified, a row of each line's numbers, sign
// and code; split, a grid of the old side and the new, a changed line
// beside the line it replaces. A changed line is tinted toward its role,
// its changed words more so (k-w).
//
// The selected hunk is the host's focus while the diff has the keyboard
// (Rendition.focusOn), so that a host scrolls it into view (SPEC §5.3) as
// cells keeps it in sight; its block is focusable for that alone
// (tabindex -1), and the diff's keys reach it from the box (data-keys
// holds for what is inside it).
//
// A split diff stays split however narrow the host makes it: its sides
// share the width and wrap. Unlike cells, the markup does not know the
// width, and a host has no container queries.
func (m *markup) diffView(e *view.Element) *node {
	class := "k-diff"
	if !e.Wrap {
		class += " k-nowrap"
	}
	split := e.Variant == "split"
	if split {
		class += " k-split"
	}
	if m.keyboard == e.ID {
		class += " k-on"
	}
	box := el("div", "id", domID(e.ID), "class", class)
	if len(e.RowIDs) > 0 {
		box.set("tabindex", "0").set("role", "listbox").set("data-keys", diffKeys)
	}
	o, n := 0, 0
	if e.Numbers {
		for _, f := range e.Diff.Files {
			for _, h := range f.Hunks {
				for _, l := range h.Lines {
					o, n = max(o, len(strconv.Itoa(l.Old))), max(n, len(strconv.Itoa(l.New)))
				}
			}
		}
		if split {
			o, n = max(o, n), max(o, n)
		}
		box.set("style", "--k-lo: "+strconv.Itoa(o)+"ch; --k-ln: "+strconv.Itoa(n)+"ch")
	}
	sel := e.SelectedRow()
	k := 0
	for _, f := range e.Diff.Files {
		if f.Name() != "" || len(e.Diff.Files) > 1 {
			box.add(fileHead(f))
		}
		for _, h := range f.Hunks {
			if h.Skipped > 0 {
				box.add(fold(h.Skipped))
			}
			hunk := el("div", "id", partID(e.ID, partHunk+strconv.Itoa(k)), "class", "k-hunk", "tabindex", "-1",
				"role", "option", "aria-selected", strconv.FormatBool(k == sel), "data-on", "click")
			if k == sel {
				hunk.set("class", "k-hunk k-sel")
			}
			head := el("div", "class", "k-hh").add(el("span", "class", "k-hh-at").add(txt(h.Header())))
			if h.Section != "" {
				head.add(el("span", "class", "k-hh-sec").add(texts(" " + h.Section)...))
			}
			hunk.add(head)
			if split {
				hunk.add(splitLines(e, h)...)
			} else {
				for _, l := range h.Lines {
					hunk.add(diffLine(e, l, l.Old, l.New, "k-dl"))
				}
			}
			box.add(hunk)
			k++
		}
		if f.After > 0 && len(f.Hunks) > 0 {
			box.add(fold(f.After))
		}
	}
	if k == 0 && len(box.kids) == 0 {
		box.add(el("div", "class", "k-diff-note").add(txt("No changes")))
	}
	return box
}

// fold is a run of n unchanged lines the diff leaves out, as cells
// folds it: before a hunk, and after the last where the diff knows.
func fold(n int) *node {
	s := "⋯ " + strconv.Itoa(n) + " unchanged lines"
	if n == 1 {
		s = "⋯ 1 unchanged line"
	}
	return el("div", "class", "k-fold").add(txt(s))
}

// fileHead is a file's name, old → new when it was renamed, then what
// the change adds and removes, or that the file is new, deleted or binary.
func fileHead(f change.File) *node {
	name := f.Name()
	if f.Old != "" && f.New != "" && f.Old != f.New {
		name = f.Old + " → " + f.New
	}
	n := el("div", "class", "k-diff-file").add(el("span", "class", "k-diff-name").add(texts(name)...))
	stat := el("span", "class", "k-diff-stat")
	switch add, rm := f.Stat(); {
	case f.Binary:
		stat.add(txt("binary"))
	case f.Old == "" && f.New != "":
		stat.add(el("span", "class", "k-add").add(txt("new")))
	case f.New == "" && f.Old != "":
		stat.add(el("span", "class", "k-del").add(txt("deleted")))
	default:
		if add > 0 {
			stat.add(el("span", "class", "k-add").add(txt("+" + strconv.Itoa(add))))
		}
		if add > 0 && rm > 0 {
			stat.add(txt(" "))
		}
		if rm > 0 {
			stat.add(el("span", "class", "k-del").add(txt("-" + strconv.Itoa(rm))))
		}
	}
	if len(stat.kids) > 0 {
		n.add(stat)
	}
	return n
}

// diffLine is a line's row, of class class: its old and new numbers
// (aria-hidden, as a HottyCode's; 0 is none, a side without it), or one
// of them in a split's side (new < 0), its sign, then its code, each
// token coloured by its kind and the words it changes marked.
func diffLine(e *view.Element, l change.Line, old, new int, class string) *node {
	row := el("div", "class", class)
	switch l.Op {
	case change.Removed:
		row.set("class", class+" k-d-del")
	case change.Added:
		row.set("class", class+" k-d-add")
	}
	if e.Numbers {
		row.add(lineNumber(old, "k-ln k-lo"))
		if new >= 0 {
			row.add(lineNumber(new, "k-ln"))
		}
	}
	sign := el("span", "class", "k-sign")
	switch l.Op {
	case change.Removed:
		sign.set("aria-label", "removed").add(txt("-"))
	case change.Added:
		sign.set("aria-label", "added").add(txt("+"))
	}
	src := el("span", "class", "k-src")
	for _, s := range l.Segments() {
		c := view.TokenClass(s.Kind)
		switch {
		case s.Changed && c != "":
			c += " k-w"
		case s.Changed:
			c = "k-w"
		}
		if c == "" {
			src.add(texts(s.Text)...)
			continue
		}
		src.add(el("span", "class", c).add(texts(s.Text)...))
	}
	return row.add(sign, src)
}

func lineNumber(v int, class string) *node {
	n := el("span", "class", class, "aria-hidden", "true")
	if v > 0 {
		n.add(txt(strconv.Itoa(v)))
	}
	return n
}

// splitLines are a hunk's lines side by side, as cells pairs them: a
// context line on both sides, and in a run of removed lines followed by
// added ones, each removed line beside the added line that replaces it; a
// side with no line is blank (k-d-none). They are the cells of the hunk's
// grid, old then new.
func splitLines(e *view.Element, h change.Hunk) []*node {
	var out []*node
	side := func(l *change.Line, num int, class string) *node {
		if l == nil {
			return el("div", "class", class+" k-d-none")
		}
		return diffLine(e, *l, num, -1, class)
	}
	pair := func(a, b *change.Line) {
		var o, n int
		if a != nil {
			o = a.Old
		}
		if b != nil {
			n = b.New
		}
		out = append(out, side(a, o, "k-ds k-old"), side(b, n, "k-ds k-new"))
	}
	ls := h.Lines
	for i := 0; i < len(ls); {
		if ls[i].Op == change.Context {
			pair(&ls[i], &ls[i])
			i++
			continue
		}
		j := i
		for j < len(ls) && ls[j].Op == change.Removed {
			j++
		}
		k := j
		for k < len(ls) && ls[k].Op == change.Added {
			k++
		}
		for m := 0; m < max(j-i, k-j); m++ {
			var a, b *change.Line
			if i+m < j {
				a = &ls[i+m]
			}
			if j+m < k {
				b = &ls[j+m]
			}
			pair(a, b)
		}
		i = k
	}
	return out
}

// focusOn is what has the host's keyboard for the element that has the
// controller's: the element, or a HottyDiff's selected hunk (a DOM id),
// which the host then scrolls into view (diffView).
func (r *Rendition) focusOn(id string) string {
	if e := r.C.V.Find(id); e != nil && e.Kind == view.DiffView {
		if i := e.SelectedRow(); i >= 0 {
			return partID(id, partHunk+strconv.Itoa(i))
		}
	}
	return id
}
