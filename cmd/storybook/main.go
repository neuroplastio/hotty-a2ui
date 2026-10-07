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

	"github.com/charmbracelet/x/term"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/rendition/text"
	"github.com/neuroplastio/hotty-a2ui/story"
)

func main() {
	list := flag.Bool("list", false, "list the stories")
	plain := flag.Bool("text", false, "print plain text, with no terminal (as when stdout is not one)")
	stream := flag.String("stream", "", "read A2UI from `file` (- for stdin) instead of a story")
	out := flag.String("out", "", "write what the renderer sends the agent (actions, errors, function calls) to `file` as JSON lines (- for stdout)")
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
	if err := run(*list, *plain, *stream, *out, args); err != nil {
		fmt.Fprintln(os.Stderr, "storybook:", err)
		os.Exit(1)
	}
}

func run(list, plain bool, stream, outPath string, args []string) error {
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
	if plain || !term.IsTerminal(os.Stdout.Fd()) {
		if outPath == "-" {
			return errors.New("-out - needs stdout for the text: name a file")
		}
		return printText(os.Stdout, st, in, w)
	}
	return errors.New("the interactive storybook is not built yet: use -text")
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
