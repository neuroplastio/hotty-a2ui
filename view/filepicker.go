package view

import (
	"cmp"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// A HottyFilePicker (profile §6.31) in the view: the directory of a root
// the program grants, its entries listed, and the entry the keyboard is
// on. The renderer reads the filesystem; the agent never does, and a
// picked path reaches it only through the data model.

// Grant gives a surface's file picker a root to list, by name, or takes
// it back (nil): the program's act, the only one that names a filesystem
// (profile §6.31). A HottyFilePicker whose `root` this does not name
// shows its error. The picker reads no path outside fsys, but fsys may
// follow a symbolic link out of its directory, as os.DirFS does: grant
// os.OpenRoot(dir) and its FS, which do not. Every picker reads its
// directory again at its next build.
func (c *Controller) Grant(name string, fsys fs.FS) {
	if fsys == nil {
		delete(c.St.Roots, name)
	} else {
		if c.St.Roots == nil {
			c.St.Roots = map[string]fs.FS{}
		}
		c.St.Roots[name] = fsys
	}
	for id, st := range c.St.File {
		st.Listing = nil
		c.St.File[id] = st
	}
}

// cleanFile is a directory path relative to a root, with no `.` or `..`
// that leaves it: a path the agent gives is clamped inside the root, as
// its grant says (L10).
func cleanFile(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	return strings.TrimPrefix(path.Clean("/"+p), "/")
}

// joinFile is a name under a directory, relative to the root.
func joinFile(dir, name string) string { return cleanFile(path.Join(dir, name)) }

// readFiles reads directory dir of the root a picker names: its entries,
// directories first, then files, each by name, case aside; or why it
// lists nothing instead: no root, one not granted, or a directory it
// cannot read.
func readFiles(roots map[string]fs.FS, root, dir string) *FileListing {
	l := &FileListing{Root: root, Dir: dir}
	fsys := roots[root]
	switch {
	case root == "":
		l.Failed = "A file picker needs a root."
		return l
	case fsys == nil:
		l.Failed = fmt.Sprintf("No root named %q.", root)
		return l
	}
	ents, err := fs.ReadDir(fsys, cmp.Or(dir, "."))
	if err != nil {
		l.Failed = "Cannot read this directory."
		return l
	}
	for _, de := range ents {
		f := FileEntry{Name: de.Name(), Dir: de.IsDir(), Size: -1}
		if info, err := de.Info(); err == nil && !f.Dir {
			f.Size = info.Size()
		}
		l.Files = append(l.Files, f)
	}
	slices.SortFunc(l.Files, func(a, b FileEntry) int {
		if a.Dir != b.Dir {
			if a.Dir {
				return -1
			}
			return 1
		}
		return cmp.Or(cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), cmp.Compare(a.Name, b.Name))
	})
	return l
}

// filterFiles are the entries whose names hold the filter, case aside.
func filterFiles(files []FileEntry, filter string) []FileEntry {
	if filter == "" {
		return files
	}
	filter = strings.ToLower(filter)
	var out []FileEntry
	for _, f := range files {
		if strings.Contains(strings.ToLower(f.Name), filter) {
			out = append(out, f)
		}
	}
	return out
}

