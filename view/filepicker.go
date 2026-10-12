package view

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// A HottyFilePicker (profile §6.31) in the view: the directory of a root
// the program grants, its entries listed, and the entry the keyboard is
// on. The renderer reads the filesystem; the agent never does, and a
// picked path reaches it only through the data model.

// Grant gives a surface's file picker a root to list, by name: the
// program's act, the only one that names a filesystem (profile §6.31).
// A HottyFilePicker whose `root` this does not name shows its error.
func (c *Controller) Grant(name string, fsys fs.FS) {
	if fsys == nil {
		delete(c.St.Roots, name)
		return
	}
	if c.St.Roots == nil {
		c.St.Roots = map[string]fs.FS{}
	}
	c.St.Roots[name] = fsys
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

// fsDir is a relative directory as io/fs names it: "" is ".".
func fsDir(dir string) string {
	if dir == "" {
		return "."
	}
	return dir
}

// fileMatch reports whether a filter leaves a name: a case-insensitive
// substring; every name for the empty filter.
func fileMatch(filter, name string) bool {
	if filter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(name), strings.ToLower(filter))
}

// sortFile orders a directory's entries as a file picker lists them:
// directories first, then files, each by name, case aside.
func sortFile(files []FileEntry) {
	sort.SliceStable(files, func(i, j int) bool {
		if files[i].Dir != files[j].Dir {
			return files[i].Dir
		}
		a, b := strings.ToLower(files[i].Name), strings.ToLower(files[j].Name)
		if a != b {
			return a < b
		}
		return files[i].Name < files[j].Name
	})
}

// mapFilePicker makes a HottyFilePicker's element: its root, the
// directory it shows, and the entries it lists (directories first, then
// files, by name), the entry it has highlighted, and the rows a height
// shows. The filesystem is the program's grant (Controller.Grant); a
// root it did not grant, or a directory it cannot read, is the picker's
// error.
func mapFilePicker(b *Builder, n *a2ui.Node) *Element {
	if b.St.File == nil {
		b.St.File = map[string]FilePick{}
	}
	st := b.St.File[n.Key]
	if st.Dirs == nil {
		st.Dirs = map[string]DirPick{}
	}
	e := &Element{Kind: FilePicker, Root: b.String(n, "root"),
		Placeholder: b.String(n, "emptyText"),
		Value:       a2ui.ToString(b.Value(n, "selected")),
		Query:       Query{Text: b.String(n, "filter")}}
	if e.Placeholder == "" {
		e.Placeholder = "Empty."
	}
	if bd, ok := n.Props["path"].(a2ui.Bound); ok && bd.Writable() {
		dir := cleanFile(a2ui.ToString(b.Value(n, "path")))
		if dir != st.Path {
			if saved, has := st.Dirs[dir]; has {
				st.Index = saved.Index
				b.St.Scroll[n.Key] = saved.Top
			} else {
				st.Index = -1
				b.St.Scroll[n.Key] = 0
			}
			st.Path = dir
			delete(b.St.Revealed, n.Key)
		}
		e.FileDir = dir
	} else {
		e.FileDir = cleanFile(st.Path)
		st.Path = e.FileDir
	}
	if h := a2ui.ToNumber(b.Raw(n, "height")); h >= 1 {
		e.Height = int(h)
	}
	fsys, granted := b.St.Roots[e.Root]
	switch {
	case e.Root == "":
		e.Error = "A file picker needs a root."
	case !granted || fsys == nil:
		e.Error = fmt.Sprintf("No root named %q.", e.Root)
	default:
		ents, err := fs.ReadDir(fsys, fsDir(e.FileDir))
		if err != nil {
			e.Error = "Cannot read this directory."
			break
		}
		for _, de := range ents {
			if !fileMatch(e.Query.Text, de.Name()) {
				continue
			}
			f := FileEntry{Name: de.Name(), Dir: de.IsDir(), Size: -1}
			if info, err := de.Info(); err == nil && !f.Dir {
				f.Size = info.Size()
			}
			e.Files = append(e.Files, f)
		}
		sortFile(e.Files)
	}
	e.Selected = st.Index
	if e.Selected >= len(e.Files) {
		e.Selected = len(e.Files) - 1
	}
	if e.Height > 0 && len(e.Files) > 0 {
		top := b.St.Scroll[n.Key]
		fileKey := e.FileDir + ":" + strconv.Itoa(e.Selected)
		if e.Selected >= 0 && b.St.Revealed[n.Key] != fileKey {
			top = min(top, e.Selected)
			top = max(top, e.Selected-e.Height+1)
			b.St.Revealed[n.Key] = fileKey
		}
		e.Top = max(min(top, len(e.Files)-e.Height), 0)
		b.St.Scroll[n.Key] = e.Top
	}
	st.Dirs[e.FileDir] = DirPick{Index: e.Selected, Top: e.Top}
	b.St.File[n.Key] = st
	return e
}

