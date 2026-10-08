// Package storybook is the HOTTY kit's storybook (NEIO-11): every story in
// the rendition picked, live, with its actions, data model and messages;
// or what an agent streams. cmd/storybook runs it on the whole screen. A
// Bubble Tea program with screens of its own (hotty-demo) shows it in a
// part of one, over the program's hottytea.Session.
package storybook

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytea"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/rendition/html"
	"github.com/neuroplastio/hotty-a2ui/rendition/text"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/view"
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
	rect  hottytea.Rect // in the Book, not on the screen
	frame *cells.Frame
	top   int // the frame's first row shown
}

// takesInput: text is for reading. Cells beside a surface take input as
// well: the keyboard is in one of the two at a time (b.focus), and the
// other shows the same state.
func (p *pane) takesInput() bool { return p.kind != asText }

// Options are how a Book starts.
type Options struct {
	// First is the story to open first, by name; the list's first if "".
	First string
	// Stream: the Book shows what Feed hands it, under an entry of its
	// own, first in the list. Source says where it comes from.
	Stream bool
	Source string
	// Out, if set, hears what the renderer sends the agent: actions,
	// errors and function calls.
	Out func(a2ui.Outbound)
	// Rendition is the one to start in, if the terminal can show it:
	// "surfaces", "cells", "text" or "side" (surfaces beside cells).
	Rendition string
	// Theme is the theme to start in; the zero theme is theme.Default.
	Theme theme.Theme
	// Plain draws the cells without colours (NO_COLOR).
	Plain bool
	// Bars: the Book has the screen to itself, and draws its name and the
	// story shown on its first row, and its keys on its last.
	Bars bool
	// Prefix goes before the name of every surface the Book places, so
	// that they keep clear of the program's own.
	Prefix string
}

// Book is the storybook. Its methods are for the program's Update and
// View, on Bubble Tea's goroutine.
type Book struct {
	o     Options
	at    hottytea.Rect // where the last View put the Book on the screen
	mode  hottytea.Mode
	theme theme.Theme // the colours the cells and the kit paint with

	ch     *chrome
	list   []entry
	cur    string
	run    *story.Run
	stream *story.Run
	rend   string
	gen    int
	seq    map[*a2ui.Surface]int

	panes  map[string]*pane
	order  []*pane // this frame's panes, as drawn: nav, pick, the story's, panel
	focus  *pane   // the pane with the keyboard, if any
	last   *pane   // the pane that had it last: Tab goes on from there
	cursor *tea.Cursor
	status string

	// pickRows is pick's height with its lists closed: the host's fit on a
	// host, the frame's rows in cells; 0 until known.
	pickRows int
}

// New is a Book. It starts once the terminal is known: on the Session's
// ReadyMsg, or at the first View after it.
func New(o Options) *Book {
	b := &Book{o: o, theme: o.Theme, seq: map[*a2ui.Surface]int{}, panes: map[string]*pane{}}
	if b.theme.Name == "" {
		b.theme = theme.Default
	}
	if o.Stream {
		b.stream = story.NewRun()
		b.stream.Out = o.Out
		if b.o.First == "" {
			b.o.First = streamName
		}
	}
	return b
}

// Update takes a message, as the Session's Update left it: the user's
// keys and clicks, and the host's events. Its commands to the host go out
// with the Session's next Flush. quit reports Control+C, Control+Q, or q
// while no pane has the keyboard.
func (b *Book) Update(msg tea.Msg, h *hottytea.Session) (quit bool) {
	switch msg := msg.(type) {
	case hottytea.ReadyMsg:
		b.ready(msg.Mode)
	case hottytea.EventMsg:
		b.event(msg.Event)
	case hottytea.ErrorMsg:
		if b.mine(msg.Reply.Surface) {
			b.status = "✗ the host: " + msg.Reply.Err().Error()
		}
	case tea.KeyPressMsg:
		quit = b.key(keyValue(msg.Key()), h)
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			b.click(msg.X-b.at.X, msg.Y-b.at.Y)
		}
	case tea.MouseMotionMsg:
		if p := b.focus; msg.Button == tea.MouseLeft && p != nil && p.kind == asCells {
			b.fail(p.cells.Drag(msg.X-b.at.X-p.rect.X, msg.Y-b.at.Y-p.rect.Y+p.top))
		}
	case tea.MouseReleaseMsg:
		if p := b.focus; p != nil && p.kind == asCells {
			p.cells.Release()
		}
	}
	b.settle()
	return quit
}

