package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go/hottytest"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/storybook"
)

// The storybook's own surfaces, as the host has them.
const (
	pickID  = "pick"
	navID   = "nav"
	panelID = "panel"
)

// runBook runs the interactive storybook on a test host until the test
// ends, and returns what the renderer sent the agent so far.
func runBook(t *testing.T, h *hottytest.Host, first string, stream bool) func() []a2ui.Outbound {
	sent, _ := storybookProgram(t, h, first, stream)
	return sent
}

func storybookProgram(t *testing.T, h *hottytest.Host, first string, stream bool) (func() []a2ui.Outbound, *tea.Program) {
	t.Helper()
	var mu sync.Mutex
	var sent []a2ui.Outbound
	m := newModel(storybook.Options{First: first, Stream: stream, Source: "a test", Out: func(o a2ui.Outbound) {
		mu.Lock()
		defer mu.Unlock()
		sent = append(sent, o)
	}})
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
	return func() []a2ui.Outbound {
		mu.Lock()
		defer mu.Unlock()
		return append([]a2ui.Outbound(nil), sent...)
	}, p
}

// eventually waits for cond, a few seconds at most.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	for end := time.Now().Add(5 * time.Second); time.Now().Before(end); time.Sleep(5 * time.Millisecond) {
		if cond() {
			return
		}
	}
	t.Fatalf("never: %s", what)
}

// TestOnHost: on a host, the storybook's own surfaces and the story's are
// HOTTY surfaces. A click in nav opens a story; the user's typing and
// Enter reach the agent as the Form's action, and the panel lists it.
func TestOnHost(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	sent := runBook(t, h, "hotty/form", false)
	eventually(t, "nav, the story and panel on the host", func() bool {
		return h.Surface(navID) != nil && h.Surface(panelID) != nil && h.Surface("s1-0-h") != nil
	})
	story := "s1-0-h"
	if got := h.Surface(story).Text(); !strings.Contains(got, "Sign up") {
		t.Fatalf("the story shows %q", got)
	}
	if err := h.Fill(story, "name", "Ada"); err != nil {
		t.Fatal(err)
	}
	if err := h.Fill(story, "email", "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := h.Check(story, "agree", true); err != nil {
		t.Fatal(err)
	}
	if err := h.Submit(story, "email"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the signUp action", func() bool {
		for _, o := range sent() {
			if o.Action != nil && o.Action.Name == "signUp" && o.Action.Context["email"] == "ada@example.com" {
				return true
			}
		}
		return false
	})
	eventually(t, "the action in panel", func() bool {
		return strings.Contains(h.Surface(panelID).TextOf("actions"), "signUp")
	})

	// Another story, from nav: its surface replaces this one's.
	if err := h.Click(navID, "story_0"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the first story", func() bool {
		s := h.Surface("s2-0-h")
		return s != nil && s.Text() != ""
	})
	if h.Surface(story) != nil {
		t.Error("the story before is still on the host")
	}
}

// TestInCells: in a terminal that is no host, the storybook draws
// everything in cells, and the keyboard goes from pane to pane by Tab.
func TestInCells(t *testing.T) {
	h := hottytest.New(t, hottytest.Text(), hottytest.Size(120, 40))
	sent := runBook(t, h, "hotty/shortcut-press", false)
	eventually(t, "the story in cells", func() bool { return strings.Contains(h.Screen(), "A note") })
	screen := h.Screen()
	for _, want := range []string{"HOTTY kit storybook", "Rendition", "Actions"} {
		if !strings.Contains(screen, want) {
			t.Errorf("the screen lacks %q:\n%s", want, screen)
		}
	}
	// A click on the note's field gives it the keyboard; what is typed
	// goes in it, and Control+s presses Save with the note current.
	row := lineOf(screen, "Note") + 1
	h.Type("\x1b[<0;36;" + itoa(row+1) + "M\x1b[<0;36;" + itoa(row+1) + "m")
	h.Type("hello")
	eventually(t, "the typing in the field", func() bool { return strings.Contains(h.Screen(), "hello") })
	h.Type("\x13") // Control+s
	eventually(t, "the save action", func() bool {
		for _, o := range sent() {
			if o.Action != nil && o.Action.Name == "save" && o.Action.Context["note"] == "hello" {
				return true
			}
		}
		return false
	})
	eventually(t, "the action in the panel", func() bool { return strings.Contains(h.Screen(), `"name":"save"`) })
}

func lineOf(screen, s string) int {
	for i, l := range strings.Split(screen, "\n") {
		if strings.Contains(l, s) {
			return i
		}
	}
	return -1
}

func itoa(n int) string { return strconv.Itoa(n) }