// mapFilePicker makes a HottyFilePicker's element: its root, the
// directory it shows, the entries the filter leaves of those it read
// (readFiles, FileListing) or its error, the entry it highlights, and the
// rows a height shows. Each directory shows
// the highlight and the rows it was left with, whoever changed the
// directory; one it has not shown highlights its first entry, as
// bubbles' filepicker does. A `path` not bound is where it starts.
func mapFilePicker(b *Builder, n *a2ui.Node) *Element {
	if b.St.File == nil {
		b.St.File = map[string]FilePick{}
	}
	st := b.St.File[n.Key]
	if st.Dirs == nil {
		st.Dirs = map[string]DirPick{}
		st.Dir = cleanFile(a2ui.ToString(b.Raw(n, "path")))
	}
	e := &Element{Kind: FilePicker, Root: b.String(n, "root"), FileDir: st.Dir,
		Placeholder: cmp.Or(b.String(n, "emptyText"), "Empty."),
		Value:       a2ui.ToString(b.Value(n, "selected")),
		Query:       Query{Text: b.String(n, "filter")}}
	if bd, ok := n.Props["path"].(a2ui.Bound); ok && bd.Writable() {
		e.FileDir = cleanFile(a2ui.ToString(b.Value(n, "path")))
	}
	if h := a2ui.ToNumber(b.Raw(n, "height")); h >= 1 {
		e.Height = int(h)
	}
	left := st.Dirs[e.FileDir]
	if e.FileDir != st.Shown {
		b.St.Scroll[n.Key] = left.Top
		st.Shown = e.FileDir
	}
	if l := st.Listing; l == nil || l.Root != e.Root || l.Dir != e.FileDir {
		st.Listing = readFiles(b.St.Roots, e.Root, e.FileDir)
	}
	e.Files, e.Error = filterFiles(st.Listing.Files, e.Query.Text), st.Listing.Failed
	e.Selected = slices.IndexFunc(e.Files, func(f FileEntry) bool { return f.Name == left.Name })
	if e.Selected < 0 && len(e.Files) > 0 {
		e.Selected = 0
	}
	if e.Height > 0 && e.Selected >= 0 {
		top := b.St.Scroll[n.Key]
		if at := joinFile(e.FileDir, e.Files[e.Selected].Name); b.St.Revealed[n.Key] != at {
			top = min(top, e.Selected)
			top = max(top, e.Selected-e.Height+1)
			b.St.Revealed[n.Key] = at
		}
		e.Top = max(min(top, len(e.Files)-e.Height), 0)
		b.St.Scroll[n.Key] = e.Top
	}
	st.Dirs[e.FileDir] = DirPick{Name: e.File().Name, Top: e.Top}
	b.St.File[n.Key] = st
	return e
}

// File is the entry a HottyFilePicker highlights; none (no name) when it
// lists nothing.
func (e *Element) File() FileEntry {
	if e.Selected < 0 || e.Selected >= len(e.Files) {
		return FileEntry{}
	}
	return e.Files[e.Selected]
}

// Label is an entry's name as a HottyFilePicker lists it: a directory's
// with a trailing "/".
func (f FileEntry) Label() string {
	if f.Dir {
		return f.Name + "/"
	}
	return f.Name
}

// SizeText is a file's size as bubbles' filepicker shows it: bytes, then
// K, M, G and T to one decimal; "" for a directory, or a size not known.
func (f FileEntry) SizeText() string {
	if f.Dir || f.Size < 0 {
		return ""
	}
	if f.Size < 1024 {
		return strconv.FormatInt(f.Size, 10) + "B"
	}
	n := float64(f.Size)
	for _, u := range []string{"K", "M", "G"} {
		if n /= 1024; n < 1024 {
			return strconv.FormatFloat(n, 'f', 1, 64) + u
		}
	}
	return strconv.FormatFloat(n/1024, 'f', 1, 64) + "T"
}

// ClickFile is a click on a HottyFilePicker's entry i: it highlights it,
// or opens or picks it when it was highlighted already, as a second click
// does in a HottyList.
func (c *Controller) ClickFile(id string, i int) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != FilePicker || i < 0 || i >= len(e.Files) {
		return nil
	}
	if i != e.Selected {
		c.fileSelect(e, i)
		return nil
	}
	return c.fileOpen(e)
}

