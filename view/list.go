package view

import (
	"sort"
	"strconv"

	"github.com/sahilm/fuzzy"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// A HottyList (profile §6.9) in the view: its items, those the filter
// leaves in the order it ranks them, and the page that shows. The page
// and the filter are the view's, so every rendition shows the same items.

// Entry is one of a HottyList's items: its label, and the line under
// it.
type Entry struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// Query is a HottyList's filter as the user types it: its text, and
// whether the user is typing it now (Editing) or it applies (a text, not
// editing). The empty query shows every item.
type Query struct {
	Text    string `json:"text,omitempty"`
	Editing bool   `json:"editing,omitempty"`
}

// mapList makes a HottyList's element: its items, an id each (its value,
// as text, else its index), the items the query leaves and where they
// matched it, and the page that shows the selected item.
func mapList(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: RichList, Label: b.String(n, "title"), Value: a2ui.ToString(b.Value(n, "selected")),
		Filter: b.Bool(n, "filterable"), Placeholder: b.String(n, "emptyText")}
	if _, set := n.Props["emptyText"]; !set {
		e.Placeholder = "No items."
	}
	e.Movable, e.ItemType = b.Bool(n, "reorderable"), b.String(n, "dragType")
	items := b.List(n, "items")
	labels := make([]string, 0, len(items))
	for i, it := range items {
		m, _ := it.(map[string]any)
		item := Entry{Label: a2ui.ToString(m["label"]), Description: a2ui.ToString(m["description"])}
		id := strconv.Itoa(i)
		if v, ok := m["value"]; ok && v != nil {
			id = a2ui.ToString(v)
		}
		e.Items = append(e.Items, item)
		e.RowIDs = append(e.RowIDs, id)
		labels = append(labels, item.Label)
	}
	if e.Filter {
		e.Query = b.St.Query[n.Key]
	}
	e.Shown, e.Matched = filterItems(e.Query.Text, labels)
	if h := a2ui.ToNumber(b.Raw(n, "height")); h >= 1 {
		e.Height = int(h)
		top := b.St.Scroll[n.Key]
		if at := e.SelectedShown(); at >= 0 {
			top = at
		}
		top = min(top, len(e.Shown)-1)
		e.Top = max(top-top%e.Height, 0)
		b.St.Scroll[n.Key] = e.Top
	}
	return e
}

// filterItems are the items a query leaves, by index, and where each
// matched: every item in order for the empty query, else those whose
// label holds the query's characters in order, best first, as bubbles'
// list ranks them (sahilm/fuzzy, MIT). Matched is a label's matched
// characters, by their byte offsets.
func filterItems(query string, labels []string) (shown []int, matched [][]int) {
	if query == "" {
		shown = make([]int, len(labels))
		for i := range labels {
			shown[i] = i
		}
		return shown, nil
	}
	ms := fuzzy.Find(query, labels)
	sort.Stable(ms)
	shown, matched = make([]int, len(ms)), make([][]int, len(ms))
	for i, m := range ms {
		shown[i], matched[i] = m.Index, m.MatchedIndexes
	}
	return shown, matched
}

// SelectedShown is where a HottyList's selected item is among those it
// shows, or -1: none is selected, or the filter leaves it out.
func (e *Element) SelectedShown() int {
	sel := e.SelectedRow()
	for at, i := range e.Shown {
		if i == sel {
			return at
		}
	}
	return -1
}

// Pages is how many pages a HottyList's shown items take, and the one
// that shows: 1 and 0 without a height.
func (e *Element) Pages() (n, at int) {
	if e.Height <= 0 {
		return 1, 0
	}
	return max((len(e.Shown)+e.Height-1)/e.Height, 1), e.Top / e.Height
}

// ListStatus is a HottyList's status line, as bubbles' list has it: how
// many items show, the query that leaves them, and how many it leaves
// out.
func (e *Element) ListStatus() string {
	n, total := len(e.Shown), len(e.Items)
	var s string
	switch {
	case total == 0:
		s = "No items"
	case n == 0:
		s = "Nothing matched"
	case n == 1:
		s = "1 item"
	default:
		s = strconv.Itoa(n) + " items"
	}
	if e.Query.Text != "" && !e.Query.Editing {
		s = "“" + e.Query.Text + "” " + s
	}
	if out := total - n; out > 0 {
		s += " • " + strconv.Itoa(out) + " filtered"
	}
	return s
}

