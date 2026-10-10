package html

import (
	"slices"
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// treeKeys are the keys a HottyTree gives the program, which moves its
// selection and opens and closes its branches (view.Controller.TreeKey):
// a HottyList's. Its letters (k, j, h, l, g, G) and Enter reach the
// program without a binding, as a focused box uses no keys (SPEC §10.2).
const treeKeys = listKeys

// tree is a HottyTree on a host (profile §2): a focusable box, as a
// HottyList is, holding the nodes the view shows (the same rows as
// cells), each a treeitem whose id is the tree's, "~q" and the node's
// index. A row is indented a step a level (--k-level) and starts with its
// fold, ▸ or ▾ before a branch and blank before a leaf, so that the
// labels of a level line up; then its icon, as an Icon draws it (blank,
// for the same reason, when another node has one and it has none), and
// its label, the filter's matches underlined. Guides are cells' (profile
// §3.4): a host has the space to indent instead.
func tree(e *view.Element) *node {
	box := el("div", "id", domID(e.ID), "class", "k-tree", "tabindex", "0", "role", "tree", "data-keys", treeKeys)
	end := len(e.Shown)
	if e.Height > 0 {
		end = min(e.Top+e.Height, end)
	}
	sel := e.SelectedRow()
	icons := slices.ContainsFunc(e.Nodes, func(n view.TreeNode) bool { return n.Icon != "" })
	for k := e.Top; k < end; k++ {
		i := e.Shown[k]
		n := e.Nodes[i]
		var match []int
		if k < len(e.Matched) {
			match = e.Matched[k]
		}
		row := el("div", "id", partID(e.ID, partNode+strconv.Itoa(i)), "class", "k-node", "role", "treeitem",
			"aria-level", strconv.Itoa(n.Level+1), "aria-selected", strconv.FormatBool(i == sel), "data-on", "click",
			"style", "--k-level: "+strconv.Itoa(n.Level))
		if i == sel {
			row.set("class", "k-node k-sel")
		}
		fold := ""
		if n.Branch() {
			open := n.Open || e.Filtering()
			row.set("aria-expanded", strconv.FormatBool(open))
			fold = "▸"
			if open {
				fold = "▾"
			}
		}
		row.add(el("span", "class", "k-node-fold", "aria-hidden", "true").add(txt(fold)))
		switch {
		case n.Icon != "":
			row.add(shape(n.Icon, "", ""))
		case icons:
			row.add(el("span", "class", "k-icon k-icon-blank", "aria-hidden", "true"))
		}
		row.add(el("span", "class", "k-node-label").add(marked(n.Label, match)...))
		box.add(row)
	}
	if len(e.Shown) == 0 {
		box.add(el("div", "class", "k-tree-empty").add(texts(e.Placeholder)...))
	}
	// A tree with a height keeps it, as in cells: hidden rows stand in for
	// the nodes it lacks.
	if e.Height > 0 {
		for range e.Height - max(end-e.Top, 1) {
			box.add(el("div", "class", "k-node k-tree-gap", "aria-hidden", "true").add(txt(" ")))
		}
	}
	return box
}
