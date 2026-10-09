// Command ref shows what the kit's components are held against (vault
// KIT-REF): Bubble Tea's bubbles, huh and lipgloss, each with the content
// of the kit's story for the same component, so that a shot of one can be
// set beside a shot of the other (scripts/ref-shot.sh), and both can be
// tried by hand.
//
//	ref -list      the references, with the story each is held against
//	ref form       one, on the whole screen; Control+C quits
//
// It is a module of its own, so that bubbles, huh and lipgloss stay out of
// the kit's go.mod.
package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// A reference is a Bubble Tea program, and the kit's story it is held
// against.
type reference struct {
	story string
	about string
	model func() tea.Model
}

var refs = map[string]reference{}

func register(name, story, about string, model func() tea.Model) {
	refs[name] = reference{story: story, about: about, model: model}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: ref -list | ref <name>")
		os.Exit(2)
	}
	if os.Args[1] == "-list" {
		var names []string
		for n := range refs {
			names = append(names, n)
		}
		slices.Sort(names)
		for _, n := range names {
			fmt.Printf("%s\t%s\t%s\n", n, refs[n].story, refs[n].about)
		}
		return
	}
	r, ok := refs[os.Args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "ref: no %q (ref -list)\n", os.Args[1])
		os.Exit(2)
	}
	if _, err := tea.NewProgram(r.model()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "ref:", err)
		os.Exit(1)
	}
}

// screen is a model drawn on the alternate screen, which quits on
// Control+C; what it holds draws the rest.
type screen struct {
	inner interface {
		Init() tea.Cmd
		Update(tea.Msg) tea.Cmd
		View() string
	}
}

func (s screen) Init() tea.Cmd { return s.inner.Init() }

func (s screen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "ctrl+c" {
		return s, tea.Quit
	}
	return s, s.inner.Update(msg)
}

func (s screen) View() tea.View {
	v := tea.NewView(strings.TrimRight(s.inner.View(), "\n"))
	v.AltScreen = true
	return v
}
