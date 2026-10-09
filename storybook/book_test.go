package storybook

import (
	"encoding/json"
	"maps"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytea"
	"github.com/neuroplastio/hotty-go/hottytest"

	"github.com/neuroplastio/hotty-a2ui/rendition/html"
	"github.com/neuroplastio/hotty-a2ui/story"
)

// host is a program with a screen of its own: a row of its own on top,
// and the Book under it, as hotty-demo shows it.
type host struct {
	s     *hottytea.Session
	b     *Book
	w, h  int
	frame string
}

func (m *host) Init() tea.Cmd { return m.s.Detect() }

func (m *host) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	msg, cmd := m.s.Update(msg)
	switch msg := msg.(type) {
	case nil:
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	default:
		m.b.Update(msg, m.s)
	}
	var want []hottytea.Surface
	m.frame, want = m.b.View(hottytea.Rect{X: 0, Y: 1, W: m.w, H: m.h - 1}, m.s)
	m.frame = "the program's own row\n" + m.frame
	m.s.Layout(want)
	m.b.LaidOut(m.s)
	return m, tea.Batch(cmd, m.s.Flush(), m.b.Tick())
}

func (m *host) View() tea.View {
	v := tea.NewView(m.frame)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.Cursor = m.b.Cursor()
	return v
}

func run(t *testing.T, h *hottytest.Host, o Options) {
	t.Helper()
	m := &host{s: hottytea.New(), b: New(o)}
	p := tea.NewProgram(m, tea.WithInput(h), tea.WithOutput(m.s.Watch(h)), tea.WithWindowSize(120, 40),
		tea.WithoutSignalHandler(), tea.WithEnvironment([]string{"TERM=xterm-256color"}))
	m.s.Attach(p.Send)
	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := p.Run(); err != nil {
			t.Error(err)
		}
	}()
	t.Cleanup(func() {
		p.Quit()
		<-done
	})
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	for end := time.Now().Add(5 * time.Second); time.Now().Before(end); time.Sleep(5 * time.Millisecond) {
		if cond() {
			return
		}
	}
	t.Fatalf("never: %s", what)
}

