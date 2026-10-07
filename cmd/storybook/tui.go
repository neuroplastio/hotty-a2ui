package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytea"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/rendition/html"
	"github.com/neuroplastio/hotty-a2ui/rendition/text"
	"github.com/neuroplastio/hotty-a2ui/story"
)

// The renditions a story is shown in. Surfaces and side by side need a
// HOTTY host.
const (
	rendSurfaces = "surfaces"
	rendCells    = "cells"
	rendText     = "text"
	rendSide     = "side"
)

// kind is how a pane draws its surface.
type kind int

const (
	asSurface kind = iota // a HOTTY surface: HTML the host lays out
	asCells               // cells the storybook paints
	asText                // plain text
)

// pane is one surface on the screen, in one rendition.
type pane struct {
	name  string // the HOTTY surface's name, and the pane's key
	s     *story.Surface
	kind  kind
	html  *html.Rendition
	cells *cells.Rendition
	rect  hottytea.Rect
	// mirror: a cells pane beside the same surface on the host, which
	// shows its state and takes no input: the keyboard is in one place.
	mirror bool
	frame  *cells.Frame
	top    int // the frame's first row shown
}

func (p *pane) takesInput() bool { return p.kind != asText && !p.mirror }

type (
	streamMsg struct{ raw json.RawMessage }
	streamEnd struct{ err error }
)

// model is the interactive storybook.
type model struct {
	s      *hottytea.Session
	w, h   int
	colors bool

	ch     *chrome
	list   [][2]string
	first  string // the story to open first
	cur    string
	run    *story.Run
	stream *story.Run
	source string // where the stream comes from, for its heading
	rend   string
	gen    int
	seq    map[*a2ui.Surface]int
	out    func(a2ui.Outbound)

	panes  map[string]*pane
	order  []*pane // this frame's panes, in Tab order: nav, the story's, panel
	focus  *pane   // the pane with the keyboard, if any
	last   *pane   // the pane that had it last: Tab goes on from there
	frame  string
	cursor *tea.Cursor
	status string
}

func newModel(first string, stream bool, source string, out func(a2ui.Outbound)) *model {
	m := &model{s: hottytea.New(), colors: os.Getenv("NO_COLOR") == "", first: first, source: source,
		out: out, seq: map[*a2ui.Surface]int{}, panes: map[string]*pane{}}
	if stream {
		m.stream = story.NewRun()
		m.stream.Out = out
		if first == "" {
			m.first = streamName
		}
	}
	return m
}

func (m *model) Init() tea.Cmd { return m.s.Detect() }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	msg, cmd := m.s.Update(msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case hottytea.ReadyMsg:
		m.ready(msg.Mode)
	case hottytea.EventMsg:
		m.event(msg.Event)
	case hottytea.ErrorMsg:
		m.status = "✗ the host: " + msg.Reply.Err().Error()
	case tea.KeyPressMsg:
		if m.key(keyValue(msg.Key())) {
			return m, tea.Sequence(m.s.Close(), tea.Quit)
		}
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			m.click(msg.X, msg.Y)
		}
	case streamMsg:
		if err := m.stream.Feed(msg.raw); err != nil {
			m.status = "✗ " + strings.TrimPrefix(err.Error(), "a2ui: ")
		}
	case streamEnd:
		m.status = "the stream ended"
		if msg.err != nil {
			m.status = "✗ the stream: " + msg.err.Error()
		}
	}
	m.settle()
	return m, tea.Batch(cmd, m.draw())
}

func (m *model) View() tea.View {
	v := tea.NewView(m.frame)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.Cursor = m.cursor
	v.WindowTitle = "storybook · " + m.cur
	return v
}

// ready starts the storybook once the terminal is known: surfaces on a
// host, cells elsewhere.
func (m *model) ready(mode hottytea.Mode) {
	opts := []renditionOption{{rendCells, "Cells"}, {rendText, "Text"}}
	m.rend = rendCells
	if mode == hottytea.Native {
		opts = append([]renditionOption{{rendSurfaces, "Surfaces"}}, append(opts, renditionOption{rendSide, "Side by side"})...)
		m.rend = rendSurfaces
	}
	m.list = entries(m.stream != nil)
	m.ch = newChrome(m.list, opts, m.rend)
	first := m.first
	if first == "" {
		first = m.list[0][0]
	}
	m.open(first)
}

