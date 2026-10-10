// Command review shows pictures one at a time, beside a box for feedback
// on each, on a HOTTY host: an A2UI surface that the kit renders, with the
// pictures sent to the host in band (SPEC §7.1) and shown by cid: URLs.
// The feedback is kept in a Markdown file, a section for each picture,
// where whoever asked for it reads it. Whoever asks says what to look for
// in each the same way, in brief.md beside the pictures: a section headed
// `## <name>`, shown under the picture. For a review that mustn't know
// which side of a picture is which, scripts/blind-shots.sh makes blind
// copies to show instead.
//
//	review                  the pictures in this directory
//	review .shots           in another
//	review a.png b.png      these
//	review -out notes.md    the feedback in notes.md (feedback.md beside the pictures without it)
//	review -brief ask.md    what to look for, from ask.md (brief.md beside the pictures without it)
//
// The list picks the picture; Enter on it goes to the feedback box.
// Control+s saves, and so does moving to another picture or quitting.
// Control+c quits, as does q with nothing focused.
package main

import (
	"errors"
	"flag"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	tea "charm.land/bubbletea/v2"

	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
)

func main() {
	out := flag.String("out", "", "keep the feedback in `file` (default: feedback.md beside the pictures)")
	brief := flag.String("brief", "", "what to look for in each picture, from `file` (default: brief.md beside the pictures)")
	themeName := flag.String("theme", "", "paint in this `theme`; the host's own colours without it")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: review [flags] [directory | picture...]")
		flag.PrintDefaults()
	}
	// Flags may come after the pictures too.
	var args []string
	for rest := os.Args[1:]; ; rest = flag.Args()[1:] {
		_ = flag.CommandLine.Parse(rest)
		if flag.NArg() == 0 {
			break
		}
		args = append(args, flag.Arg(0))
	}
	if err := run(args, *out, *brief, *themeName); err != nil {
		fmt.Fprintln(os.Stderr, "review:", err)
		os.Exit(1)
	}
}

func run(args []string, out, brief, themeName string) error {
	var th theme.Theme
	if themeName != "" {
		t, ok := theme.ByName(themeName)
		if !ok {
			return fmt.Errorf("no theme %q", themeName)
		}
		th = t
	}
	shots, dir, err := find(args)
	if err != nil {
		return err
	}
	if out == "" {
		out = filepath.Join(dir, "feedback.md")
	}
	if brief == "" {
		brief = filepath.Join(dir, "brief.md")
	}
	n, err := loadNotes(out)
	if err != nil {
		return err
	}
	a, err := newApp(shots, n, brief, th)
	if err != nil {
		return err
	}
	p := tea.NewProgram(a, tea.WithOutput(a.s.WatchFile(os.Stdout)), tea.WithoutSignalHandler())
	a.s.Attach(p.Send)
	endOnSignal(p)
	_, err = p.Run()
	return err
}

// The pictures review shows, by extension, and their types.
var mimes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
}

// find is the pictures the arguments name: a directory's, in name order,
// or files; this directory's without any. dir is where the feedback goes
// by default: the first directory named, or the first file's.
func find(args []string) (shots []shot, dir string, err error) {
	if len(args) == 0 {
		args = []string{"."}
	}
	var paths []string
	for _, arg := range args {
		fi, err := os.Stat(arg)
		if err != nil {
			return nil, "", err
		}
		if !fi.IsDir() {
			if dir == "" {
				dir = filepath.Dir(arg)
			}
			paths = append(paths, arg)
			continue
		}
		if dir == "" {
			dir = arg
		}
		entries, err := os.ReadDir(arg)
		if err != nil {
			return nil, "", err
		}
		for _, e := range entries {
			if _, ok := mimes[strings.ToLower(filepath.Ext(e.Name()))]; ok && !e.IsDir() {
				paths = append(paths, filepath.Join(arg, e.Name()))
			}
		}
	}
	for _, p := range paths {
		mime, ok := mimes[strings.ToLower(filepath.Ext(p))]
		if !ok {
			return nil, "", fmt.Errorf("%s: not a picture review shows (%s)", p, strings.Join(slices.Sorted(maps.Keys(mimes)), " "))
		}
		name, err := filepath.Rel(dir, p)
		if err != nil || strings.HasPrefix(name, "..") {
			name = p
		}
		shots = append(shots, shot{name: filepath.ToSlash(name), path: p, mime: mime, id: fmt.Sprintf("shot-%d", len(shots))})
	}
	if len(shots) == 0 {
		return nil, "", errors.New("no pictures in " + strings.Join(args, " "))
	}
	return shots, dir, nil
}

// endSignal is a signal to end (endOnSignal).
type endSignal struct{}

// endOnSignal ends the program on TERM, HUP or INT as Control+C does: the
// feedback saved and the surface deleted first, where Bubble Tea's own
// handling would leave it on the host (SPEC §5.4).
func endOnSignal(p *tea.Program) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		<-sig
		signal.Stop(sig)
		p.Send(endSignal{})
	}()
}