// TestInAProgram: in a program's screen, the Book's surfaces take its
// Prefix and sit where the program put the Book, under its row.
func TestInAProgram(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	run(t, h, Options{First: "hotty/form", Prefix: "sb-"})
	eventually(t, "nav, the story and panel on the host", func() bool {
		return h.Surface("sb-nav") != nil && h.Surface("sb-panel") != nil && h.Surface("sb-s1-0-h") != nil
	})
	if col, row := h.Surface("sb-pick").At(); row != 1 || col != 0 {
		t.Errorf("pick at row %d, col %d: want under the program's row", row, col)
	}
	eventually(t, "nav under pick and its rule", func() bool {
		_, row := h.Surface("sb-nav").At()
		return row == 1+h.Surface("sb-pick").Placement().Rows+1
	})
	if h.Surface("nav") != nil {
		t.Error("a surface without the prefix")
	}
	if !strings.Contains(h.Screen(), "the program's own row") {
		t.Errorf("the program's row is gone:\n%s", h.Screen())
	}
	if err := h.Click("sb-nav", "story_0"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the first story", func() bool { return h.Surface("sb-s2-0-h") != nil })
}

// TestInAProgramInCells: in cells, a click lands where the program put
// the Book, and what is typed goes in the field clicked.
func TestInAProgramInCells(t *testing.T) {
	h := hottytest.New(t, hottytest.Text(), hottytest.Size(120, 40))
	run(t, h, Options{First: "hotty/shortcut-press", Prefix: "sb-"})
	eventually(t, "the story in cells", func() bool { return strings.Contains(h.Screen(), "A note") })
	screen := h.Screen()
	row := 0
	for i, l := range strings.Split(screen, "\n") {
		if strings.Contains(l, "Note") {
			row = i + 1
			break
		}
	}
	at := "\x1b[<0;36;" + strconv.Itoa(row+1) + "M\x1b[<0;36;" + strconv.Itoa(row+1) + "m"
	h.Type(at)
	h.Type("hello")
	eventually(t, "the typing in the field", func() bool { return strings.Contains(h.Screen(), "hello") })
}

// TestNoSurfaceLeftBehind: a story's surfaces go when another story
// opens, and when the rendition moves to cells: the host keeps only the
// chrome's.
func TestNoSurfaceLeftBehind(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	run(t, h, Options{First: "hotty/form", Prefix: "sb-"})
	eventually(t, "the first story on the host", func() bool { return h.Surface("sb-s1-0-h") != nil })
	if err := h.Click("sb-nav", "story_0"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the next story on the host", func() bool { return h.Surface("sb-s2-0-h") != nil })
	if h.Surface("sb-s1-0-h") != nil {
		t.Error("the first story's surface stayed after the next opened")
	}
	h.Type("\x1bOQ") // F2: the next rendition, cells
	eventually(t, "only the chrome on the host", func() bool {
		for _, s := range h.Surfaces() {
			if !slices.Contains([]string{"sb-nav", "sb-pick", "sb-panel"}, s.Name()) {
				return false
			}
		}
		return true
	})
}

// TestSpinnersMove: with the Book's Tick among its commands, a program's
// cells move by themselves: a spinner's frame changes with no input.
func TestSpinnersMove(t *testing.T) {
	h := hottytest.New(t, hottytest.Text(), hottytest.Size(120, 40))
	run(t, h, Options{First: "hotty/spinner", Prefix: "sb-"})
	line := func() string {
		for _, l := range strings.Split(h.Screen(), "\n") {
			if strings.Contains(l, " Line") {
				return l
			}
		}
		return ""
	}
	eventually(t, "the spinners in cells", func() bool { return line() != "" })
	// Past the start's own redraws, only the clock draws.
	time.Sleep(300 * time.Millisecond)
	seen := map[string]bool{}
	for end := time.Now().Add(time.Second); time.Now().Before(end); time.Sleep(20 * time.Millisecond) {
		seen[line()] = true
	}
	if len(seen) < 3 {
		t.Errorf("the Line spinner drew %d frames in a second: %q", len(seen), slices.Collect(maps.Keys(seen)))
	}
}

// TestSpinnersMoveOnHost: on a HOTTY host the clock ticks for the story's
// surface as for cells: the Line spinner's frame changes by deltas.
func TestSpinnersMoveOnHost(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	run(t, h, Options{First: "hotty/spinner", Prefix: "sb-"})
	frame := html.DOMID("s-line") + "~f"
	eventually(t, "the spinners on the host", func() bool {
		s := h.Surface("sb-s1-0-h")
		return s != nil && s.TextOf(frame) != ""
	})
	time.Sleep(300 * time.Millisecond)
	seen := map[string]bool{}
	for end := time.Now().Add(time.Second); time.Now().Before(end); time.Sleep(20 * time.Millisecond) {
		seen[h.Surface("sb-s1-0-h").TextOf(frame)] = true
	}
	if len(seen) < 3 {
		t.Errorf("the Line spinner showed %d frames in a second: %q", len(seen), slices.Collect(maps.Keys(seen)))
	}
}

// TestPickStaysPut: the pickers are a surface of their own over nav, as
// tall as they need, and nav, the list, is what scrolls. A select's open
// list is a surface of its own above the others, and pick and nav stay as
// they are, until a pick closes it.
func TestPickStaysPut(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	run(t, h, Options{First: "hotty/form", Prefix: "sb-"})
	eventually(t, "pick and nav", func() bool {
		p, n := h.Surface("sb-pick"), h.Surface("sb-nav")
		return p != nil && n != nil && p.Placed() && n.Placed()
	})
	eventually(t, "pick as high as its fit, nav under it", func() bool {
		fit := 0
		for _, ev := range h.Events() {
			var d struct{ R int }
			if ev.Surface == "sb-pick" && ev.Kind == hotty.EventFit && json.Unmarshal(ev.Detail, &d) == nil {
				fit = d.R
			}
		}
		rows := h.Surface("sb-pick").Placement().Rows
		_, row := h.Surface("sb-nav").At()
		return rows == fit && row == 1+rows+1
	})
	pick, nav := h.Surface("sb-pick").Placement(), h.Surface("sb-nav").Placement()
	if pick.Rows >= 39/2 || pick.Z != 0 {
		t.Fatalf("pick, closed: %+v", pick)
	}
	if err := h.Click("sb-pick", "theme_p"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the open list above the others", func() bool {
		l := h.Surface("sb-pick-list")
		return l != nil && l.Placed() && l.Placement().Z == 1
	})
	if _, row := h.Surface("sb-pick-list").At(); row <= 1 {
		t.Errorf("the open list at row %d: want under the select", row)
	}
	if p, n := h.Surface("sb-pick").Placement(), h.Surface("sb-nav").Placement(); p != pick || n != nav {
		t.Errorf("open, pick %+v and nav %+v: want %+v and %+v", p, n, pick, nav)
	}
	if err := h.Click("sb-pick-list", "theme_p~o3"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the list gone, the theme picked", func() bool {
		l := h.Surface("sb-pick-list")
		return (l == nil || !l.Placed()) && strings.Contains(h.Surface("sb-pick").Text(), "Nord")
	})
}

