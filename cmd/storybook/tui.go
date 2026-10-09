package main

import (
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go/hottytea"

	"github.com/neuroplastio/hotty-a2ui/storybook"
)

type (
	streamMsg struct{ raw json.RawMessage }
	streamEnd struct{ err error }
	// endSignal is a signal to end (endOnSignal).
	endSignal struct{}
)

// endOnSignal ends the program on TERM (kill's), HUP (the terminal gone)
// or INT (Control+C with no terminal to make it a key) as Control+C does:
// its surfaces deleted first. Bubble Tea's own handling quits at once and
// leaves them on the host, where they stay until something deletes them
// (SPEC.md §5.4). A second signal ends the program as it would without
// this.
func endOnSignal(p *tea.Program) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		<-sig
		signal.Stop(sig)
		p.Send(endSignal{})
	}()
}

// model is the interactive storybook: the Book on the whole screen.
type model struct {
	s     *hottytea.Session
	b     *storybook.Book
	w, h  int
	frame string
}

func newModel(o storybook.Options) *model {
	o.Bars = true
	o.Plain = os.Getenv("NO_COLOR") != ""
	return &model{s: hottytea.New(), b: storybook.New(o)}
}

func (m *model) Init() tea.Cmd { return m.s.Detect() }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	msg, cmd := m.s.Update(msg)
	switch msg := msg.(type) {
	case nil:
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case streamMsg:
		m.b.Feed(msg.raw)
	case streamEnd:
		m.b.End(msg.err)
	case endSignal:
		return m, tea.Sequence(m.s.Close(), tea.Quit)
	default:
		if m.b.Update(msg, m.s) {
			return m, tea.Sequence(m.s.Close(), tea.Quit)
		}
	}
	return m, tea.Batch(cmd, m.draw(), m.b.Tick())
}

func (m *model) View() tea.View {
	v := tea.NewView(m.frame)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.Cursor = m.b.Cursor()
	v.WindowTitle = "storybook · " + m.b.Story()
	return v
}

// draw draws the Book, and lays its surfaces out.
func (m *model) draw() tea.Cmd {
	if m.w == 0 {
		m.frame = "Finding out whether the terminal is a HOTTY host…"
		return nil
	}
	var want []hottytea.Surface
	m.frame, want = m.b.View(hottytea.Rect{W: m.w, H: m.h}, m.s)
	m.s.Layout(want)
	m.b.LaidOut(m.s)
	return m.s.Flush()
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
