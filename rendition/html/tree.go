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
// HottyList is, holding the nodes the view shows, each a treeitem whose id
// is the tree's, "~q" and the node's index. With a height, the box is that
// many rows tall and the host scrolls it (SPEC §5.3), wheel and all, with
// every node that shows in it. The selected node is the host's focus while
// the tree has the keyboard (Rendition.focusOn), so that the host scrolls
// it into view as the selection moves; a node is focusable for that alone
// (tabindex -1), and the tree's keys reach it from the box. A row is indented a step a level (--k-level) and starts with its
// fold, ▸ or ▾ before a branch and blank before a leaf, so that the
// labels of a level line up (none in a tree with no branches, view's
// Flat); then its icon, as an Icon draws it (blank,
// for the same reason, when another node has one and it has none), and
// its label, the filter's matches underlined, and a closed branch's
// count, the nodes it holds, muted. Guides are cells' (profile §3.4): a
// host has the space to indent instead.
func (m *markup) tree(e *view.Element) *node {
	class := "k-tree"
	if m.keyboard == e.ID {
		class += " k-on"
	}
	box := el("div", "id", domID(e.ID), "class", class, "tabindex", "0", "role", "tree", "data-keys", treeKeys)
	if e.Height > 0 {
		box.set("class", class+" k-scrolls").set("style", "--k-rows: "+strconv.Itoa(e.Height))
	}
	sel := e.SelectedRow()
	icons := slices.ContainsFunc(e.Nodes, func(n view.TreeNode) bool { return n.Icon != "" })
	flat := e.Flat()
	for k := range e.Shown {
		i := e.Shown[k]
		n := e.Nodes[i]
		var match []int
		if k < len(e.Matched) {
			match = e.Matched[k]
		}
		row := el("div", "id", partID(e.ID, partNode+strconv.Itoa(i)), "class", "k-node", "role", "treeitem", "tabindex", "-1",
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
		if !flat {
			row.add(el("span", "class", "k-node-fold", "aria-hidden", "true").add(txt(fold)))
		}
		switch {
		case n.Icon != "":
			row.add(shape(n.Icon, "", 0, ""))
		case icons:
			row.add(el("span", "class", "k-icon k-icon-blank", "aria-hidden", "true"))
		}
		row.add(el("span", "class", "k-node-label").add(marked(n.Label, match)...))
		if n.Branch() && !n.Open && !e.Filtering() {
			row.add(el("span", "class", "k-node-count").add(txt(strconv.Itoa(n.Kids))))
		}
		box.add(row)
	}
	if len(e.Shown) == 0 {
		box.add(el("div", "class", "k-tree-empty").add(texts(e.Placeholder)...))
	}
	return box
}