// TestPaneScrolls: in cells, a pane taller than its rows scrolls to what
// the keyboard shows only when that changes: a HottyDiff's selected hunk,
// G to the last, g back to the first (the first key picks the first),
// and up one at a time; and the wheel scrolls the pane
// where no scroll view takes it, all the way back up.
func TestPaneScrolls(t *testing.T) {
	s := hottytea.New()
	s.Mode = hottytea.Text
	b := New(Options{First: "hotty/diff"})
	view := func() *pane {
		t.Helper()
		b.View(hottytea.Rect{W: 120, H: 24}, s)
		for _, p := range b.order {
			if p.kind == asCells && !p.chrome(b) {
				return p
			}
		}
		t.Fatal("no pane for the story")
		return nil
	}
	p := view()
	if p.frame.Rows <= p.rect.H {
		t.Fatalf("the story fits its %d rows: %d", p.rect.H, p.frame.Rows)
	}
	if !b.give(p, false) || p.s.C.St.Focus != "review" {
		t.Fatalf("the keyboard on %q, want the diff", p.s.C.St.Focus)
	}
	if p = view(); p.top != 0 {
		t.Errorf("the diff focused: top %d, want its start", p.top)
	}
	b.key("j", s) // from none, the first hunk
	b.key("G", s)
	p = view()
	_, row, _, h, _ := p.cells.Sight("review")
	if p.top == 0 || row < p.top || row+h > p.top+p.rect.H {
		t.Errorf("G: top %d, the last hunk at %d+%d", p.top, row, h)
	}
	b.key("g", s)
	if p = view(); p.top != 0 {
		t.Errorf("g: top %d, want 0", p.top)
	}
	// Up a hunk at a time from the last: each in sight whole, with its
	// file's name, and the first at the top.
	b.key("G", s)
	p = view()
	for i := range len(p.s.C.V.Find("review").RowIDs) - 1 {
		b.key("ArrowUp", s)
		p = view()
		_, row, _, h, _ := p.cells.Sight("review")
		if row < p.top || row+h > p.top+p.rect.H {
			t.Errorf("ArrowUp %d: top %d, the hunk at %d+%d", i+1, p.top, row, h)
		}
	}
	if p.top != 0 {
		t.Errorf("up to the first hunk: top %d, want 0", p.top)
	}
	b.key("G", s)
	p = view()
	for range 20 {
		b.wheel(p.rect.X+1, p.rect.Y+1, tea.MouseWheelMsg{Button: tea.MouseWheelUp})
		p = view()
	}
	if p.top != 0 {
		t.Errorf("the wheel up from the end: top %d, want 0", p.top)
	}
	for range 3 {
		b.wheel(p.rect.X+1, p.rect.Y+1, tea.MouseWheelMsg{Button: tea.MouseWheelDown})
		p = view()
	}
	if p.top != 3*wheelRows {
		t.Errorf("three notches down: top %d, want %d", p.top, 3*wheelRows)
	}
}

