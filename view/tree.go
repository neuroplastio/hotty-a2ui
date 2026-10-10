package view

import (
	"slices"
	"strconv"

	"github.com/sahilm/fuzzy"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// A HottyTree (profile §6.14) in the view: its nodes in pre-order, an id
// each, those that show, and the rows a height shows. What shows is the
// view's, so every rendition shows the same rows.

// TreeNode is one of a HottyTree's nodes as the view has it: its label
// and icon; its depth (0 for a root); its parent's index (-1 for a root);
// how many children it has (a branch has some); whether it takes children,
// a moved node (it has a list of them, empty or not: a folder); whether it
// is the last of its siblings that show, for the guides; and whether it is
// open: unfolded by the user or the agent (expanded), or because the
// selection is inside it. Href is where it leads, a page's path or a
// #fragment (profile §6.14): the agent's to follow, a docs site's nav.
type TreeNode struct {
	Label  string `json:"label"`
	Icon   string `json:"icon,omitempty"`
	Href   string `json:"href,omitempty"`
	Level  int    `json:"level,omitempty"`
	Parent int    `json:"parent"`
	Kids   int    `json:"kids,omitempty"`
	Takes  bool   `json:"takes,omitempty"`
	Last   bool   `json:"last,omitempty"`
	Open   bool   `json:"open,omitempty"`
}

// Branch reports whether a node has children.
func (n TreeNode) Branch() bool { return n.Kids > 0 }

// mapTree makes a HottyTree's element: its nodes, an id each (its value,
// as text, else its place, "0.2.1"), those that show and where they
// matched the filter, and the rows a height shows: where the wheel left
// them (ScrollTree), or when the selection moved, moved as little as
// brings the selected node into view.
func mapTree(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: Tree, Value: a2ui.ToString(b.Value(n, "selected")), Placeholder: b.String(n, "emptyText"),
		Query: Query{Text: b.String(n, "filter")}}
	if _, set := n.Props["emptyText"]; !set {
		e.Placeholder = "Nothing here."
	}
	if bd, ok := n.Props["expanded"].(a2ui.Bound); !ok || !bd.Writable() {
		if open, ok := b.St.Open[n.Key]; ok {
			e.Expanded = open
		}
	}
	if e.Expanded == nil {
		e.Expanded = strs(b.Raw(n, "expanded"))
	}
	e.Movable, e.ItemType = b.Bool(n, "reorderable"), b.String(n, "dragType")
	e.addNodes(b.List(n, "items"), -1, "")
	open := map[string]bool{}
	for _, id := range e.Expanded {
		open[id] = true
	}
	for i := range e.Nodes {
		e.Nodes[i].Open = e.Nodes[i].Branch() && open[e.RowIDs[i]]
	}
	if sel := e.SelectedRow(); sel >= 0 {
		for p := e.Nodes[sel].Parent; p >= 0; p = e.Nodes[p].Parent {
			e.Nodes[p].Open = true
		}
	}
	e.Shown, e.Matched = e.treeShown()
	later := map[int]bool{} // parents with a sibling shown after
	for k := len(e.Shown) - 1; k >= 0; k-- {
		i := e.Shown[k]
		p := e.Nodes[i].Parent
		e.Nodes[i].Last = !later[p]
		later[p] = true
	}
	if h := a2ui.ToNumber(b.Raw(n, "height")); h >= 1 {
		e.Height = int(h)
		top := b.St.Scroll[n.Key]
		if at := e.SelectedShown(); at >= 0 && b.St.Revealed[n.Key] != e.RowIDs[e.Shown[at]] {
			top = min(top, at)
			top = max(top, at-e.Height+1)
			b.St.Revealed[n.Key] = e.RowIDs[e.Shown[at]]
		}
		e.Top = max(min(top, len(e.Shown)-e.Height), 0)
		b.St.Scroll[n.Key] = e.Top
	}
	return e
}

// addNodes adds items, the children of node parent (-1 for the roots),
// and theirs, in pre-order; place is the parent's place ("0.2").
func (e *Element) addNodes(items []any, parent int, place string) {
	level := 0
	if parent >= 0 {
		level = e.Nodes[parent].Level + 1
	}
	for k, it := range items {
		m, _ := it.(map[string]any)
		at := strconv.Itoa(k)
		if place != "" {
			at = place + "." + at
		}
		id := at
		if v, ok := m["value"]; ok && v != nil {
			id = a2ui.ToString(v)
		}
		kids, takes := m["children"].([]any)
		i := len(e.Nodes)
		e.Nodes = append(e.Nodes, TreeNode{Label: a2ui.ToString(m["label"]), Icon: a2ui.ToString(m["icon"]), Href: a2ui.ToString(m["href"]),
			Level: level, Parent: parent, Kids: len(kids), Takes: takes})
		e.RowIDs = append(e.RowIDs, id)
		e.addNodes(kids, i, at)
	}
}

