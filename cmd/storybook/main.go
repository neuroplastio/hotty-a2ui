// Command storybook shows A2UI surfaces in a terminal, as the HOTTY kit
// renders them (NEIO-11): A2UI's basic examples, the hotty catalog's
// stories, and the fallbacks; or A2UI streamed from a file or stdin, by
// an agent, with what the user does written back.
//
//	storybook                         the stories, in the terminal
//	storybook 00_simple-login-form    one story
//	storybook -stream - -out actions.jsonl
//	                                  what an agent streams on stdin
//	storybook -text 36_modal | cat    a story as plain text
//	storybook -html 36_modal          a story as the document a host gets
//	storybook -page 24_recipe-card hotty/tree > page.html
//	                                  stories as an ordinary web page (page mode)
//	storybook -theme nord             in a theme (storybook -list names them)
//	storybook -bare hotty/keys        a story's surfaces alone, in cells
//	storybook -list                   the stories' names
//
// With no terminal on stdout, or with -text, it prints each surface as
// plain text: a story once, a stream after every message.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/term"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/rendition/html"
	"github.com/neuroplastio/hotty-a2ui/rendition/text"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/story"
	"github.com/neuroplastio/hotty-a2ui/storybook"
)

func main() {
	list := flag.Bool("list", false, "list the stories")
	plain := flag.Bool("text", false, "print plain text, with no terminal (as when stdout is not one)")
	doc := flag.Bool("html", false, "print each surface as the HTML document a HOTTY host gets")
	page := flag.Bool("page", false, "print a web page of the stories' surfaces (or -stream's), in page mode: the kit's stylesheet once, then each surface")
	themeName := flag.String("theme", "", "paint in this `theme`: "+themeNames())
	stream := flag.String("stream", "", "read A2UI from `file` (- for stdin) instead of a story")
	rend := flag.String("rendition", "", "start in this `rendition`: surfaces, cells, text or side (surfaces beside cells)")
	out := flag.String("out", "", "write what the renderer sends the agent (actions, errors, function calls) to `file` as JSON lines (- for stdout)")
	bare := flag.Bool("bare", false, "show one story's surfaces alone, in cells, on the whole screen: no list, pickers or panel")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: storybook [flags] [story]")
		flag.PrintDefaults()
	}
	// Flags may come after the story too.
	var args []string
	for rest := os.Args[1:]; ; rest = flag.Args()[1:] {
		_ = flag.CommandLine.Parse(rest)
		if flag.NArg() == 0 {
			break
		}
		args = append(args, flag.Arg(0))
	}
	th := theme.Default
	if *themeName != "" {
		t, ok := theme.ByName(*themeName)
		if !ok {
			fmt.Fprintf(os.Stderr, "storybook: no theme %q: %s\n", *themeName, themeNames())
			os.Exit(2)
		}
		th = t
	}
	if *doc {
		if err := printHTML(os.Stdout, args, th); err != nil {
			fmt.Fprintln(os.Stderr, "storybook:", err)
			os.Exit(1)
		}
		return
	}
	if *page {
		if err := printPage(os.Stdout, args, *stream, th); err != nil {
			fmt.Fprintln(os.Stderr, "storybook:", err)
			os.Exit(1)
		}
		return
	}
	if *bare {
		var st *story.Story
		if len(args) == 1 {
			st = story.Find(args[0])
		}
		if err := runBare(st, th); err != nil {
			fmt.Fprintln(os.Stderr, "storybook:", err)
			os.Exit(1)
		}
		return
	}
	if err := run(*list, *plain, *stream, *out, *rend, th, args); err != nil {
		fmt.Fprintln(os.Stderr, "storybook:", err)
		os.Exit(1)
	}
}

func run(list, plain bool, stream, outPath, rend string, th theme.Theme, args []string) error {
	if list {
		for _, st := range story.All() {
			fmt.Printf("%s\t%s\n", st.Name, st.Title)
		}
		return nil
	}
	if len(args) > 1 || len(args) == 1 && stream != "" {
		flag.Usage()
		return errors.New("one story, or -stream")
	}
	var st *story.Story
	if len(args) == 1 {
		if st = story.Find(args[0]); st == nil {
			return fmt.Errorf("no story %q (storybook -list)", args[0])
		}
	}
	var in io.Reader
	switch stream {
	case "":
	case "-":
		in = os.Stdin
	default:
		f, err := os.Open(stream)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}
	w, closeOut, err := output(outPath)
	if err != nil {
		return err
	}
	defer closeOut()
	stdoutTTY := term.IsTerminal(os.Stdout.Fd())
	if plain || !stdoutTTY && outPath != "-" {
		if outPath == "-" {
			return errors.New("-out - needs stdout for the text: name a file")
		}
		return printText(os.Stdout, st, in, w)
	}
	if outPath == "-" && stdoutTTY {
		return errors.New("-out - writes to stdout, which is the terminal: pipe it, or name a file")
	}
	return interactive(st, in, stream, rend, th, w)
}