// mine reports whether a surface is one the Book placed, as far as its
// name tells: any, without a Prefix.
func (b *Book) mine(surface string) bool {
	return b.o.Prefix == "" || strings.HasPrefix(surface, b.o.Prefix)
}

// Feed hands the Book a message of the stream (Options.Stream).
func (b *Book) Feed(raw json.RawMessage) {
	if b.stream == nil {
		return
	}
	if err := b.stream.Feed(raw); err != nil {
		b.status = "✗ " + strings.TrimPrefix(err.Error(), "a2ui: ")
	}
	b.settle()
}

// End says the stream has ended, and with what error, if any.
func (b *Book) End(err error) {
	b.status = "the stream ended"
	if err != nil {
		b.status = "✗ the stream: " + err.Error()
	}
}

// Story is the name of the story shown.
func (b *Book) Story() string { return b.cur }

// Status is what went wrong last, or "".
func (b *Book) Status() string { return b.status }

// Typing reports whether a pane has the keyboard: keys go to the story
// or the Book's own surfaces, q among them.
func (b *Book) Typing() bool { return b.focus != nil }

// Keyboard reports whether a phone should show its keyboard: a field has
// the keyboard in cells. A host's surface asks for its own.
func (b *Book) Keyboard() bool {
	p := b.focus
	if p == nil || p.kind != asCells || !p.s.C.St.Keyboard {
		return false
	}
	e := p.s.C.V.Find(p.s.C.St.Focus)
	return e != nil && (e.Kind == view.TextField || e.Kind == view.DateTime)
}

// Cursor is where the terminal's cursor goes, on the screen: in the field
// that has the keyboard in cells; nil for none.
func (b *Book) Cursor() *tea.Cursor { return b.cursor }

// ready starts the storybook once the terminal is known: surfaces on a
// host, cells elsewhere.
func (b *Book) ready(mode hottytea.Mode) {
	if b.ch != nil || mode == hottytea.Detecting {
		return
	}
	b.mode = mode
	opts := []renditionOption{{rendCells, "Cells"}, {rendText, "Text"}}
	b.rend = rendCells
	if mode == hottytea.Native {
		opts = append([]renditionOption{{rendSurfaces, "Surfaces"}}, append(opts, renditionOption{rendSide, "Side by side"})...)
		b.rend = rendSurfaces
	}
	for _, o := range opts {
		if o.value == b.o.Rendition {
			b.rend = o.value
		}
	}
	b.list = entries(b.stream != nil)
	b.ch = newChrome(b.list, opts, b.rend, b.theme.Name, mode == hottytea.Native)
	b.open(cmp.Or(b.o.First, b.list[0].name))
}

// open shows a story: a new run of it, or the stream as it is.
func (b *Book) open(name string) {
	head := ""
	if name == streamName && b.stream != nil {
		b.run = b.stream
		head = "**The stream** · A2UI from " + b.o.Source
	} else {
		st := story.Find(name)
		if st == nil {
			b.status = "✗ no story " + name
			return
		}
		run := story.NewRun()
		run.Out = b.o.Out
		for _, msg := range st.Messages {
			if err := run.Feed(msg); err != nil {
				b.status = "✗ " + strings.TrimPrefix(err.Error(), "a2ui: ")
				break
			}
		}
		b.run = run
		head = "**" + st.Title + "** · " + st.Description
	}
	if f := b.focus; f != nil && f.s != b.ch.surface(navID) && f.s != b.ch.surface(pickID) {
		b.focus = nil
	}
	b.cur = name
	b.gen++
	b.seq = map[*a2ui.Surface]int{}
	b.ch.showing(name, head)
}

// settle does what the last message left to do: the storybook's own
// actions, the rendition picked, the panel.
func (b *Book) settle() {
	if b.ch == nil {
		return
	}
	acts := b.ch.acts
	b.ch.acts = nil
	for _, a := range acts {
		if name, _ := a.Context["name"].(string); a.Name == "open" && name != "" {
			b.open(name)
		}
	}
	if t, ok := theme.ByName(b.ch.theme()); ok && t.Name != b.theme.Name {
		b.theme = t
		b.gen++
	}
	if r := b.ch.rendition(); r != "" && r != b.rend {
		b.rend = r
		if b.focus != nil && !b.focus.chrome(b) {
			b.focus = nil
		}
	}
	if b.focus != nil && !b.focus.s.C.St.Keyboard {
		b.focus = nil
	}
	if b.run != nil {
		b.ch.report(b.run)
	}
}