// open shows a story: a new run of it, or the stream as it is.
func (m *model) open(name string) {
	head := ""
	if name == streamName && m.stream != nil {
		m.run = m.stream
		head = "**The stream** · A2UI from " + m.source
	} else {
		st := story.Find(name)
		if st == nil {
			m.status = "✗ no story " + name
			return
		}
		run := story.NewRun()
		run.Out = m.out
		for _, msg := range st.Messages {
			if err := run.Feed(msg); err != nil {
				m.status = "✗ " + strings.TrimPrefix(err.Error(), "a2ui: ")
				break
			}
		}
		m.run = run
		head = "**" + st.Title + "** · " + st.Description
	}
	if m.focus != nil && m.focus.s != m.ch.surface(navID) {
		m.focus = nil
	}
	m.cur = name
	m.gen++
	m.seq = map[*a2ui.Surface]int{}
	m.ch.showing(m.list, name, head)
}

// settle does what the last message left to do: the storybook's own
// actions, the rendition picked, the panel.
func (m *model) settle() {
	if m.ch == nil {
		return
	}
	acts := m.ch.acts
	m.ch.acts = nil
	for _, a := range acts {
		if name, _ := a.Context["name"].(string); a.Name == "open" && name != "" {
			m.open(name)
		}
	}
	if r := m.ch.rendition(); r != "" && r != m.rend {
		m.rend = r
		if m.focus != nil && !m.focus.chrome(m) {
			m.focus = nil
		}
	}
	if m.focus != nil && !m.focus.s.C.St.Keyboard {
		m.focus = nil
	}
	if m.run != nil {
		m.ch.report(m.run)
	}
}

func (p *pane) chrome(m *model) bool {
	return p.s == m.ch.surface(navID) || p.s == m.ch.surface(panelID)
}

// key handles a key that reached the storybook: the pane with the
// keyboard first (on a host, only what the surface left), then the
// story's Shortcuts, then the storybook's own keys.
func (m *model) key(k string) (quit bool) {
	m.status = ""
	if k == "Control+c" || k == "Control+q" {
		return true
	}
	if m.ch == nil {
		return false
	}
	if p := m.focus; p != nil {
		switch p.kind {
		case asCells:
			had := p.s.C.St.Keyboard
			handled, err := p.cells.Key(k)
			m.fail(err)
			if handled {
				if had && !p.s.C.St.Keyboard && (k == "Tab" || k == "Shift+Tab") {
					m.focus = nil
					m.cycle(k == "Shift+Tab", p)
				}
				return false
			}
		case asSurface:
			cmds, ok, err := p.html.Key(k)
			m.s.Send(cmds...)
			m.fail(err)
			if ok {
				return false
			}
		}
	} else if m.shortcut(k) {
		return false
	}
	switch k {
	case "Tab", "Shift+Tab":
		m.cycle(k == "Shift+Tab", cmp.Or(m.focus, m.last))
	case "Escape":
		m.escape()
	case "F2":
		m.nextRendition()
	case "q":
		return m.focus == nil
	}
	return false
}

// shortcut offers a key to the story's surfaces, while no pane has the
// keyboard: the story's is the surface keys apply to.
func (m *model) shortcut(k string) bool {
	seen := map[*story.Surface]bool{}
	for _, p := range m.order {
		if !p.takesInput() || p.chrome(m) || seen[p.s] {
			continue
		}
		seen[p.s] = true
		var ok bool
		var err error
		switch p.kind {
		case asCells:
			ok, err = p.cells.Key(k)
		case asSurface:
			var cmds []string
			cmds, ok, err = p.html.Key(k)
			m.s.Send(cmds...)
		}
		m.fail(err)
		if ok {
			return true
		}
	}
	return false
}

// cycle gives the keyboard to the next pane that takes it, after from
// (or before it, back), at its first focusable element (or its last).
func (m *model) cycle(back bool, from *pane) {
	var ps []*pane
	for _, p := range m.order {
		if p.takesInput() {
			ps = append(ps, p)
		}
	}
	i := slices.Index(ps, from)
	for range ps {
		switch {
		case back && i <= 0:
			i = len(ps) - 1
		case back:
			i--
		default:
			i = (i + 1) % len(ps)
		}
		if m.give(ps[i], back) {
			return
		}
	}
}

// give gives a pane the keyboard, at its first focusable element (its
// last, back); false if it has none.
func (m *model) give(p *pane, back bool) bool {
	c := p.s.C
	c.St.Keyboard = false
	if !c.FocusNext(back) {
		return false
	}
	if f := m.focus; f != nil && f != p {
		f.s.C.St.Keyboard = false
	}
	m.focus, m.last = p, p
	return true
}