// treeShown are the nodes a HottyTree shows, in order, and where each
// matched its filter. Without a filter: the roots, and the children of
// every open node that shows. With one: each node whose label holds the
// filter's characters in order (sahilm/fuzzy, as a HottyList's), with its
// ancestors, which lead to it, and its descendants, which it holds; every
// branch among them shows open.
func (e *Element) treeShown() (shown []int, matched [][]int) {
	if e.Query.Text == "" {
		for i, n := range e.Nodes {
			if e.visible(i, n) {
				shown = append(shown, i)
			}
		}
		return shown, nil
	}
	labels := make([]string, len(e.Nodes))
	for i, n := range e.Nodes {
		labels[i] = n.Label
	}
	hits := map[int][]int{}
	for _, m := range fuzzy.Find(e.Query.Text, labels) {
		hits[m.Index] = m.MatchedIndexes
	}
	keep := make([]bool, len(e.Nodes))
	for i := range e.Nodes {
		if _, ok := hits[i]; !ok {
			continue
		}
		keep[i] = true
		for p := e.Nodes[i].Parent; p >= 0; p = e.Nodes[p].Parent {
			keep[p] = true
		}
		for d := i + 1; d < len(e.Nodes) && e.Nodes[d].Level > e.Nodes[i].Level; d++ {
			keep[d] = true
		}
	}
	for i := range e.Nodes {
		if keep[i] {
			shown = append(shown, i)
			matched = append(matched, hits[i])
		}
	}
	return shown, matched
}

// visible reports whether a node shows without a filter: every node
// above it is open.
func (e *Element) visible(i int, n TreeNode) bool {
	for p := n.Parent; p >= 0; p = e.Nodes[p].Parent {
		if !e.Nodes[p].Open {
			return false
		}
	}
	return true
}

// Filtering reports whether a HottyTree's filter applies: then every
// branch that shows is drawn open, whatever its Open.
// Flat reports whether a HottyTree has no branches, not even one with
// nothing in it yet: a list of pages, as a docs site's nav groups are,
// which a rendition draws with no fold column.
func (e *Element) Flat() bool {
	return !slices.ContainsFunc(e.Nodes, func(n TreeNode) bool { return n.Branch() || n.Takes })
}

func (e *Element) Filtering() bool { return e.Kind == Tree && e.Query.Text != "" }

// Guides are what a HottyTree's guides draw left of node i, a level at a
// time from the roots' children down: for each ancestor below the roots,
// whether a sibling of it follows (a line goes on down), then whether node
// i is the last of its siblings. A root has none.
func (e *Element) Guides(i int) (through []bool, last bool) {
	n := e.Nodes[i]
	if n.Level == 0 {
		return nil, n.Last
	}
	through = make([]bool, n.Level-1)
	for p, l := n.Parent, n.Level-2; p >= 0 && l >= 0; p, l = e.Nodes[p].Parent, l-1 {
		through[l] = !e.Nodes[p].Last
	}
	return through, n.Last
}

// TreeKey works a HottyTree by a key, named as SPEC §10.4 has it: the
// arrows up and down, k and j, move to the node above or below; the arrow
// right and l open a closed branch, else go to its first child; the arrow
// left and h close an open branch, else go to the parent; Page Up and Page
// Down move by the rows a height shows; Home and g go to the first node,
// End and G to the last; Enter and Space open or close a branch and act on
// a leaf (Activate). A move from no selection selects the first node
// shown. ok reports whether the tree took the key.
func (c *Controller) TreeKey(id, key string) (ok bool, err error) {
	e := c.V.Find(id)
	if e == nil || e.Kind != Tree {
		return false, nil
	}
	cur, last := e.SelectedShown(), len(e.Shown)-1
	page := e.Height
	if page <= 0 {
		page = len(e.Shown)
	}
	to := cur
	switch key {
	case "Enter", "Space":
		return true, c.Activate(id)
	case "ArrowUp", "k":
		to = cur - 1
	case "ArrowDown", "j":
		to = cur + 1
	case "PageUp":
		to = cur - page
	case "PageDown":
		to = cur + page
	case "Home", "g":
		to = 0
	case "End", "G":
		to = last
	case "ArrowRight", "l":
		if cur < 0 {
			break
		}
		i := e.Shown[cur]
		if n := e.Nodes[i]; n.Branch() && !n.Open && !e.Filtering() {
			return true, c.ToggleNode(id, i)
		}
		if cur < last && e.Nodes[e.Shown[cur+1]].Parent == i {
			to = cur + 1
		}
	case "ArrowLeft", "h":
		if cur < 0 {
			break
		}
		i := e.Shown[cur]
		if n := e.Nodes[i]; n.Branch() && n.Open && !e.Filtering() {
			return true, c.ToggleNode(id, i)
		}
		if p := e.Nodes[i].Parent; p >= 0 {
			return true, c.SelectNode(id, p)
		}
	default:
		return false, nil
	}
	if last < 0 {
		return true, nil
	}
	if cur < 0 {
		to = 0
	}
	return true, c.SelectNode(id, e.Shown[min(max(to, 0), last)])
}

