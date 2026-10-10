package main

import (
	"errors"
	"fmt"
	stdhtml "html"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/rendition/html"
	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
	"github.com/neuroplastio/hotty-a2ui/story"
)

// pageCSS is the page's own: a column of the stories, each under its
// title. It sets no palette, so the kit's default shows (profile §2.1),
// light or dark as the reader's browser prefers.
const pageCSS = `:root { color-scheme: light dark; }
body { margin: 0; padding: 1.5rem 1rem 4rem; font: 16px/1.5 system-ui, -apple-system, "Segoe UI", sans-serif; background: Canvas; color: CanvasText; }
main { max-width: 46rem; margin: 0 auto; }
.story > h2 { margin: 2.5rem 0 1rem; font-size: 0.8125rem; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: GrayText; }
.story:first-child > h2 { margin-top: 0; }
.story > h2 code { font-weight: 400; letter-spacing: 0; text-transform: none; }
.story > .k-page + .k-page { margin-top: 1.5rem; }`

// printPage prints stories' surfaces as one web page, in page mode
// (rendition/html's Page): the kit's stylesheet once, in the head, then
// each story's title and its surfaces' markup; with a stream, its surfaces
// as they are when it ends. A surface's name on the page is its story's
// place and its id, so that two stories' surfaces share no id.
func printPage(w io.Writer, args []string, stream string, th theme.Theme) error {
	type part struct {
		title, name string
		ss          []*story.Surface
	}
	var parts []part
	switch {
	case stream != "" && len(args) > 0:
		return errors.New("-page takes stories, or -stream")
	case stream != "":
		in := io.Reader(os.Stdin)
		if stream != "-" {
			f, err := os.Open(stream)
			if err != nil {
				return err
			}
			defer f.Close()
			in = f
		}
		run := story.NewRun()
		err := run.Stream(in, func(err error) {
			if err != nil {
				fmt.Fprintln(os.Stderr, "storybook:", strings.TrimPrefix(err.Error(), "a2ui: "))
			}
		})
		if err != nil {
			return err
		}
		parts = append(parts, part{title: filepath.Base(stream), ss: run.Surfaces()})
	case len(args) == 0:
		return errors.New("-page takes stories: storybook -list names them")
	}
	for _, a := range args {
		st := story.Find(a)
		if st == nil {
			return fmt.Errorf("no story %q (storybook -list)", a)
		}
		run := story.NewRun()
		for _, m := range st.Messages {
			if err := run.Feed(m); err != nil {
				return fmt.Errorf("%s: %w", st.Name, err)
			}
		}
		parts = append(parts, part{st.Title, st.Name, run.Surfaces()})
	}
	var b strings.Builder
	b.WriteString("<!doctype html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n" +
		"<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n<title>Page mode</title>\n")
	b.WriteString("<style>\n" + html.PageCSS() + "</style>\n<style>\n" + pageCSS + "\n</style>\n</head>\n<body>\n<main>\n")
	for i, p := range parts {
		b.WriteString("<section class=\"story\"><h2>" + stdhtml.EscapeString(p.title))
		if p.name != "" {
			b.WriteString(" <code>" + stdhtml.EscapeString(p.name) + "</code>")
		}
		b.WriteString("</h2>\n")
		for _, s := range p.ss {
			r := html.New(s.C, "s"+strconv.Itoa(i)+"-"+s.S.ID)
			r.SetTheme(th)
			b.WriteString(r.Page() + "\n")
		}
		b.WriteString("</section>\n")
	}
	b.WriteString("</main>\n</body>\n</html>\n")
	_, err := io.WriteString(w, b.String())
	return err
}
