package storybook

import (
	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/neuroplastio/hotty-go/hottytea"

	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
)

// screen is the terminal's cells, as one frame the panes are copied into.
type screen struct{ *cells.Frame }

func newScreen(cols, rows int) screen {
	f := &cells.Frame{Cols: cols, Rows: rows, Cells: make([][]cells.Cell, rows)}
	for y := range f.Cells {
		f.Cells[y] = make([]cells.Cell, cols)
		for x := range f.Cells[y] {
			f.Cells[y][x] = cells.Cell{Text: " ", Width: 1}
		}
	}
	return screen{f}
}

// put writes s at x, y, cut at the right edge or at max columns.
func (s screen) put(x, y, max int, str string, role cells.Role, attr cells.Attr) int {
	if y < 0 || y >= s.Rows {
		return 0
	}
	end := min(s.Cols, x+max)
	start := x
	for g := graphemes.FromString(str); g.Next(); {
		c := g.Value()
		if c == "\n" {
			break
		}
		w := cells.Width(c)
		if w == 0 {
			continue
		}
		if x+w > end {
			break
		}
		s.Cells[y][x] = cells.Cell{Text: c, Width: w, Role: role, Attr: attr}
		if w == 2 {
			s.Cells[y][x+1] = cells.Cell{Width: 0, Role: role, Attr: attr}
		}
		x += w
	}
	return x - start
}

// blit copies a frame's rows from top on into r; a wide cluster cut by
// r's right edge becomes a space.
func (s screen) blit(f *cells.Frame, r hottytea.Rect, top int) {
	for y := 0; y < r.H; y++ {
		fy := top + y
		if fy < 0 || fy >= f.Rows || r.Y+y >= s.Rows {
			continue
		}
		for x := 0; x < r.W && x < f.Cols && r.X+x < s.Cols; x++ {
			c := f.Cells[fy][x]
			if c.Width == 2 && (x+1 >= r.W || r.X+x+1 >= s.Cols) {
				c = cells.Cell{Text: " ", Width: 1, Role: c.Role, Attr: c.Attr, To: c.To, Mix: c.Mix}
			}
			s.Cells[r.Y+y][r.X+x] = c
		}
	}
}

// vrule draws a vertical rule down column x.
func (s screen) vrule(x, y, h int) {
	for i := 0; i < h; i++ {
		s.put(x, y+i, 1, "│", cells.Border, 0)
	}
}

// hrule draws a horizontal rule along row y.
func (s screen) hrule(x, y, w int) {
	for i := 0; i < w; i++ {
		s.put(x+i, y, 1, "─", cells.Border, 0)
	}
}
