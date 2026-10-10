package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// notes is the feedback kept in a Markdown file: a section for each
// picture, headed by its name, in the order they were first written. With
// one box a section is its note; with several (a blind pair's A and B) it
// holds a part for each box that has one, headed by the box's label:
//
//	## form.png
//
//	### A
//
//	The caret is too thin.
type notes struct {
	path  string
	order []string
	// text is each picture's notes, by box: "" when there is one box.
	text map[string]map[string]string
}

// loadNotes reads the notes a file has, if it is there.
func loadNotes(path string) (*notes, error) {
	n := &notes{path: path, text: map[string]map[string]string{}}
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return n, nil
	}
	if err != nil {
		return nil, err
	}
	name, box := "", ""
	var body []string
	end := func() {
		if name != "" {
			n.set(name, box, strings.Join(body, "\n"))
		}
		body = nil
	}
	for l := range strings.Lines(string(b)) {
		l = strings.TrimRight(l, "\r\n")
		if h, ok := strings.CutPrefix(l, "## "); ok {
			end()
			name, box = strings.TrimSpace(h), ""
			continue
		}
		if h, ok := strings.CutPrefix(l, "### "); ok && name != "" {
			end()
			box = strings.TrimSpace(h)
			continue
		}
		if name != "" {
			body = append(body, l)
		}
	}
	end()
	return n, nil
}

// get is a picture's note in a box.
func (n *notes) get(name, box string) string { return n.text[name][box] }

// set is a picture's note in a box: none when it is blank.
func (n *notes) set(name, box, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		delete(n.text[name], box)
		if len(n.text[name]) == 0 {
			delete(n.text, name)
			n.order = slices.DeleteFunc(n.order, func(s string) bool { return s == name })
		}
		return
	}
	if n.text[name] == nil {
		n.text[name] = map[string]string{}
		n.order = append(n.order, name)
	}
	n.text[name][box] = text
}

// summary is a line for the list, short so that it never widens it: the
// labels of the boxes a picture has notes in, or "noted" with one box;
// nothing without notes.
func (n *notes) summary(name string) string {
	bs := n.boxes(name)
	if len(bs) == 1 && bs[0] == "" {
		return "noted"
	}
	return strings.Join(bs, " · ")
}

// boxes are the boxes a picture has notes in: the one box's first, then
// the labels in order.
func (n *notes) boxes(name string) []string {
	var bs []string
	for box := range n.text[name] {
		bs = append(bs, box)
	}
	slices.Sort(bs)
	return bs
}

// write writes the file whole, by a rename, so that a reader never sees
// half of it.
func (n *notes) write() error {
	var b strings.Builder
	b.WriteString("# Feedback\n")
	for _, name := range n.order {
		b.WriteString("\n## " + name + "\n")
		for _, box := range n.boxes(name) {
			if box != "" {
				b.WriteString("\n### " + box + "\n")
			}
			b.WriteString("\n" + n.text[name][box] + "\n")
		}
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
