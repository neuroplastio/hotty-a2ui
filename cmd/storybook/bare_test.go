package main

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/storybook"
)

// -bare draws the story's surface alone, and Tab gives it the keyboard:
// what a reference shot needs from the kit's side.
func TestBare(t *testing.T) {
	st := story.Find("hotty/form")
	if st == nil {
		t.Fatal("no story hotty/form")
	}
	run, err := story.Start(st)
	if err != nil {
		t.Fatal(err)
	}
	m := &bareModel{run: run, th: theme.Default, plain: true}
	for _, s := range run.Surfaces() {
		m.panes = append(m.panes, &barePane{s: s, r: cells.New(s.C)})
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !strings.Contains(m.frame, "Sign up") || strings.Contains(m.frame, "Simple Login Form") {
		t.Fatalf("frame is not the story's surface alone:\n%s", m.frame)
	}
	if m.keyboard() != nil || m.cur != nil {
		t.Fatal("the surface has the keyboard before Tab")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if p := m.keyboard(); p == nil || p.s.C.St.Focus != "name" {
		t.Fatalf("Tab gave the keyboard to %v, want the Name field", m.keyboard())
	}
	if m.cur == nil {
		t.Fatal("no cursor in the focused field")
	}
	m.Update(tea.KeyPressMsg{Code: 'a', Text: "A", Mod: tea.ModShift})
	if !strings.Contains(m.frame, "A") {
		t.Fatalf("the field did not take a key:\n%s", m.frame)
	}
	ctrlC := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	m.Update(ctrlC)
	if m.quitAt.IsZero() || !strings.Contains(m.frame, "Ctrl+C again to quit") {
		t.Fatalf("one ctrl+c did not ask for another:\n%s", m.frame)
	}
	m.Update(tea.KeyPressMsg{Code: 'b', Text: "B", Mod: tea.ModShift})
	if !m.quitAt.IsZero() || strings.Contains(m.frame, "again to quit") {
		t.Fatal("another key did not end the ask")
	}
	m.Update(ctrlC)
	m.Update(bareQuitEnd{m.quitAt})
	if !m.quitAt.IsZero() || strings.Contains(m.frame, "again to quit") {
		t.Fatal("the window's end did not end the ask")
	}
	m.Update(ctrlC)
	m.quitAt = m.quitAt.Add(-2 * storybook.QuitWindow)
	if m.Update(ctrlC); m.quitAt.IsZero() {
		t.Fatal("a second ctrl+c after the window quit")
	}
	if _, cmd := m.Update(ctrlC); cmd == nil || !m.quitAt.IsZero() {
		t.Fatal("a quick second ctrl+c does not quit")
	}
}

// -bare scrolls a story taller than the screen as one page: a jump to a
// heading puts it at the top, Space scrolls a page on, and the frame is
// the screen's rows of it.
func TestBarePage(t *testing.T) {
	run, err := story.Start(story.Find("hotty/markdown"))
	if err != nil {
		t.Fatal(err)
	}
	m := &bareModel{run: run, th: theme.Default, plain: true}
	for _, s := range run.Surfaces() {
		m.panes = append(m.panes, &barePane{s: s, r: cells.New(s.C)})
	}
	sgr := regexp.MustCompile("\x1b\\[[0-9;]*m")
	text := func() string { return sgr.ReplaceAllString(m.frame, "") }
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 12})
	if m.rows <= 24 || strings.Count(text(), "\n") != 11 || !strings.HasPrefix(text(), "Lovelace") {
		t.Fatalf("%d rows, the screen's 12 from the top:\n%s", m.rows, text())
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // configuration
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !strings.HasPrefix(text(), "Configuration") {
		t.Fatalf("after the jump, top %d:\n%s", m.scroll.Top, text())
	}
	top := m.scroll.Top
	m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if m.scroll.Top != min(top+12, m.rows-12) || m.scroll.Top == top {
		t.Errorf("Space: top %d, want a page on from %d", m.scroll.Top, top)
	}
}
