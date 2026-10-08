package storybook

import (
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
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
	return m, tea.Batch(cmd, m.s.Flush())
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
	eventually(t, "nav under pick", func() bool {
		_, row := h.Surface("sb-nav").At()
		return row == 1+h.Surface("sb-pick").Placement().Rows
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

// TestPickStaysPut: the pickers are a surface of their own over nav, as
// tall as they need, and nav, the list, is what scrolls. A select's open
// list takes the column, over nav, until a pick closes it.
func TestPickStaysPut(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	run(t, h, Options{First: "hotty/form", Prefix: "sb-"})
	eventually(t, "pick and nav", func() bool {
		p, n := h.Surface("sb-pick"), h.Surface("sb-nav")
		return p != nil && n != nil && p.Placed() && n.Placed()
	})
	closed := h.Surface("sb-pick").Placement()
	if closed.Rows >= 39/2 || closed.Z != 0 {
		t.Fatalf("pick, closed: %+v", closed)
	}
	if err := h.Click("sb-pick", "theme_p"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the open list over nav", func() bool {
		p := h.Surface("sb-pick").Placement()
		return p.Rows == 39 && p.Z == 1
	})
	if err := h.Click("sb-pick", "theme_p~o3"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "pick closed again", func() bool {
		p := h.Surface("sb-pick").Placement()
		return p.Rows < 39/2 && p.Z == 0
	})
}
