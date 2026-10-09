package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go/hottytea"

	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/storybook"
)

// bareModel is one story's surfaces alone, in cells, on the whole screen:
// no list of stories, no pickers, no panel. It is what a reference shot
// sets beside the same content in Bubble Tea (vault KIT-REF), and a way to
// try a component's keys with nothing else around it.
//
// The surfaces are drawn one under another, a row apart. Keys go to the
// surface that has the keyboard; while none has it, Tab gives it to the
// first, and q quits. Escape that the surface does not take gives the
// keyboard back. Control+C quits.
type bareModel struct {
	run   *story.Run
	th    theme.Theme
	plain bool
	panes []*barePane
	w, h  int
	frame string
	cur   *tea.Cursor
	err   string
	// anim is how often the frame wants drawing again, 0 when nothing in
	// it moves; ticking is a tick on its way.
	anim    time.Duration
	ticking bool
}

// bareTick is the frame clock's tick (bareModel.tick).
type bareTick struct{}

type barePane struct {
	s     *story.Surface
	r     *cells.Rendition
	f     *cells.Frame
	top   int
	press bool
}

func runBare(st *story.Story, th theme.Theme) error {
	if st == nil {
		return errors.New("-bare shows one story: name it (storybook -list)")
	}
	run, err := story.Start(st)
	if err != nil {
		return err
	}
	m := &bareModel{run: run, th: th, plain: os.Getenv("NO_COLOR") != ""}
	for _, s := range run.Surfaces() {
		m.panes = append(m.panes, &barePane{s: s, r: cells.New(s.C)})
	}
	_, err = tea.NewProgram(m).Run()
	return err
}

func (m *bareModel) Init() tea.Cmd { return nil }

func (m *bareModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case bareTick:
		m.ticking = false
	case tea.KeyPressMsg:
		k := hottytea.KeyName(msg.Key())
		if k == "" {
			break
		}
		if k == "Control+c" || k == "q" && m.keyboard() == nil {
			return m, tea.Quit
		}
		m.key(k)
	case tea.MouseClickMsg:
		if p, row := m.at(msg.Y); p != nil && msg.Button == tea.MouseLeft {
			m.give(p)
			p.press = true
			m.fail(p.r.Click(msg.X, row))
		}
	case tea.MouseWheelMsg:
		if p, row := m.at(msg.Y); p != nil {
			dx, dy := storybook.WheelDelta(msg)
			p.r.Wheel(msg.X, row, dx, dy)
		}
	case tea.MouseMotionMsg:
		if p := m.keyboard(); p != nil && p.press && msg.Button == tea.MouseLeft {
			m.fail(p.r.Drag(msg.X, msg.Y-p.top))
		}
	case tea.MouseReleaseMsg:
		for _, p := range m.panes {
			if p.press {
				p.r.Release()
				p.press = false
			}
		}
	}
	m.draw()
	return m, m.tick()
}

// tick asks for the next frame while something in this one moves (a
// spinner, a progress bar without a value), on the clock, as the Book's
// Tick does.
func (m *bareModel) tick() tea.Cmd {
	if m.anim == 0 || m.ticking {
		return nil
	}
	m.ticking = true
	return tea.Every(m.anim, func(time.Time) tea.Msg { return bareTick{} })
}

// key gives a key to the surface that has the keyboard, or Tab to the
// first surface when none has it.
func (m *bareModel) key(k string) {
	p := m.keyboard()
	if p == nil {
		if (k == "Tab" || k == "Shift+Tab") && len(m.panes) > 0 {
			p = m.panes[0]
			p.s.C.FocusNext(k == "Shift+Tab")
		}
		return
	}
	ok, err := p.r.Key(k)
	m.fail(err)
	if !ok && k == "Escape" {
		p.s.C.Focus("")
	}
}

// give gives the keyboard to p's surface alone: a click lands there.
func (m *bareModel) give(p *barePane) {
	for _, o := range m.panes {
		if o != p && o.s.C.St.Keyboard {
			o.s.C.Focus("")
		}
	}
}

func (m *bareModel) keyboard() *barePane {
	for _, p := range m.panes {
		if p.s.C.St.Keyboard {
			return p
		}
	}
	return nil
}

// at is the pane drawn at a row of the screen, and the row in its frame.
func (m *bareModel) at(y int) (*barePane, int) {
	for _, p := range m.panes {
		if p.f != nil && y >= p.top && y < p.top+p.f.Rows {
			return p, y - p.top
		}
	}
	return nil, 0
}

func (m *bareModel) fail(err error) {
	if err != nil {
		m.err = "✗ " + strings.TrimPrefix(err.Error(), "a2ui: ")
	}
}

func (m *bareModel) draw() {
	if m.w == 0 {
		return
	}
	var b strings.Builder
	m.cur, m.anim = nil, 0
	top := 0
	for i, p := range m.panes {
		if i > 0 {
			b.WriteString("\n\n")
			top++
		}
		p.f, p.top = p.r.Draw(m.w), top
		if d := p.r.Animating(); d > 0 && (m.anim == 0 || d < m.anim) {
			m.anim = d
		}
		if c, row, ok := p.f.Cursor(); ok && p.s.C.St.Keyboard {
			m.cur = tea.NewCursor(c, top+row)
		}
		if m.plain {
			b.WriteString(strings.TrimRight(p.f.ANSI(false), "\n"))
		} else {
			b.WriteString(strings.TrimRight(p.f.Themed(m.th), "\n"))
		}
		top += p.f.Rows
	}
	if m.err != "" {
		fmt.Fprintf(&b, "\n\n%s", m.err)
	}
	m.frame = b.String()
}

func (m *bareModel) View() tea.View {
	v := tea.NewView(m.frame)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.Cursor = m.cur
	return v
}
