package storybook

import (
	"slices"

	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// Scroll is the window a program shows a cells frame taller than its room
// through, as a host scrolls a surface (SPEC §5.3): a Book's pane, and
// storybook -bare's screen. A story whose surface is a page, a docs page
// in a HottyMarkdown (profile §6.27), scrolls in it as a browser's page.
type Scroll struct {
	// Top is the frame's first row shown.
	Top int
	// sight is what Keep last kept in sight.
	sight Sight
}

// Sight is what a Scroll keeps in its window: the element with the
// keyboard, or a follow while none has it, and its rows in the frame
// (cells.Rendition.Sight: a HottyDiff's selected hunk, the heading a
// HottyMarkdown went to); the surface's jump (view.Jump.Seq); and whether
// the element is a page's HottyMarkdown (OnPage), whose jumps go to the
// top.
type Sight struct {
	ID       string
	Row, H   int
	Keyboard bool
	Jump     int
	Page     bool
}

// SightOf is a surface's sight in its frame, ok false where nothing has
// the keyboard and there is no follow, or the frame does not show it.
func SightOf(c *view.Controller, r *cells.Rendition, follow string) (s Sight, ok bool) {
	s = Sight{Keyboard: c.St.Keyboard, Jump: c.St.Jump.Seq}
	id := follow
	if c.St.Keyboard {
		id = c.St.Focus
	}
	if id == "" {
		return s, false
	}
	_, row, _, h, ok := r.Sight(id)
	if !ok {
		return s, false
	}
	s.ID, s.Row, s.H, s.Page = id, row, h, OnPage(c, id)
	return s, true
}

// OnPage reports whether id is a HottyMarkdown that is its surface's
// page, held by no HottyScrollView, which would scroll to its jumps
// itself.
func OnPage(c *view.Controller, id string) bool {
	e := c.V.Find(id)
	return e != nil && e.Kind == view.Document && !slices.ContainsFunc(c.Ancestors(id), func(a *view.Element) bool {
		return a.Kind == view.ScrollView
	})
}

// Keep is the first row to show of a frame frameRows tall in rows rows:
// Top (where the wheel or the page keys left it), moved as little as
// brings now into sight, and to now's start where it is taller than the
// window; for a new jump on a page, to the heading, at the top, as a
// browser scrolls to a fragment. It moves only when the sight changed, as
// focus scrolls a page, so that the wheel can move on from there.
func (s *Scroll) Keep(now Sight, ok bool, rows, frameRows int) int {
	top := s.Top
	switch {
	case !ok || now == s.sight:
	case now.Page && now.Jump != s.sight.Jump:
		top = now.Row
	default:
		if now.Row+now.H > top+rows {
			top = now.Row + now.H - rows
		}
		if now.Row < top {
			top = now.Row
		}
	}
	s.sight = now
	s.Top = max(0, min(top, frameRows-rows))
	return s.Top
}

// Page scrolls a frame taller than its window as a browser scrolls a
// page, with a key the element with the keyboard left: Page Down and
// Space a page down, Page Up and Shift+Space a page up, as a
// HottyScrollView pages. Keep holds it within the frame.
func (s *Scroll) Page(k string, rows, frameRows int) bool {
	if frameRows <= rows {
		return false
	}
	switch k {
	case "PageDown", "Space":
		s.Top += rows
	case "PageUp", "Shift+Space":
		s.Top -= rows
	default:
		return false
	}
	return true
}

// Wheel scrolls it a notch's rows (wheelRows), dy notches down (up where
// less than 0), where nothing in the frame took the wheel.
func (s *Scroll) Wheel(dy int) { s.Top += dy * wheelRows }