// escape is Escape that no surface took: it closes a Modal, else the
// keyboard goes back to the storybook.
func (m *model) escape() {
	for _, p := range m.order {
		if p.s.C.St.Modal != "" && (m.focus == nil || m.focus.s == p.s) {
			p.s.C.CloseModal()
			return
		}
	}
	if m.focus != nil {
		m.focus.s.C.St.Keyboard = false
		m.focus = nil
	}
}

func (m *model) nextRendition() {
	var opts []string
	for _, o := range []string{rendSurfaces, rendCells, rendText, rendSide} {
		if m.s.Mode == hottytea.Native || o == rendCells || o == rendText {
			opts = append(opts, o)
		}
	}
	next := opts[(slices.Index(opts, m.rend)+1)%len(opts)]
	delete(m.ch.sent, navID+"/rendition")
	m.ch.set(navID, "/rendition", []any{next})
}

// click is a primary click on the cells: in a cells pane, its rendition
// takes it; anywhere else the keyboard goes back to the storybook.
func (m *model) click(x, y int) {
	for _, p := range m.order {
		if p.kind != asCells || p.mirror || x < p.rect.X || x >= p.rect.X+p.rect.W || y < p.rect.Y || y >= p.rect.Y+p.rect.H {
			continue
		}
		if f := m.focus; f != nil && f != p {
			f.s.C.St.Keyboard = false
		}
		m.fail(p.cells.Click(x-p.rect.X, y-p.rect.Y+p.top))
		if p.s.C.St.Keyboard {
			m.focus = p
		} else if m.focus == p {
			m.focus = nil
		}
		return
	}
	if f := m.focus; f != nil && f.kind == asCells {
		f.s.C.St.Keyboard = false
		m.focus = nil
	}
}

// event is what the user did in a surface on the host.
func (m *model) event(ev hotty.Event) {
	p := m.panes[ev.Surface]
	if p == nil || p.html == nil {
		return
	}
	m.fail(p.html.Event(ev))
	switch ev.Kind {
	case hotty.EventFocus:
		if f := m.focus; f != nil && f != p && f.kind == asCells {
			f.s.C.St.Keyboard = false
		}
		m.focus, m.last = p, p
	case hotty.EventBlur:
		m.last = p
	}
}

func (m *model) fail(err error) {
	if err != nil {
		m.status = "✗ " + strings.TrimPrefix(err.Error(), "a2ui: ")
	}
}

// pane is the pane of a surface in a rendition, kept from frame to frame
// so its rendition keeps what it has (the document the host has, a
// field's cursor).
func (m *model) pane(name string, s *story.Surface, k kind, r hottytea.Rect) *pane {
	p := m.panes[name]
	if p == nil || p.s != s || p.kind != k {
		p = &pane{name: name, s: s, kind: k}
		switch k {
		case asSurface:
			p.html = html.New(s.C, name)
		case asCells:
			p.cells = cells.New(s.C)
		}
		m.panes[name] = p
	}
	p.rect = r
	return p
}

