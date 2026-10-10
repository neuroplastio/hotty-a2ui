package main

import (
	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"

	"github.com/neuroplastio/hotty-a2ui/storybook"
)

// pixels has the pointer reported in pixels where the terminal can: it
// asks whether the terminal knows SGR-Pixels (DECRQM for mode 1016) and
// how big its cells are (XTWINOPS 16), sets the mode once both answers
// came, and turns each report back into cells, the pointer's moves with
// the primary button down into storybook.DragMsg, which also says how far
// down its cell the pointer is: a drag's line then goes by halves and
// thirds of a row (profile §6.21). Until then, and where the terminal does
// not answer, reports stay in cells.
type pixels struct {
	has    bool // the terminal knows mode 1016
	cw, ch int  // a cell's width and height in pixels
	on     bool // mode 1016 is set
}

// ask asks the terminal both questions; again after a resize, which a
// font's size changes too.
func (p *pixels) ask() tea.Cmd {
	q := ansi.WindowOp(16) // the cell's size: CSI 6 ; height ; width t
	if !p.has {
		q = ansi.RequestModeMouseExtSgrPixel + q
	}
	return tea.Raw(q)
}

// off is the sequence that resets the mode, for the program's end: a
// terminal keeps it, and the next program to ask for the mouse would get
// pixels.
func (p *pixels) off() tea.Cmd {
	if !p.on {
		return nil
	}
	return tea.Raw(ansi.ResetModeMouseExtSgrPixel)
}

// update takes the answers, nil for the message then, and turns mouse
// reports in pixels into cells.
func (p *pixels) update(msg tea.Msg) (tea.Msg, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		return msg, p.ask()
	case tea.ModeReportMsg:
		if m.Mode != ansi.ModeMouseExtSgrPixel {
			return msg, nil
		}
		p.has = !m.Value.IsNotRecognized()
		return nil, p.set()
	case uv.CellSizeEvent:
		p.cw, p.ch = m.Width, m.Height
		return nil, p.set()
	}
	if !p.on {
		return msg, nil
	}
	switch m := msg.(type) {
	case tea.MouseClickMsg:
		m.X, m.Y, _ = p.cell(m.X, m.Y)
		return m, nil
	case tea.MouseReleaseMsg:
		m.X, m.Y, _ = p.cell(m.X, m.Y)
		return m, nil
	case tea.MouseWheelMsg:
		m.X, m.Y, _ = p.cell(m.X, m.Y)
		return m, nil
	case tea.MouseMotionMsg:
		x, y, sub := p.cell(m.X, m.Y)
		if m.Button == tea.MouseLeft {
			return storybook.DragMsg{X: x, Y: y, Sub: sub}, nil
		}
		m.X, m.Y = x, y
		return m, nil
	}
	return msg, nil
}

// set sets the mode once the terminal said it knows it and how big its
// cells are.
func (p *pixels) set() tea.Cmd {
	if p.on || !p.has || p.cw <= 0 || p.ch <= 0 {
		return nil
	}
	p.on = true
	return tea.Raw(ansi.SetModeMouseExtSgrPixel)
}

// cell is the cell a pointer at pixel (x, y) is in, and how far down it,
// from 0 to 1: the middle of the pixel's row.
func (p *pixels) cell(x, y int) (col, row int, sub float64) {
	x, y = max(x, 0), max(y, 0)
	return x / p.cw, y / p.ch, (float64(y%p.ch) + 0.5) / float64(p.ch)
}
