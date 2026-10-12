package cells

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyFilePicker in cells (profile §3.3, §3.4, §3.7), as bubbles'
// filepicker has one: the directory's path on a title row, then its
// entries, directories first, each a name (a directory with a trailing
// "/") and, for a file, its size. Every line starts two columns in; the
// highlighted entry's two columns are a "│" bar, as a HottyList's
// selected item has.

// fileIndent is the columns before every line of a HottyFilePicker.
const fileIndent = 2

// fileHead is the rows a HottyFilePicker's title takes: its path and a
// blank row.
const fileHead = 2

// fileBody is the rows a HottyFilePicker's entries take: its height's
// worth, else as many as it has; at least a row, for the empty text.
func fileBody(e *view.Element) int {
	n := len(e.Files)
	if e.Height > 0 {
		n = e.Height
	}
	return max(n, 1)
}

// fileTitle is the title a HottyFilePicker draws: " / " at the root,
// or " ◂ /path " in a subdirectory, whose chevron clicks up.
func fileTitle(e *view.Element) string {
	if e.FileDir == "" {
		return " / "
	}
	return " ◂ /" + e.FileDir + " "
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
// then K, M, G, T, to one decimal from K up. A directory's ("-").
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

// fileSizeWidth is the columns the size column takes: the widest size
// among the files, plus a space.
func fileSizeWidth(e *view.Element) int {
	n := 0
	for _, f := range e.Files {
		n = max(n, Width(fileSizeText(f)))
	}
	if n > 0 {
		n++
	}
	return n
}

// fileWidth is a HottyFilePicker's natural width: its widest line.
func fileWidth(e *view.Element) int {
	n := Width(fileTitle(e))
	for _, f := range e.Files {
		n = max(n, Width(fileLabel(f)))
	}
	n += fileSizeWidth(e)
	if e.Error != "" {
		n = max(n, Width(e.Error))
	}
	return fileIndent + n
}

// fileMinimum is a HottyFilePicker's minimum: its indent and a few
// columns, so that its names cut.
func fileMinimum(e *view.Element) int { return fileIndent + 4 }

// fileHeight is a HottyFilePicker's height.
func fileHeight(e *view.Element) int { return fileHead + fileBody(e) }

// paintFilePicker paints a HottyFilePicker at (x, y), w wide. Its title
// row is the directory shown, reversed in the accent, as a HottyList's
// title; the highlighted entry's bar, name and size are in the accent
// while the picker has the keyboard, else its name is bold and its bar
// muted. A file's size is right-aligned in muted; the empty text, or the
// error, shows where its entries would.
func (l *layout) paintFilePicker(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	focused := r.focused(e.ID)
	in := x + fileIndent
	iw := max(w-fileIndent, 0)
	row := y
	cv.write(in, row, iw, fit(line(fileTitle(e), style{role: Accent, attr: Reverse}), iw))
	r.hits = append(r.hits, hit{x: x, y: row, w: w, h: 1, id: e.ID, opt: -1})
	row += fileHead

	switch {
	case e.Error != "":
		cv.write(in, row, iw, fit(line("✗ "+e.Error, style{role: Error}), iw))
	case len(e.Files) == 0:
		cv.write(in, row, iw, fit(line(e.Placeholder, style{role: Muted}), iw))
	}
	if e.Error != "" || len(e.Files) == 0 {
		r.hits = append(r.hits, hit{x: x, y: row, w: w, h: 1, id: e.ID, opt: -2})
		r.boxes[e.ID] = box{x, y, w, row + 1 - y}
		return
	}
	sizeW := fileSizeWidth(e)
	end := len(e.Files)
	if e.Height > 0 {
		end = min(e.Top+e.Height, end)
	}
	at := row
	for i := e.Top; i < end; i++ {
		f := e.Files[i]
		name, size, bar := style{}, style{role: Muted}, style{}
		if i == e.Selected {
			name, bar = style{attr: Bold}, style{role: Muted}
			if focused {
				name, size, bar = style{role: Accent}, style{role: Accent}, style{role: Accent}
			}
		}
		if i == e.Selected {
			cv.write(x, at, w, line("│", bar))
		}
		nw := max(iw-sizeW, 0)
		cv.write(in, at, nw, fit(line(fileLabel(f), name), nw))
		if s := fileSizeText(f); s != "" && sizeW > 0 {
			cv.write(in+max(iw-sizeW, 0), at, sizeW, fit(line(s, size), sizeW))
		}
		r.hits = append(r.hits, hit{x: x, y: at, w: w, h: 1, id: e.ID, opt: i})
		at++
	}
	r.boxes[e.ID] = box{x, y, w, at - y}
}

// clickFile is a click on a HottyFilePicker, which has the keyboard now:
// on an entry, it selects it, or opens or picks it when it was selected
// already; on the title in a subdirectory, it goes up one level; elsewhere,
// it does nothing more.
func (r *Rendition) clickFile(e *view.Element, item int) error {
	if item == -1 && e.FileDir != "" {
		return r.c.FileUp(e.ID)
	}
	if item < 0 {
		return nil
	}
	return r.c.ClickFile(e.ID, item)
}