// draw lays the screen out, draws the cells, and says which surfaces go
// where: nav on the left; the story above panel on the right.
func (m *model) draw() tea.Cmd {
	m.cursor = nil
	if m.ch == nil || m.w == 0 {
		m.frame = "Finding out whether the terminal is a HOTTY host…"
		return nil
	}
	W, H := m.w, m.h
	if W < 40 || H < 12 {
		m.s.Layout(nil)
		m.frame = "The storybook needs 40×12 at least."
		return m.s.Flush()
	}
	scr := newScreen(W, H)
	native := m.s.Mode == hottytea.Native

	x := scr.put(0, 0, W, "HOTTY kit storybook", cells.Fg, cells.Bold)
	scr.put(x, 0, W-x, " · "+m.cur+" · "+m.rend, cells.Muted, 0)
	help := "Tab next · Esc leave · F2 rendition · Ctrl+C quit"
	if m.status != "" {
		help = m.status + " · " + help
	}
	scr.put(0, H-1, W, help, cells.Muted, 0)

	bodyH := H - 2
	navW := min(max(W/4, 22), 36, W/2)
	scr.vrule(navW, 1, bodyH)
	rx := navW + 1
	rw := W - rx
	panelH := min(max(bodyH/3, 6), 16)
	canvasH := bodyH - panelH - 1
	scr.hrule(rx, 1+canvasH, rw)

	chromeKind := asCells
	if native {
		chromeKind = asSurface
	}
	order := []*pane{m.pane(navID, m.ch.surface(navID), chromeKind, hottytea.Rect{X: 0, Y: 1, W: navW, H: bodyH})}

	kinds := map[string][]kind{rendSurfaces: {asSurface}, rendCells: {asCells}, rendText: {asText}, rendSide: {asSurface, asCells}}[m.rend]
	if !native {
		kinds = slices.DeleteFunc(slices.Clone(kinds), func(k kind) bool { return k == asSurface })
		if len(kinds) == 0 {
			kinds = []kind{asCells}
		}
	}
	var ss []*story.Surface
	if m.run != nil {
		ss = m.run.Surfaces()
	}
	colW := (rw - (len(kinds) - 1)) / len(kinds)
	for ki, k := range kinds {
		cx := rx + ki*(colW+1)
		if ki > 0 {
			scr.vrule(cx-1, 1, canvasH)
		}
		y := 1
		for i, s := range ss {
			h := canvasH / len(ss)
			if i == len(ss)-1 {
				h = canvasH - (y - 1)
			}
			if h <= 0 {
				continue
			}
			if _, ok := m.seq[s.S]; !ok {
				m.seq[s.S] = len(m.seq)
			}
			name := fmt.Sprintf("s%d-%d-%c", m.gen, m.seq[s.S], "hct"[k])
			p := m.pane(name, s, k, hottytea.Rect{X: cx, Y: y, W: colW, H: h})
			p.mirror = k == asCells && len(kinds) > 1
			order = append(order, p)
			y += h
		}
	}
	if len(ss) == 0 {
		scr.put(rx+1, 1, rw-1, "No surface yet.", cells.Muted, 0)
	}
	order = append(order, m.pane(panelID, m.ch.surface(panelID), chromeKind, hottytea.Rect{X: rx, Y: 2 + canvasH, W: rw, H: panelH}))
	m.order = order
	for name, p := range m.panes {
		if !slices.Contains(order, p) {
			delete(m.panes, name)
			if m.focus == p {
				m.focus = nil
			}
			if m.last == p {
				m.last = nil
			}
		}
	}

	var want []hottytea.Surface
	for _, p := range order {
		switch p.kind {
		case asCells:
			p.frame = p.cells.Draw(p.rect.W)
			p.top = scrollTo(p)
			scr.blit(p.frame, p.rect, p.top)
		case asText:
			lines := strings.Split(strings.TrimRight(text.Render(p.s.C.V), "\n"), "\n")
			for i, l := range lines {
				if i < p.rect.H {
					scr.put(p.rect.X, p.rect.Y+i, p.rect.W, l, cells.Fg, 0)
				}
			}
		case asSurface:
			m.s.Send(p.html.Update()...)
			want = append(want, hottytea.Surface{Name: p.name, Rect: p.rect, Doc: p.html.Doc, Scroll: hotty.ScrollVertical})
		}
	}
	m.s.Layout(want)
	// A surface that just got its document gets the keyboard now, if its
	// view has it (autofocus).
	for _, p := range order {
		if p.kind == asSurface {
			m.s.Send(p.html.Update()...)
		}
	}
	if f := m.focus; f != nil && f.kind == asCells && f.s.C.St.Keyboard && f.frame != nil {
		if c, r, ok := f.frame.Cursor(); ok && r >= f.top && r < f.top+f.rect.H {
			m.cursor = tea.NewCursor(f.rect.X+c, f.rect.Y+r-f.top)
		}
	}
	m.frame = scr.ANSI(m.colors)
	return m.s.Flush()
}

// scrollTo is the first row a cells pane shows: the one before, moved as
// little as keeps the focused element in sight.
func scrollTo(p *pane) int {
	top := p.top
	c := p.s.C
	if c.St.Keyboard && c.St.Focus != "" {
		if _, row, _, h, ok := p.cells.Box(c.St.Focus); ok {
			if row < top {
				top = row
			}
			if row+h > top+p.rect.H {
				top = row + h - p.rect.H
			}
		}
	}
	return max(0, min(top, p.frame.Rows-p.rect.H))
}

// readStream hands the program every JSON value of the stream as it
// comes.
func readStream(in io.Reader, send func(tea.Msg)) {
	dec := json.NewDecoder(in)
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if err == io.EOF {
				err = nil
			}
			send(streamEnd{err})
			return
		}
		send(streamMsg{raw})
	}
}