func (p *pane) chrome(b *Book) bool {
	return p.s == b.ch.surface(pickID) || p.s == b.ch.surface(navID) || p.s == b.ch.surface(panelID)
}

// key handles a key that reached the storybook: the pane with the
// keyboard first (on a host, only what the surface left), then the
// story's Shortcuts, then the storybook's own keys.
func (b *Book) key(k string, h *hottytea.Session) (quit bool) {
	b.status = ""
	if k == "Control+c" || k == "Control+q" {
		return true
	}
	if b.ch == nil {
		return false
	}
	if p := b.focus; p != nil {
		switch p.kind {
		case asCells:
			had := p.s.C.St.Keyboard
			handled, err := p.cells.Key(k)
			b.fail(err)
			if handled {
				if had && !p.s.C.St.Keyboard && (k == "Tab" || k == "Shift+Tab") {
					b.focus = nil
					b.cycle(k == "Shift+Tab", p)
				}
				return false
			}
		case asSurface:
			cmds, ok, err := p.html.Key(k)
			h.Send(cmds...)
			b.fail(err)
			if ok {
				return false
			}
		}
	} else if b.shortcut(k, h) {
		return false
	}
	switch k {
	case "Tab", "Shift+Tab":
		b.cycle(k == "Shift+Tab", cmp.Or(b.focus, b.last))
	case "Escape":
		b.escape()
	case "F2":
		b.nextRendition()
	case "F3":
		b.nextTheme()
	case "q":
		return b.focus == nil
	}
	return false
}

// shortcut offers a key to the story's surfaces, while no pane has the
// keyboard: the story's is the surface keys apply to.
func (b *Book) shortcut(k string, h *hottytea.Session) bool {
	seen := map[*story.Surface]bool{}
	for _, p := range b.order {
		if !p.takesInput() || p.chrome(b) || seen[p.s] {
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
			h.Send(cmds...)
		}
		b.fail(err)
		if ok {
			return true
		}
	}
	return false
}