// TestTabsAndSidebarOnHost: panel is the right column's tab bar, as high
// as the host's fit, with the story under it; another tab takes the
// whole column and hides the story's surfaces, which come back on the
// preview as they were. F1 hides the sidebar's surfaces, and shows them
// again; the toggle's cell does too.
func TestTabsAndSidebarOnHost(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	run(t, h, Options{First: "hotty/form", Prefix: "sb-"})
	st := storyByName(t, "hotty/form")
	story := func() *hottytest.Surface { return h.Surface("sb-s1-0-h") }
	eventually(t, "the story under panel's bar, as high as its fit", func() bool {
		fit := 0
		for _, ev := range h.Events() {
			var d struct{ R int }
			if ev.Surface == "sb-panel" && ev.Kind == hotty.EventFit && json.Unmarshal(ev.Detail, &d) == nil {
				fit = d.R
			}
		}
		p, s := h.Surface("sb-panel"), story()
		if fit == 0 || p == nil || s == nil || !s.Placed() {
			return false
		}
		_, prow := p.At()
		_, srow := s.At()
		return p.Placement().Rows == fit && srow == prow+fit
	})
	h.Type("\x1b[15~") // F5: About
	eventually(t, "About over the whole column, the story hidden", func() bool {
		p, s := h.Surface("sb-panel"), story()
		return s != nil && !s.Placed() && p.Placement().Rows == 39 && !p.Placement().Fit && strings.Contains(p.Text(), st.Title)
	})
	for range 4 {
		h.Type("\x1b[15~") // round to the preview
	}
	eventually(t, "the story back, the same surface", func() bool { s := story(); return s != nil && s.Placed() })
	h.Type("\x1bOP") // F1
	eventually(t, "the sidebar hidden", func() bool {
		p, n := h.Surface("sb-pick"), h.Surface("sb-nav")
		col, _ := h.Surface("sb-panel").At()
		return p != nil && n != nil && !p.Placed() && !n.Placed() && col == 2
	})
	h.Type("\x1b[<0;1;2M\x1b[<0;1;2m") // a click on ▸, at the rule's top
	eventually(t, "the sidebar back", func() bool { return h.Surface("sb-pick").Placed() && h.Surface("sb-nav").Placed() })
}

func storyByName(t *testing.T, name string) *story.Story {
	t.Helper()
	st := story.Find(name)
	if st == nil {
		t.Fatalf("no story %s", name)
	}
	return st
}

// TestTabsKeepTheStory: in cells, another tab hides the story's panes
// and takes the keyboard from them; back on the preview, they are as they
// were (the same pane, scrolled where it was, its hunk selected). F1 hides
// the sidebar and gives its columns to the right.
func TestTabsKeepTheStory(t *testing.T) {
	s := hottytea.New()
	s.Mode = hottytea.Text
	b := New(Options{First: "hotty/diff"})
	storyPane := func() *pane {
		b.View(hottytea.Rect{W: 120, H: 24}, s)
		for _, p := range b.order {
			if !p.chrome(b) {
				return p
			}
		}
		return nil
	}
	press := func(k string) { b.key(k, s); b.settle() } // as Update does
	p := storyPane()
	b.give(p, false)
	press("j")
	press("G")
	p = storyPane()
	top := p.top
	press("F5")
	if q := storyPane(); q != nil || b.focus != nil {
		t.Fatalf("on About: the story's pane drawn (%v), the keyboard on %v", q != nil, b.focus)
	}
	if !slices.Contains(b.hidden, p) || b.panes[b.o.Prefix+panelID].rect.H != 24 {
		t.Errorf("on About: the story's pane not kept, or panel not the whole column")
	}
	for range panelTabs - 1 {
		press("F5")
	}
	q := storyPane()
	if q != p || q.top != top || q.s.C.S.Data.Value("/hunk") == nil {
		t.Errorf("back on the preview: same pane %v, top %d (was %d), hunk %v", q == p, q.top, top, q.s.C.S.Data.Value("/hunk"))
	}
	x := q.rect.X
	press("F1")
	if q = storyPane(); q.rect.X != 1 || q.rect.W <= 120-x-1 {
		t.Errorf("without the sidebar: the story at %d, %d wide", q.rect.X, q.rect.W)
	}
}