// SelectNode selects a HottyTree's node i: its id goes to where selected
// is bound, else to the renderer's state. What is open now stays open
// (expanded), the branches the old selection opened among it, so that
// moving out of a branch does not close it.
func (c *Controller) SelectNode(id string, i int) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != Tree || i < 0 || i >= len(e.Nodes) {
		return nil
	}
	err := c.setProp(e, "expanded", e.openIDs(-1, -1))
	if err2 := c.set(e, e.RowIDs[i]); err == nil {
		err = err2
	}
	c.Rebuild()
	return err
}

// ToggleNode opens a HottyTree's closed branch i, or closes an open one.
// Closing the branch the selection is in selects the branch.
func (c *Controller) ToggleNode(id string, i int) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != Tree || i < 0 || i >= len(e.Nodes) || !e.Nodes[i].Branch() {
		return nil
	}
	open := e.openIDs(-1, i)
	if e.Nodes[i].Open {
		open = e.openIDs(i, -1)
	}
	err := c.setProp(e, "expanded", open)
	if sel := e.SelectedRow(); e.Nodes[i].Open && sel >= 0 && e.inside(sel, i) {
		if err2 := c.set(e, e.RowIDs[i]); err == nil {
			err = err2
		}
	}
	c.Rebuild()
	return err
}

// ScrollTree moves the rows a HottyTree with a height shows by rows (up
// when negative), as the wheel does, clamped, leaving the selection where
// it is; moved reports whether they moved.
func (c *Controller) ScrollTree(id string, rows int) (moved bool) {
	e := c.V.Find(id)
	if e == nil || e.Kind != Tree || e.Height <= 0 {
		return false
	}
	top := max(min(e.Top+rows, len(e.Shown)-e.Height), 0)
	if top == e.Top {
		return false
	}
	c.St.Scroll[id] = top
	c.Rebuild()
	return true
}

// ClickNode is a click on a HottyTree's node i, after the tree took the
// keyboard: on a branch, it selects it and opens or closes it (while no
// filter shows every branch open); on a leaf, it selects it, or acts on it
// when it was selected already (Activate). i < 0, a click beside the
// nodes, does nothing more.
func (c *Controller) ClickNode(id string, i int) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != Tree || i < 0 || i >= len(e.Nodes) {
		return nil
	}
	switch {
	case e.Nodes[i].Branch():
		if err := c.SelectNode(id, i); err != nil || e.Filtering() {
			return err
		}
		return c.ToggleNode(id, i)
	case i == e.SelectedRow():
		return c.Activate(id)
	}
	return c.SelectNode(id, i)
}

// FoldAll opens every branch of a HottyTree (open), or closes them all,
// as hottyExpandAll and hottyCollapseAll do. Closing them selects the root
// the selection is in, as closing the branch it is in would.
func (c *Controller) FoldAll(id string, open bool) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != Tree {
		return nil
	}
	ids := []string{}
	if open {
		for i, n := range e.Nodes {
			if n.Branch() {
				ids = append(ids, e.RowIDs[i])
			}
		}
	}
	err := c.setProp(e, "expanded", ids)
	if sel := e.SelectedRow(); !open && sel >= 0 && e.Nodes[sel].Parent >= 0 {
		root := sel
		for e.Nodes[root].Parent >= 0 {
			root = e.Nodes[root].Parent
		}
		if err2 := c.set(e, e.RowIDs[root]); err == nil {
			err = err2
		}
	}
	c.Rebuild()
	return err
}

// openIDs are the ids of a HottyTree's open nodes, in order, with node
// add, and without node drop and those inside it (which the selection can
// no longer open); -1 for neither.
func (e *Element) openIDs(drop, add int) []string {
	open := []string{}
	for i, n := range e.Nodes {
		if (n.Open || i == add) && (drop < 0 || i != drop && !e.inside(i, drop)) {
			open = append(open, e.RowIDs[i])
		}
	}
	return open
}

// inside reports whether node i is below node branch.
func (e *Element) inside(i, branch int) bool {
	for p := e.Nodes[i].Parent; p >= 0; p = e.Nodes[p].Parent {
		if p == branch {
			return true
		}
	}
	return false
}

// activateTree is Enter or a second click on a HottyTree: its selected
// branch opens or closes, and its selected leaf, if it shows, is acted on
// (onActivate).
func (c *Controller) activateTree(e *Element) error {
	sel := e.SelectedRow()
	if sel < 0 || !slices.Contains(e.Shown, sel) {
		return nil
	}
	if e.Nodes[sel].Branch() {
		if e.Filtering() {
			return nil
		}
		return c.ToggleNode(e.ID, sel)
	}
	if n := c.V.Node(e.ID); n.Props["onActivate"] != nil {
		return c.invoke(n, "onActivate", true)
	}
	return nil
}

// strs is a list of strings from a value: a list's items as text, nil
// for anything else.
func strs(v any) []string {
	l, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(l))
	for _, x := range l {
		out = append(out, a2ui.ToString(x))
	}
	return out
}