// FileKey works a HottyFilePicker by a key, named as SPEC §10.4 has it,
// as bubbles' filepicker takes them: the arrows up and down (j and k)
// move a row, Page Up and Page Down by the rows it shows, Home and End
// (g and G) to the first entry and the last; the arrow right, Enter and
// l open a directory or pick a file; the arrow left, Backspace, h and
// Escape go up to the parent directory. ok reports whether the key is one
// of those; at the root, those that go up are not, and pass on (Escape to
// close a Modal, say).
func (c *Controller) FileKey(id, key string) (ok bool, err error) {
	e := c.V.Find(id)
	if e == nil || e.Kind != FilePicker {
		return false, nil
	}
	page := e.Height
	if page <= 0 {
		page = len(e.Files)
	}
	switch key {
	case "Enter", "ArrowRight", "l":
		return true, c.fileOpen(e)
	case "ArrowLeft", "Backspace", "h", "Escape":
		if e.FileDir == "" {
			return false, nil
		}
		return true, c.fileUp(e)
	}
	to := e.Selected
	switch key {
	case "ArrowUp", "k":
		to--
	case "ArrowDown", "j":
		to++
	case "PageUp":
		to -= page
	case "PageDown":
		to += page
	case "Home", "g":
		to = 0
	case "End", "G":
		to = len(e.Files) - 1
	default:
		return false, nil
	}
	if len(e.Files) > 0 {
		c.fileSelect(e, min(max(to, 0), len(e.Files)-1))
	}
	return true, nil
}

// fileSelect highlights entry i and rebuilds, so that every rendition
// shows the same entry.
func (c *Controller) fileSelect(e *Element, i int) {
	c.fileHighlight(e.ID, e.FileDir, e.Files[i].Name)
	c.Rebuild()
}

// fileHighlight makes the entry named name the one a HottyFilePicker
// highlights in directory dir, as it next shows it.
func (c *Controller) fileHighlight(id, dir, name string) {
	st := c.St.File[id]
	if st.Dirs == nil {
		st.Dirs = map[string]DirPick{}
	}
	d := st.Dirs[dir]
	d.Name = name
	st.Dirs[dir] = d
	c.St.File[id] = st
}

// fileOpen opens the highlighted entry: a directory is entered (fileGo),
// a file is picked: its path goes to `selected`, and onPick runs (its
// context reads the path, an action carrying no payload).
func (c *Controller) fileOpen(e *Element) error {
	f := e.File()
	switch {
	case f.Name == "":
		return nil
	case f.Dir:
		return c.fileGo(e, joinFile(e.FileDir, f.Name), "")
	}
	if err := c.setProp(e, "selected", joinFile(e.FileDir, f.Name)); err != nil {
		return err
	}
	c.Rebuild()
	if n := c.V.Node(e.ID); n != nil && n.Props["onPick"] != nil {
		return c.invoke(n, "onPick", true)
	}
	return nil
}

// FileUp goes to the parent of the directory shown in HottyFilePicker id
// (fileUp), as a click on its title does.
func (c *Controller) FileUp(id string) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != FilePicker {
		return nil
	}
	return c.fileUp(e)
}

// fileUp goes to the parent of the directory shown, highlighting the
// directory it left; at the root, nothing happens.
func (c *Controller) fileUp(e *Element) error {
	if e.FileDir == "" {
		return nil
	}
	return c.fileGo(e, cleanFile(path.Dir(e.FileDir)), path.Base(e.FileDir))
}

// fileGo shows directory dir of the root, highlighting the entry named
// from there when it is not "": it writes the directory to the bound
// `path` (else to the renderer's state), rebuilds, and runs onChange
// (whose context reads `path`).
func (c *Controller) fileGo(e *Element, dir, from string) error {
	if from != "" {
		c.fileHighlight(e.ID, dir, from)
	}
	n := c.V.Node(e.ID)
	if bd, ok := n.Props["path"].(a2ui.Bound); ok && bd.Writable() {
		if err := c.S.Write(bd.Path, dir); err != nil {
			return err
		}
	} else {
		st := c.St.File[e.ID]
		st.Dir = dir
		c.St.File[e.ID] = st
	}
	c.Rebuild()
	if n.Props["onChange"] != nil {
		return c.invoke(n, "onChange", true)
	}
	return nil
}
