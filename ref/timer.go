package main

import (
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/stopwatch"
	"charm.land/bubbles/v2/timer"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func init() {
	register("timer", "hotty/timer", "bubbles: timer and stopwatch (KIT-14)", func() tea.Model {
		return screen{newTimers()}
	})
}

// timersRef: the story hotty/timer's two timers and its stopwatch, as
// bubbles' timer and stopwatch examples show them: the time after a
// label, View as it is (Go's way of writing a duration), and the
// stopwatch's keys in a help line, s starting and stopping it and r
// resetting it. bubbles has no clock format: Focus shows 24m57s where the
// story asks for 24:57.
type timersRef struct {
	tea, focus timer.Model
	lap        stopwatch.Model
	help       help.Model
	toggle     key.Binding
	reset      key.Binding
}

func newTimers() *timersRef {
	return &timersRef{
		tea:    timer.New(10 * time.Second),
		focus:  timer.New(25 * time.Minute),
		lap:    stopwatch.New(stopwatch.WithInterval(100 * time.Millisecond)),
		help:   help.New(),
		toggle: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start/stop lap")),
		reset:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reset lap")),
	}
}

func (m *timersRef) Init() tea.Cmd {
	return tea.Batch(m.tea.Init(), m.focus.Init(), m.lap.Init())
}

func (m *timersRef) Update(msg tea.Msg) tea.Cmd {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(k, m.toggle):
			return m.lap.Toggle()
		case key.Matches(k, m.reset):
			return m.lap.Reset()
		}
	}
	var cmds [3]tea.Cmd
	m.tea, cmds[0] = m.tea.Update(msg)
	m.focus, cmds[1] = m.focus.Update(msg)
	m.lap, cmds[2] = m.lap.Update(msg)
	return tea.Batch(cmds[:]...)
}

func (m *timersRef) View() string {
	bold := lipgloss.NewStyle().Bold(true)
	return bold.Render("Timer and stopwatch") + "\n\n" +
		"Lap " + m.lap.View() + "\n" +
		"Focus " + m.focus.View() + "\n" +
		"Tea " + m.tea.View() + "\n\n" +
		m.help.ShortHelpView([]key.Binding{m.toggle, m.reset})
}
