package cells

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyTree in cells (profile §3.3, §3.4, §3.7): a row a node, as
// lipgloss's tree draws one, its guides in border: "├── " before a node a
// sibling follows, "└── " before the last, "│   " down past a level whose
// line goes on. A branch has "▶ " closed or "▼ " open before its label, as bubbles' tree has,
// and closed, how many nodes it holds after it, in muted; a root leaf has
// two blank columns there, so that the roots line up. A node's icon is a
// host's: the guides and the markers say what a tree is in cells, and few
// icons have a glyph of one column. Every row starts two columns in; the
// selected node's two columns are a "│" bar, as a HottyList's.

// treeIndent is the columns before every row of a HottyTree: the selected
// node's bar and a space.
const treeIndent = 2

// treeGuide is the columns a level of guides takes.
const treeGuide = 4

// nodeParts are what a HottyTree's node i draws after its indent: its
// guides, its marker (a branch's fold, a root leaf's blank), its label
// (matched, the filter's matches underlined; st, the label's style) and
// its count.
func nodeParts(e *view.Element, i int, matched []int, st style) (guides, marker, label, count []glyph) {
	n := e.Nodes[i]
	through, last := e.Guides(i)
	if n.Level > 0 {
		var s string
		for _, on := range through {
			if on {
				s += "│   "
			} else {
				s += "    "
			}
		}
		if last {
			s += "└── "
		} else {
			s += "├── "
		}
		guides = line(s, style{role: Border})
	}
	open := n.Open || e.Filtering()
	switch {
	case n.Branch() && open:
		marker = line("▼ ", style{role: Muted})
	case n.Branch() || n.Takes:
		// A branch with nothing in it yet shows as one, closed, with 0.
		marker = line("▶ ", style{role: Muted})
		count = line(" "+strconv.Itoa(n.Kids), style{role: Muted})
	case n.Level == 0:
		marker = line("  ", style{})
	}
	return guides, marker, marked(n.Label, matched, st), count
}

// treeWidth is a HottyTree's natural width: its widest node, whether it
// shows or not, so that it keeps its width as branches open and close.
func treeWidth(e *view.Element) int {
	n := Width(e.Placeholder)
	for i := range e.Nodes {
		g, m, l, _ := nodeParts(e, i, nil, style{})
		w := width(g) + width(m) + width(l)
		if n := e.Nodes[i]; n.Branch() || n.Takes {
			w += 1 + len(strconv.Itoa(n.Kids))
		}
		n = max(n, w)
	}
	return treeIndent + n
}

// treeMinimum is a HottyTree's minimum: its indent, a level of guides and
// a few columns.
func treeMinimum(e *view.Element) int { return treeIndent + treeGuide + 3 + 1 }

// treeHeight is a HottyTree's height: its height's worth of rows, else a
// row a node it shows; at least a row, for the empty text.
func treeHeight(e *view.Element) int {
	if e.Height > 0 {
		return e.Height
	}
	return max(len(e.Shown), 1)
}

// paintTree paints a HottyTree at (x, y), w wide. The selected node's bar
// and label are in the accent while the tree has the keyboard, the label
// bold, as bubbles' tree has it; otherwise its bar is in muted and its
// label bold, so that the node it is on (a docs nav's page) still shows
// beside the guides. A label's characters that matched the filter are
// underlined; a row cut short ends in "…". The selected node's row is
// what to keep in sight (Rendition.Sight), when it shows.
func (l *layout) paintTree(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	focused := r.focused(e.ID)
	in := x + treeIndent
	iw := max(w-treeIndent, 0)
	end := len(e.Shown)
	if e.Height > 0 {
		end = min(e.Top+e.Height, end)
	}
	if len(e.Shown) == 0 {
		cv.write(in, y, iw, fit(line(e.Placeholder, style{role: Muted}), iw))
		r.hits = append(r.hits, hit{x: x, y: y, w: w, h: 1, id: e.ID, opt: -1})
	}
	sel := e.SelectedRow()
	for k := e.Top; k < end; k++ {
		i := e.Shown[k]
		row := y + k - e.Top
		st, bar := style{}, style{}
		if i == sel {
			st, bar = style{attr: Bold}, style{role: Muted}
			if focused {
				st, bar = style{role: Accent, attr: Bold}, style{role: Accent}
			}
			cv.write(x, row, w, line("│", bar))
		}
		var match []int
		if k < len(e.Matched) {
			match = e.Matched[k]
		}
		cv.write(in, row, iw, fit(concat(nodeParts(e, i, match, st)), iw))
		r.hits = append(r.hits, hit{x: x, y: row, w: w, h: 1, id: e.ID, opt: i})
		if i == sel {
			r.reveal[e.ID] = box{x, row, w, 1}
		}
	}
	r.boxes[e.ID] = box{x, y, w, treeHeight(e)}
}