// ListKey works a HottyList by a key, named as SPEC §10.4 has it, as
// bubbles' list takes them. While its filter is typed: a character adds
// to it, Backspace takes the last off, Enter applies it, and Escape drops
// it; the arrows up and down still move. Otherwise: the arrows up and
// down, j and k, move a row, Page Up and Page Down, the arrows left and
// right, h and l, a page, Home and g to the first item, End and G to the
// last; Enter acts on the selected one; / starts a filter on a list with
// filterable, and Escape drops an applied one. A move from no selection
// selects the first item shown. ok reports whether the list took the key.
func (c *Controller) ListKey(id, key string) (ok bool, err error) {
	e := c.V.Find(id)
	if e == nil || e.Kind != RichList {
		return false, nil
	}
	q := e.Query
	if q.Editing {
		switch key {
		case "Enter":
			// A filter that leaves nothing is dropped, as bubbles' is.
			if len(e.Shown) == 0 {
				q.Text = ""
			}
			return true, c.setQuery(e, Query{Text: q.Text})
		case "Escape":
			return true, c.setQuery(e, Query{})
		case "Backspace":
			if r := []rune(q.Text); len(r) > 0 {
				q.Text = string(r[:len(r)-1])
			}
			return true, c.setQuery(e, q)
		case "Space":
			q.Text += " "
			return true, c.setQuery(e, q)
		case "ArrowUp", "ArrowDown":
		default:
			if !typed(key) {
				return false, nil
			}
			q.Text += key
			return true, c.setQuery(e, q)
		}
	}
	cur, last := e.SelectedShown(), len(e.Shown)-1
	page := e.Height
	if page <= 0 {
		page = len(e.Shown)
	}
	to := cur
	switch key {
	case "Enter":
		return true, c.Activate(id)
	case "/":
		if !e.Filter {
			return false, nil
		}
		return true, c.setQuery(e, Query{Text: q.Text, Editing: true})
	case "Escape":
		if q.Text == "" {
			return false, nil
		}
		return true, c.setQuery(e, Query{})
	case "ArrowUp", "k":
		to = cur - 1
	case "ArrowDown", "j":
		to = cur + 1
	case "PageUp", "ArrowLeft", "h":
		// To the same place a page back, as bubbles turns a page; none
		// before the first.
		if to = cur - page; cur < page {
			to = cur
		}
	case "PageDown", "ArrowRight", "l":
		if pages, at := e.Pages(); at < pages-1 {
			to = min(cur+page, last)
		}
	case "Home", "g":
		to = 0
	case "End", "G":
		to = last
	default:
		return false, nil
	}
	if last < 0 {
		return true, nil
	}
	if cur < 0 {
		to = 0
	}
	return true, c.SelectRow(id, e.Shown[min(max(to, 0), last)])
}

// typed reports whether a key names one character the filter takes: a
// key value of one rune that is not a named key.
func typed(key string) bool {
	r := []rune(key)
	return len(r) == 1 && r[0] >= ' ' && r[0] != 0x7f
}

// setQuery sets a HottyList's filter, and selects the first item it
// leaves, as bubbles' list does when its filter changes: the page goes
// back to the first.
func (c *Controller) setQuery(e *Element, q Query) error {
	if c.St.Query == nil {
		c.St.Query = map[string]Query{}
	}
	changed := c.St.Query[e.ID].Text != q.Text
	c.St.Query[e.ID] = q
	c.Rebuild()
	if !changed {
		return nil
	}
	c.St.Scroll[e.ID] = 0
	if e = c.V.Find(e.ID); e != nil && len(e.Shown) > 0 {
		return c.SelectRow(e.ID, e.Shown[0])
	}
	c.Rebuild()
	return nil
}
