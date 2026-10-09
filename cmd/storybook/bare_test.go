package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/story"
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
	m.Update(tea.KeyPressMsg{Code: 'A', Text: "A"})
	if !strings.Contains(m.frame, "A") {
		t.Fatalf("the field did not take a key:\n%s", m.frame)
	}
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}); cmd == nil {
		t.Fatal("Control+C does not quit")
	}
}