// interactive runs the storybook in the terminal: /dev/tty when stdin
// is the stream or stdout carries the actions, as with an agent on both
// ends of a pipe.
func interactive(st *story.Story, in io.Reader, source, rend string, th theme.Theme, out func(a2ui.Outbound)) error {
	termIn, termOut := os.Stdin, os.Stdout
	if in == os.Stdin || !term.IsTerminal(os.Stdin.Fd()) || !term.IsTerminal(os.Stdout.Fd()) {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			return fmt.Errorf("no terminal for the storybook: %w", err)
		}
		defer tty.Close()
		if in == os.Stdin || !term.IsTerminal(os.Stdin.Fd()) {
			termIn = tty
		}
		if !term.IsTerminal(os.Stdout.Fd()) {
			termOut = tty
		}
	}
	first := ""
	if st != nil {
		first = st.Name
	}
	if source == "-" {
		source = "stdin"
	}
	m := newModel(storybook.Options{First: first, Stream: in != nil, Source: source, Out: out, Rendition: rend, Theme: th})
	p := tea.NewProgram(m, tea.WithInput(termIn), tea.WithOutput(m.s.WatchFile(termOut)), tea.WithoutSignalHandler())
	m.s.Attach(p.Send)
	endOnSignal(p)
	if in != nil {
		go readStream(in, p.Send)
	}
	_, err := p.Run()
	return err
}

// output is where the renderer's messages go: a JSON line each.
func output(path string) (func(a2ui.Outbound), func(), error) {
	var f io.Writer
	closeFn := func() {}
	switch path {
	case "":
		return nil, closeFn, nil
	case "-":
		f = os.Stdout
	default:
		file, err := os.Create(path)
		if err != nil {
			return nil, nil, err
		}
		f, closeFn = file, func() { _ = file.Close() }
	}
	var mu sync.Mutex
	enc := json.NewEncoder(f)
	return func(o a2ui.Outbound) {
		mu.Lock()
		defer mu.Unlock()
		_ = enc.Encode(o)
	}, closeFn, nil
}

// printText is the rendition for a pipe: a story's surfaces once, or a
// stream's surfaces each time a message changes them, each under its id
// when there may be more than one.
func printText(w io.Writer, st *story.Story, in io.Reader, out func(a2ui.Outbound)) error {
	run := story.NewRun()
	run.Out = out
	if in == nil {
		if st == nil {
			return errors.New("a story, or -stream: storybook -list names them")
		}
		for _, m := range st.Messages {
			if err := run.Feed(m); err != nil {
				return err
			}
		}
		ss := run.Surfaces()
		for i, s := range ss {
			if len(ss) > 1 {
				if i > 0 {
					fmt.Fprintln(w)
				}
				fmt.Fprintln(w, "── "+s.S.ID)
			}
			fmt.Fprint(w, text.Render(s.C.V))
		}
		return nil
	}
	var changed []*story.Surface
	run.Changed = func(s *story.Surface, deleted bool) {
		if deleted {
			fmt.Fprintln(w, "── "+s.S.ID+" (deleted)")
			return
		}
		changed = append(changed, s)
	}
	return run.Stream(in, func(err error) {
		if err != nil {
			fmt.Fprintln(os.Stderr, "storybook:", strings.TrimPrefix(err.Error(), "a2ui: "))
		}
		for _, s := range changed {
			fmt.Fprintln(w, "── "+s.S.ID)
			fmt.Fprint(w, text.Render(s.C.V))
		}
		changed = changed[:0]
	})
}

// printHTML prints a story's surfaces as the documents a host gets, each
// after an HTML comment with its id: what `hotty render` draws.
func printHTML(w io.Writer, args []string, th theme.Theme) error {
	if len(args) != 1 {
		return errors.New("-html takes one story: storybook -list names them")
	}
	st := story.Find(args[0])
	if st == nil {
		return fmt.Errorf("no story %q (storybook -list)", args[0])
	}
	run := story.NewRun()
	for _, m := range st.Messages {
		if err := run.Feed(m); err != nil {
			return err
		}
	}
	for _, s := range run.Surfaces() {
		r := html.New(s.C, s.S.ID)
		r.SetTheme(th)
		fmt.Fprintf(w, "<!-- %s -->\n%s\n", s.S.ID, r.Doc())
	}
	return nil
}

func themeNames() string {
	var names []string
	for _, t := range theme.All {
		names = append(names, strings.ToLower(t.Name))
	}
	return strings.Join(names, ", ")
}
