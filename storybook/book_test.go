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