// TestRenditionSwitch: the rendition changes and the story keeps its
// state: side by side, the cells beside the surface show what was typed
// in it; then in cells alone, the surface leaves the host.
func TestRenditionSwitch(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	runBook(t, h, "hotty/form", false)
	eventually(t, "the story on the host", func() bool { return h.Surface("s1-0-h") != nil })
	if err := h.Fill("s1-0-h", "name", "Ada Lovelace"); err != nil {
		t.Fatal(err)
	}
	pickRendition(t, h, 3)
	eventually(t, "the cells beside the surface", func() bool { return strings.Contains(h.Screen(), "Ada Lovelace") })
	if h.Surface("s1-0-h") == nil {
		t.Fatal("side by side lost the surface")
	}
	pickRendition(t, h, 1)
	eventually(t, "cells alone", func() bool { return h.Surface("s1-0-h") == nil })
	if !strings.Contains(h.Screen(), "Ada Lovelace") {
		t.Errorf("the cells lost the name:\n%s", h.Screen())
	}
}

// TestSideInput: side by side, the cells beside the surface take input
// too. A click gives their field the keyboard, the surface on the host
// gives it up, and what is typed shows on the host as well.
func TestSideInput(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	runBook(t, h, "hotty/form", false)
	eventually(t, "the story on the host", func() bool { return h.Surface("s1-0-h") != nil && h.Surface(navID) != nil })
	pickRendition(t, h, 3)
	eventually(t, "the cells beside the surface", func() bool { return lineOf(h.Screen(), "Name") >= 0 })
	// Take the keyboard on the host first: the cells must win it back.
	if err := h.Fill("s1-0-h", "email", "ada@example.com"); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(h.Screen(), "\n")
	row := lineOf(h.Screen(), "Name")
	col := utf8.RuneCountInString(lines[row][:strings.Index(lines[row], "Name")])
	click := "\x1b[<0;" + itoa(col+1) + ";" + itoa(row+2) + "M\x1b[<0;" + itoa(col+1) + ";" + itoa(row+2) + "m"
	h.Type(click)
	eventually(t, "the host's surface without the keyboard", func() bool { return h.Surface("s1-0-h").Focused() == "" })
	h.Type("Ada")
	eventually(t, "the typing on the host", func() bool {
		v, _ := h.Surface("s1-0-h").Value("name")
		return v == "Ada"
	})
	if v, _ := h.Surface("s1-0-h").Value("email"); v != "ada@example.com" {
		t.Errorf("the host's email is %q", v)
	}
}

// TestStream: what an agent streams shows as it comes, under the stream's
// entry, and the user's actions go back.
func TestStream(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	sent, p := storybookProgram(t, h, "", true)
	msgs := []string{
		`{"version":"v1.0","createSurface":{"surfaceId":"live","catalogId":"https://a2ui.org/specification/v1_0/catalogs/basic/catalog.json"}}`,
		`{"version":"v1.0","updateComponents":{"surfaceId":"live","components":[{"id":"root","component":"Column","children":["t","b"]},{"id":"t","component":"Text","text":"Streamed"},{"id":"b","component":"Button","child":"bt","action":{"event":{"name":"ok"}}},{"id":"bt","component":"Text","text":"OK"}]}}`,
	}
	for _, m := range msgs {
		p.Send(streamMsg{json.RawMessage(m)})
	}
	eventually(t, "the streamed surface", func() bool {
		s := h.Surface("s1-0-h")
		return s != nil && strings.Contains(s.Text(), "Streamed")
	})
	if err := h.Click("s1-0-h", "b"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the action back", func() bool {
		for _, o := range sent() {
			if o.Action != nil && o.Action.Name == "ok" && o.Action.SurfaceID == "live" {
				return true
			}
		}
		return false
	})
}

// pickRendition picks pick's ith rendition: the select opens its list, a
// surface of its own.
func pickRendition(t *testing.T, h *hottytest.Host, i int) {
	t.Helper()
	opt := "rend~o" + strconv.Itoa(i)
	list := pickID + "-list"
	eventually(t, "pick", func() bool { s := h.Surface(pickID); return s != nil && s.Placed() })
	eventually(t, "pick's rendition", func() bool { _, ok := h.Surface(pickID).Element("rend"); return ok })
	if err := h.Click(pickID, "rend"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "pick's renditions", func() bool {
		s := h.Surface(list)
		if s == nil || !s.Placed() {
			return false
		}
		_, ok := s.Element(opt)
		return ok
	})
	if err := h.Click(list, opt); err != nil {
		t.Fatal(err)
	}
}

// TestSignalDeletesSurfaces: kill's TERM ends the storybook as Control+C
// does, its surfaces deleted from the host first.
func TestSignalDeletesSurfaces(t *testing.T) {
	h := hottytest.New(t, hottytest.Size(120, 40))
	_, p := storybookProgram(t, h, "hotty/form", false)
	endOnSignal(p)
	eventually(t, "the storybook on the host", func() bool { return h.Surface(navID) != nil && len(h.Surfaces()) > 3 })
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	eventually(t, "no surface left", func() bool { return len(h.Surfaces()) == 0 })
}
