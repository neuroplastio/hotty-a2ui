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
// box holding its title (the directory path from the root; in a
// subdirectory, a button up to the parent) and its entries (directories
// first, then files, by name), each an option whose id is the picker's,
// "~i" and the entry's index, showing its name (a directory's with a
// trailing "/") and, for a file, its size. With a height, the host
// scrolls them.
func filePicker(e *view.Element) *node {
	path := "/" + e.FileDir
	box := el("div", "id", domID(e.ID), "class", "k-filepicker", "tabindex", "0", "role", "listbox", "data-keys", fileKeys, "aria-label", path)
	title := el("div", "id", partID(e.ID, partTitle), "class", "k-file-title")
	if e.FileDir != "" {
		title.set("class", "k-file-title k-file-up").set("data-on", "click").set("role", "button").set("aria-label", "Parent directory: "+path)
		title.add(el("span", "class", "k-file-back").add(txt("◂ ")))
	}
	box.add(title.add(el("span", "class", "k-file-path").add(texts(path)...)))

	switch {
	case e.Error != "":
		return box.add(el("div", "id", partID(e.ID, partError), "class", "k-file-error", "aria-live", "polite").add(txt("✗ " + e.Error)))
	case len(e.Files) == 0:
		return box.add(el("div", "class", "k-file-empty").add(texts(e.Placeholder)...))
	}
	list := el("div", "class", "k-file-list")
	if e.Height > 0 {
		list.set("class", "k-file-list k-scrolls").set("style", "--k-rows: "+strconv.Itoa(e.Height))
	}
	for i, f := range e.Files {
		item := el("div", "id", partID(e.ID, partItem+strconv.Itoa(i)), "class", "k-file-item", "role", "option",
			"tabindex", "-1", "aria-selected", strconv.FormatBool(i == e.Selected), "data-on", "click")
		if i == e.Selected {
			item.set("class", "k-file-item k-sel")
		}
		item.add(el("span", "class", "k-file-name").add(texts(f.Label())...))
		if s := f.SizeText(); s != "" {
			item.add(el("span", "class", "k-file-size").add(txt(s)))
		}
		list.add(item)
	}
	return box.add(list)
}