// cycle gives the keyboard to the next pane that takes it, after from
// (or before it, back), at its first focusable element (or its last).
// The story's panes come first, then panel, pick and nav: the story is
// what the keyboard is for, and nav lists every story.
func (b *Book) cycle(back bool, from *pane) {
	var ps []*pane
	for _, p := range b.order {
		if p.takesInput() && !p.chrome(b) {
			ps = append(ps, p)
		}
	}
	for _, id := range []string{panelID, pickID, navID} {
		for _, p := range b.order {
			if p.takesInput() && p.s == b.ch.surface(id) {
				ps = append(ps, p)
			}
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
		if b.give(ps[i], back) {
			return
		}
	}
}

// give gives a pane the keyboard, at its first focusable element (its
// last, back); false if it has none.
func (b *Book) give(p *pane, back bool) bool {
	c := p.s.C
	c.St.Keyboard = false
	if !c.FocusNext(back) {
		return false
	}
	if f := b.focus; f != nil && f != p && f.s != p.s {
		f.s.C.St.Keyboard = false
	}
	b.focus, b.last = p, p
	return true
}

// escape is Escape that no surface took: it closes a Modal, else the
// keyboard goes back to the storybook.
func (b *Book) escape() {
	for _, p := range b.order {
		if p.s.C.St.Modal != "" && (b.focus == nil || b.focus.s == p.s) {
			p.s.C.CloseModal()
			return
		}
	}
	if b.focus != nil {
		b.focus.s.C.St.Keyboard = false
		b.focus = nil
	}
}

func (b *Book) nextRendition() {
	var opts []string
	for _, o := range []string{rendSurfaces, rendCells, rendText, rendSide} {
		if b.mode == hottytea.Native || o == rendCells || o == rendText {
			opts = append(opts, o)
		}
	}
	next := opts[(slices.Index(opts, b.rend)+1)%len(opts)]
	delete(b.ch.sent, pickID+"/rendition")
	b.ch.set(pickID, "/rendition", []any{next})
}

// nextTheme shows the next theme in pick.
func (b *Book) nextTheme() {
	delete(b.ch.sent, pickID+"/theme")
	b.ch.set(pickID, "/theme", []any{theme.Next(b.theme).Name})
}

// click is a primary click on the cells, in the Book: in a cells pane,
// the one drawn last where they overlap (pick's open list over nav), its
// rendition takes it; anywhere else the keyboard goes back to the
// storybook.
func (b *Book) click(x, y int) {
	for _, p := range slices.Backward(b.order) {
		if p.kind != asCells || x < p.rect.X || x >= p.rect.X+p.rect.W || y < p.rect.Y || y >= p.rect.Y+p.rect.H {
			continue
		}
		if f := b.focus; f != nil && f != p && f.s != p.s {
			f.s.C.St.Keyboard = false
		}
		b.fail(p.cells.Click(x-p.rect.X, y-p.rect.Y+p.top))
		if p.s.C.St.Keyboard {
			b.focus = p
		} else if b.focus == p {
			b.focus = nil
		}
		return
	}
	if f := b.focus; f != nil && f.kind == asCells {
		f.s.C.St.Keyboard = false
		b.focus = nil
	}
}

// event is what the user did in a surface on the host.
func (b *Book) event(ev hotty.Event) {
	p := b.panes[ev.Surface]
	if p == nil || p.html == nil {
		return
	}
	b.fail(p.html.Event(ev))
	switch ev.Kind {
	case hotty.EventFit:
		// pick's height, while its lists are closed: open, it takes the
		// column, and needs what it has.
		var d struct {
			R int `json:"r"`
		}
		if p.name == b.o.Prefix+pickID && !p.html.ListOpen() && json.Unmarshal(ev.Detail, &d) == nil && d.R > 0 {
			b.pickRows = d.R
		}
	case hotty.EventFocus:
		if f := b.focus; f != nil && f != p && f.kind == asCells && f.s != p.s {
			f.s.C.St.Keyboard = false
		}
		b.focus, b.last = p, p
	case hotty.EventBlur:
		b.last = p
	}
}

func (b *Book) fail(err error) {
	if err != nil {
		b.status = "✗ " + strings.TrimPrefix(err.Error(), "a2ui: ")
	}
}

// pane is the pane of a surface in a rendition, kept from frame to frame
// so its rendition keeps what it has (the document the host has, a
// field's cursor).
func (b *Book) pane(name string, s *story.Surface, k kind, r hottytea.Rect) *pane {
	p := b.panes[name]
	if p == nil || p.s != s || p.kind != k {
		p = &pane{name: name, s: s, kind: k}
		switch k {
		case asSurface:
			p.html = html.New(s.C, name)
		case asCells:
			p.cells = cells.New(s.C)
		}
		b.panes[name] = p
	}
	p.rect = r
	return p
}

// View lays the Book out in r, on the screen, and draws it: its cells, r.W
// by r.H, and the surfaces it wants there. The program lays those out with
// its own (Session.Layout), then calls LaidOut. The surfaces' deltas are
// sent already.
func (b *Book) View(r hottytea.Rect, h *hottytea.Session) (string, []hottytea.Surface) {
	b.at, b.cursor = r, nil
	b.ready(h.Mode)
	if b.ch == nil {
		return "Finding out whether the terminal is a HOTTY host…", nil
	}
	W, H := r.W, r.H
	minH := 10
	if b.o.Bars {
		minH = 12
	}
	if W < 40 || H < minH {
		return fmt.Sprintf("The storybook needs 40×%d at least.", minH), nil
	}
	scr := newScreen(W, H)
	native := b.mode == hottytea.Native

	top, bodyH := 0, H
	if b.o.Bars {
		x := scr.put(0, 0, W, "HOTTY kit storybook", cells.Fg, cells.Bold)
		scr.put(x, 0, W-x, " · "+b.cur+" · "+b.rend, cells.Muted, 0)
		help := "Tab next · Esc leave · F2 rendition · Ctrl+C quit"
		if b.status != "" {
			help = b.status + " · " + help
		}
		scr.put(0, H-1, W, help, cells.Muted, 0)
		top, bodyH = 1, H-2
	}
	navW := min(max(W/4, 22), 36, W/2)
	scr.vrule(navW, top, bodyH)
	rx := navW + 1
	rw := W - rx
	panelH := min(max(bodyH/3, 6), 16)
	canvasH := bodyH - panelH - 1
	scr.hrule(rx, top+canvasH, rw)

	chromeKind := asCells
	if native {
		chromeKind = asSurface
	}
	// pick stays at the top of the left column, and nav scrolls under it.
	// A select's open list needs more room than pick has: pick then takes
	// the column, over nav, until the list closes.
	pick := b.pane(b.o.Prefix+pickID, b.ch.surface(pickID), chromeKind, hottytea.Rect{X: 0, Y: top, W: navW})
	pickH, open := b.pickHeight(pick, bodyH)
	pick.rect.H = pickH
	if open {
		pick.rect.H = bodyH
	}
	nav := b.pane(b.o.Prefix+navID, b.ch.surface(navID), chromeKind, hottytea.Rect{X: 0, Y: top + pickH, W: navW, H: bodyH - pickH})
	order := []*pane{nav, pick}

	kinds := map[string][]kind{rendSurfaces: {asSurface}, rendCells: {asCells}, rendText: {asText}, rendSide: {asSurface, asCells}}[b.rend]
	if !native {
		kinds = slices.DeleteFunc(slices.Clone(kinds), func(k kind) bool { return k == asSurface })
		if len(kinds) == 0 {
			kinds = []kind{asCells}
		}
	}
	var ss []*story.Surface
	if b.run != nil {
		ss = b.run.Surfaces()
	}
	colW := (rw - (len(kinds) - 1)) / len(kinds)
	for ki, k := range kinds {
		cx := rx + ki*(colW+1)
		if ki > 0 {
			scr.vrule(cx-1, top, canvasH)
		}
		y := top
		for i, s := range ss {
			sh := canvasH / len(ss)
			if i == len(ss)-1 {
				sh = canvasH - (y - top)
			}
			if sh <= 0 {
				continue
			}
			if _, ok := b.seq[s.S]; !ok {
				b.seq[s.S] = len(b.seq)
			}
			name := fmt.Sprintf("%ss%d-%d-%c", b.o.Prefix, b.gen, b.seq[s.S], "hct"[k])
			p := b.pane(name, s, k, hottytea.Rect{X: cx, Y: y, W: colW, H: sh})
			order = append(order, p)
			y += sh
		}
	}
	if len(ss) == 0 {
		scr.put(rx+1, top, rw-1, "No surface yet.", cells.Muted, 0)
	}
	order = append(order, b.pane(b.o.Prefix+panelID, b.ch.surface(panelID), chromeKind, hottytea.Rect{X: rx, Y: top + 1 + canvasH, W: rw, H: panelH}))
	b.order = order
	for name, p := range b.panes {
		if !slices.Contains(order, p) {
			delete(b.panes, name)
			if b.focus == p {
				b.focus = nil
			}
			if b.last == p {
				b.last = nil
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
			p.html.SetTheme(b.theme)
			p.html.Away = b.focus != nil && b.focus != p && b.focus.s == p.s
			h.Send(p.html.Update()...)
			at := p.rect
			at.X, at.Y = at.X+r.X, at.Y+r.Y
			s := hottytea.Surface{Name: p.name, Rect: at, Doc: p.html.Doc, Scroll: hotty.ScrollVertical}
			if p == pick {
				s.Scroll, s.Fit = 0, true
				if open {
					s.Z = 1
				}
			}
			want = append(want, s)
		}
	}
	if f := b.focus; f != nil && f.kind == asCells && f.s.C.St.Keyboard && f.frame != nil {
		if c, row, ok := f.frame.Cursor(); ok && row >= f.top && row < f.top+f.rect.H {
			b.cursor = tea.NewCursor(r.X+f.rect.X+c, r.Y+f.rect.Y+row-f.top)
		}
	}
	if b.o.Plain {
		return scr.ANSI(false), want
	}
	return scr.Themed(b.theme), want
}

// pickHeight is pick's height with its lists closed, at most half the
// column, and whether one is open.
func (b *Book) pickHeight(p *pane, column int) (h int, open bool) {
	switch p.kind {
	case asCells:
		open = p.cells.ListOpen()
		if !open || b.pickRows == 0 {
			b.pickRows = p.cells.Draw(p.rect.W).Rows
		}
	case asSurface:
		open = p.html.ListOpen()
		if b.pickRows == 0 {
			b.pickRows = 4 // until the host's fit says
		}
	}
	return min(b.pickRows, column/2), open
}

// LaidOut is for after the program's Layout: a surface that has just got
// its document gets the keyboard now, if its view has it (autofocus).
func (b *Book) LaidOut(h *hottytea.Session) {
	for _, p := range b.order {
		if p.kind == asSurface {
			h.Send(p.html.Update()...)
		}
	}
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
