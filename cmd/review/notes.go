package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// notes is the feedback, a note for each picture, kept in a Markdown file:
// a section for each, headed by the picture's name, in the order they
// were first written.
type notes struct {
	path  string
	order []string
	text  map[string]string
}

// loadNotes reads the feedback a file has, if it is there.
func loadNotes(path string) (*notes, error) {
	n := &notes{path: path, text: map[string]string{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return n, nil
	}
	if err != nil {
		return nil, err
	}
	name := ""
	var body []string
	end := func() {
		if name != "" {
			n.set(name, strings.Join(body, "\n"))
		}
		body = nil
	}
	for l := range strings.Lines(string(b)) {
		l = strings.TrimRight(l, "\r\n")
		if h, ok := strings.CutPrefix(l, "## "); ok {
			end()
			name = strings.TrimSpace(h)
			continue
		}
		if name != "" {
			body = append(body, l)
		}
	}
	end()
	return n, nil
}

// set is a picture's note: none when it is blank.
func (n *notes) set(name, text string) {
	text = strings.TrimSpace(text)
	_, had := n.text[name]
	switch {
	case text == "":
		delete(n.text, name)
		n.order = slices.DeleteFunc(n.order, func(s string) bool { return s == name })
	case !had:
		n.order = append(n.order, name)
		fallthrough
	default:
		n.text[name] = text
	}
}

// write writes the file whole, by a rename, so that a reader never sees
// half of it.
func (n *notes) write() error {
	var b strings.Builder
	b.WriteString("# Feedback\n")
	for _, name := range n.order {
		b.WriteString("\n## " + name + "\n\n" + n.text[name] + "\n")
	}
	tmp, err := os.CreateTemp(filepath.Dir(n.path), ".feedback-*")
	if err != nil {
		return err
	}
	if _, err := tmp.WriteString(b.String()); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), n.path)
}

// first is a note's first line, short enough for the list.
func first(text string) string {
	l, _, _ := strings.Cut(text, "\n")
	if r := []rune(l); len(r) > 48 {
		l = string(r[:47]) + "…"
	}
	return l
}