// FileRefs are a HottyFilePicker's entries, and the one highlighted.
func (e *Element) FileRef() FileEntry {
	if e.Selected < 0 || e.Selected >= len(e.Files) {
		return FileEntry{}
	}
	return e.Files[e.Selected]
}

// ClickFile is a click on a HottyFilePicker's entry i: it selects it, or
// opens or picks it when it was selected already, as a second click does
// in a HottyList.
func (c *Controller) ClickFile(id string, i int) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != FilePicker || i < 0 || i >= len(e.Files) {
		return nil
	}
	if i != e.Selected {
		return c.fileSelect(e, i)
	}
	return c.fileOpen(e)
}

// FileKey works a HottyFilePicker by a key, named as SPEC §10.4 has it,
// as bubbles' filepicker takes them: the arrows up and down (j and k)
// move a row, Page Up and Page Down by the rows it shows, Home and End
// (g and G) to the first entry and the last; the arrow right, Enter and
// l open a directory or pick a file; the arrow left, Backspace, h and
// Escape go up to the parent directory, stopping at the root. ok reports
// whether the key is one of those.
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
	if len(e.Files) == 0 {
		return true, nil
	}
	if e.Selected < 0 && key != "End" && key != "G" {
		to = 0
	}
	return true, c.fileSelect(e, min(max(to, 0), len(e.Files)-1))
}

// fileSelect highlights entry i and rebuilds, so that every rendition
// shows the same entry.
func (c *Controller) fileSelect(e *Element, i int) error {
	st := c.St.File[e.ID]
	st.Index = i
	if st.Dirs == nil {
		st.Dirs = map[string]DirPick{}
	}
	st.Dirs[e.FileDir] = DirPick{Index: i, Top: e.Top}
	c.St.File[e.ID] = st
	c.Rebuild()
	return nil
}

// fileOpen opens the highlighted entry: a directory is entered (fileGo),
// a file is picked: its path goes to `selected`, and onPick runs (its
// context reads the path, an action carrying no payload).
func (c *Controller) fileOpen(e *Element) error {
	f := e.FileRef()
	if f.Name == "" {
		return nil
	}
	if f.Dir {
		return c.fileGo(e, joinFile(e.FileDir, f.Name))
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

// FileUp goes to the parent of the directory shown in HottyFilePicker id,
// stopping at the root.
func (c *Controller) FileUp(id string) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != FilePicker {
		return nil
	}
	return c.fileUp(e)
}

// fileUp goes to the parent of the directory shown, stopping at the root
// (nothing happens there). It restores the focus and scrolling in the parent,
// highlighting the directory just left.
func (c *Controller) fileUp(e *Element) error {
	if e.FileDir == "" {
		return nil
	}
	parent := cleanFile(path.Dir(e.FileDir))
	child := path.Base(e.FileDir)
	return c.fileGoTo(e, parent, child)
}

// fileGo shows the directory dir of the root: it writes the directory to
// the bound `path` (else to the renderer's state), rebuilds, and runs
// onChange (whose context reads `path`).
func (c *Controller) fileGo(e *Element, dir string) error {
	return c.fileGoTo(e, dir, "")
}

// fileGoTo shows directory dir of the root. When child is non-empty,
// we are returning from child, so focus is restored to child in dir,
// and scrolling is restored to show it.
func (c *Controller) fileGoTo(e *Element, dir, child string) error {
	dir = cleanFile(dir)
	n := c.V.Node(e.ID)
	st := c.St.File[e.ID]
	if st.Dirs == nil {
		st.Dirs = map[string]DirPick{}
	}
	st.Dirs[e.FileDir] = DirPick{Index: e.Selected, Top: e.Top}

	if saved, ok := st.Dirs[dir]; ok {
		st.Index = saved.Index
		c.St.Scroll[e.ID] = saved.Top
	} else {
		st.Index = -1
		c.St.Scroll[e.ID] = 0
	}

	if child != "" && child != "." {
		if fsys, ok := c.St.Roots[e.Root]; ok && fsys != nil {
			if ents, err := fs.ReadDir(fsys, fsDir(dir)); err == nil {
				var files []FileEntry
				for _, de := range ents {
					if !fileMatch(e.Query.Text, de.Name()) {
						continue
					}
					files = append(files, FileEntry{Name: de.Name(), Dir: de.IsDir()})
				}
				sortFile(files)
				for i, f := range files {
					if f.Dir && f.Name == child {
						st.Index = i
						break
					}
				}
			}
		}
	}

	delete(c.St.Revealed, e.ID)
	if bd, ok := n.Props["path"].(a2ui.Bound); ok && bd.Writable() {
		if err := c.S.Write(bd.Path, dir); err != nil {
			return err
		}
	} else {
		st.Path = dir
	}
	c.St.File[e.ID] = st
	c.Rebuild()
	if n.Props["onChange"] != nil {
		return c.invoke(n, "onChange", true)
	}
	return nil
}