// TestPanelJSON: panel's tabs are JSON in HottyCode. The actions are
// indented, newest first. Each message is on its line under a comment
// that says which way it went and what it is, and a failed one is marked,
// with its error after it. Each surface's data model is under its id.
func TestPanelJSON(t *testing.T) {
	ch := newChrome(entries(false), []renditionOption{{rendCells, "Cells"}}, rendCells, "Terminal", keysTerminal, false)
	run := story.NewRun()
	if err := run.Feed(json.RawMessage(`{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"https://a2ui.org/specification/v1_0/catalogs/basic/catalog.json","dataModel":{"n":1}}}`)); err != nil {
		t.Fatal(err)
	}
	if run.Feed(json.RawMessage(`{"version": "v1.0", "updateDataModel": {"surfaceId": "nope", "path": "/n", "value": 2}}`)) == nil {
		t.Fatal("a surface that isn't there took a value")
	}
	run.Log = append(run.Log, story.Entry{Out: true, JSON: json.RawMessage(`{"version":"v1.0","action":{"name":"go","surfaceId":"s"}}`)})
	ch.report(run)

	data := ch.surface(panelID).S.Data
	if got, want := data.Value("/actions"), "{\n  \"version\": \"v1.0\",\n  \"action\": {\n    \"name\": \"go\",\n    \"surfaceId\": \"s\"\n  }\n}"; got != want {
		t.Errorf("actions:\n%v\nwant\n%s", got, want)
	}
	msgs, _ := data.Value("/messages").(string)
	lines := strings.Split(msgs, "\n")
	// The renderer's error back to the agent comes before the message
	// that failed.
	want := []string{
		"// ← action",
		`{"version":"v1.0","action":{"name":"go","surfaceId":"s"}}`,
		"// ← error",
		`{"version":"v1.0","error":{"code":"INTEGRITY_ERROR","message":"Surface not found for message: nope","surfaceId":"nope"}}`,
		"// → updateDataModel",
		`{"version":"v1.0","updateDataModel":{"surfaceId":"nope","path":"/n","value":2}}`,
		"// ✗ Surface not found for message: nope",
		"// → createSurface",
	}
	if len(lines) != 9 || !slices.Equal(lines[:8], want) {
		t.Errorf("messages:\n%s", msgs)
	}
	if got := ch.sent[panelID+"#marks"]; got != `[{"end":7,"kind":"error","line":6}]` {
		t.Errorf("the failed message's marks: %s", got)
	}
	if got, want := data.Value("/data"), "// s\n{\n  \"n\": 1\n}"; got != want {
		t.Errorf("data:\n%v\nwant\n%s", got, want)
	}

	// Nothing yet: a comment says so.
	ch.report(story.NewRun())
	if got := data.Value("/actions"); got != "// No action yet: the user's go here, as the agent gets them." {
		t.Errorf("no action: %v", got)
	}
	if got := ch.sent[panelID+"#marks"]; got != "null" {
		t.Errorf("marks with nothing failed: %s", got)
	}
}
