package html

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// fileKeys are the keys a HottyFilePicker gives the program, which works
// its selection and navigation (view.Controller.FileKey): the arrows, Page
// Up, Page Down, Home and End, Enter, Backspace and Escape, and bubbles'
// letters (j, k, h, l, g, G).
const fileKeys = tableKeys + " ArrowLeft=program ArrowRight=program Backspace=program Escape=program k=program j=program h=program l=program g=program G=program"

// filePicker is a HottyFilePicker on a host (profile §2, §6.31): a focusable
// box holding its title (the directory path from the root), its entries
// (directories first, then files, by name), each an option whose id is the
// picker's, "~k" and the entry's index, showing its name and, for a file,
// its size. A directory has a trailing "/".
func filePicker(e *view.Element) *node {
	box := el("div", "id", domID(e.ID), "class", "k-filepicker", "tabindex", "0", "role", "listbox", "data-keys", fileKeys)
	path := filePath(e)
	box.set("aria-label", path)
	title := el("div", "id", partID(e.ID, partTitle), "class", "k-file-title")
	if e.FileDir != "" {
		title.set("class", "k-file-title k-file-up")
		title.set("data-on", "click")
		title.set("role", "button")
		title.set("aria-label", "Parent directory: "+path)
		title.add(el("span", "class", "k-file-back").add(txt("◂\u00a0")))
	}
	title.add(el("span", "class", "k-file-path").add(texts(path)...))
	box.add(title)

	switch {
	case e.Error != "":
		box.add(el("div", "id", partID(e.ID, partError), "class", "k-file-error", "aria-live", "polite").add(txt("✗ " + e.Error)))
		return box
	case len(e.Files) == 0:
		box.add(el("div", "class", "k-file-empty").add(texts(e.Placeholder)...))
		return box
	}

	list := el("div", "class", "k-file-list")
	if e.Height > 0 {
		list.set("class", "k-file-list k-scrolls").set("style", "--k-rows: "+strconv.Itoa(e.Height))
	}
	for i, f := range e.Files {
		item := el("div", "id", partID(e.ID, partEntry+strconv.Itoa(i)), "class", "k-file-item", "role", "option",
			"tabindex", "-1", "aria-selected", strconv.FormatBool(i == e.Selected), "data-on", "click")
		if i == e.Selected {
			item.set("class", "k-file-item k-sel")
		}
		item.add(el("span", "class", "k-file-name").add(texts(fileLabel(f))...))
		if s := fileSizeText(f); s != "" {
			item.add(el("span", "class", "k-file-size").add(txt(s)))
		}
		list.add(item)
	}
	box.add(list)
	return box
}

// filePath is the title a HottyFilePicker draws: the directory shown
// from the root, "/".
func filePath(e *view.Element) string {
	if e.FileDir == "" {
		return "/"
	}
	return "/" + e.FileDir
}

// fileLabel is an entry's name as a line: a directory with a trailing
// "/".
func fileLabel(f view.FileEntry) string {
	if f.Dir {
		return f.Name + "/"
	}
	return f.Name
}

// fileSizeText is a file's size, as bubbles' filepicker has it: bytes,
// then K, M, G, T, to one decimal from K up.
func fileSizeText(f view.FileEntry) string {
	if f.Dir || f.Size < 0 {
		return ""
	}
	n := float64(f.Size)
	for _, u := range []string{"B", "K", "M", "G"} {
		switch {
		case n < 1024:
			if u == "B" {
				return strconv.FormatInt(f.Size, 10) + "B"
			}
			return strconv.FormatFloat(n, 'f', 1, 64) + u
		}
		if u == "G" {
			return strconv.FormatFloat(n/1024, 'f', 1, 64) + "T"
		}
		n /= 1024
	}
	return ""
}
