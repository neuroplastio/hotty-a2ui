package cells

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyFilePicker in cells (profile §3.3, §3.4, §3.7), as bubbles'
// filepicker has one: the directory's path on a title row, then its
// entries, directories first, each a name (a directory with a trailing
// "/") and, for a file, its size. Every line starts two columns in; the
// highlighted entry's two columns are a "│" bar, as a HottyList's
// selected item has.

const (
	// fileIndent is the columns before every line of a HottyFilePicker.
	fileIndent = 2
	// fileHead is the rows a HottyFilePicker's title takes: its path and
	// a blank row.
	fileHead = 2
)

// The hits of a HottyFilePicker that are not an entry's (whose opt is
// the entry's index): its title, which a click in a subdirectory takes up
// to the parent, and the rest of its box.
const (
	fileTitleHit = -1
	fileElseHit  = -2
)

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

// fileSizeWidth is the columns the size column takes: the widest size
// among the files, plus a space.
func fileSizeWidth(e *view.Element) int {
	n := 0
	for _, f := range e.Files {
		n = max(n, Width(f.SizeText()))
	}
	if n > 0 {
		n++
	}
	return n
}

// fileBar is the columns a HottyFilePicker's scrollbar takes: while its
// entries scroll, a blank column and the bar (scrollBar); none otherwise.
func fileBar(e *view.Element) int {
	if e.Height > 0 && len(e.Files) > e.Height {
		return scrollBar
	}
	return 0
}

// fileWidth is a HottyFilePicker's natural width: its widest line, and
// its scrollbar.
func fileWidth(e *view.Element) int {
	n := Width(fileTitle(e))
	for _, f := range e.Files {
		n = max(n, Width(f.Label()))
	}
	n += fileSizeWidth(e)
	if e.Error != "" {
		n = max(n, Width("✗ "+e.Error))
	}
	return fileIndent + n + fileBar(e)
}

// fileMinimum is a HottyFilePicker's minimum: its indent and a few
// columns, so that its names cut, and its scrollbar.
func fileMinimum(e *view.Element) int { return fileIndent + 4 + fileBar(e) }

// fileHeight is a HottyFilePicker's height.
func fileHeight(e *view.Element) int { return fileHead + fileBody(e) }

// paintFilePicker paints a HottyFilePicker at (x, y), w wide. Its title
// row is the directory shown, reversed in the accent, as a HottyList's
// title; the highlighted entry's bar, name and size are in the accent
// while the picker has the keyboard, else its name is bold and its bar
// muted. A file's size is right-aligned in muted; the empty text, or the
// error, shows where its entries would. While its entries scroll, a scroll
// view's bar runs beside them.
func (l *layout) paintFilePicker(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	h := fileHeight(e)
	r.boxes[e.ID] = box{x, y, w, h}
	// The box's hit first: the title's and the entries' are over it.
	r.hits = append(r.hits, hit{x: x, y: y, w: w, h: h, id: e.ID, opt: fileElseHit},
		hit{x: x, y: y, w: w, h: 1, id: e.ID, opt: fileTitleHit})
	bar := min(fileBar(e), w)
	in := x + fileIndent
	iw := max(w-fileIndent-bar, 0)
	cv.write(in, y, iw, fit(line(fileTitle(e), style{role: Accent, attr: Reverse}), iw))
	row := y + fileHead
	switch {
	case e.Error != "":
		cv.write(in, row, iw, fit(line("✗ "+e.Error, style{role: Error}), iw))
		return
	case len(e.Files) == 0:
		cv.write(in, row, iw, fit(line(e.Placeholder, style{role: Muted}), iw))
		return
	}
	if bar > 0 {
		l.paintScrollbar(cv, e, x+w-1, row, e.Height, len(e.Files), e.Top)
	}
	sizeW := fileSizeWidth(e)
	nw := max(iw-sizeW, 0)
	end := len(e.Files)
	if e.Height > 0 {
		end = min(e.Top+e.Height, end)
	}
	focused := r.focused(e.ID)
	for i := e.Top; i < end; i++ {
		f := e.Files[i]
		name, size := style{}, style{role: Muted}
		if i == e.Selected {
			barSt := style{role: Muted}
			name = style{attr: Bold}
			if focused {
				name, size, barSt = style{role: Accent}, style{role: Accent}, style{role: Accent}
			}
			cv.write(x, row, w, line("│", barSt))
		}
		cv.write(in, row, nw, fit(line(f.Label(), name), nw))
		if s := f.SizeText(); s != "" {
			cv.write(in+nw, row, sizeW, fit(line(s, size), sizeW))
		}
		r.hits = append(r.hits, hit{x: x, y: row, w: w, h: 1, id: e.ID, opt: i})
		row++
	}
}

// clickFile is a click on a HottyFilePicker, which has the keyboard now:
// on an entry, it highlights it, or opens or picks it when it was
// highlighted already; on the title in a subdirectory, it goes up one
// level; elsewhere, it does nothing more.
func (r *Rendition) clickFile(e *view.Element, item int) error {
	switch {
	case item >= 0:
		return r.c.ClickItem(e.ID, item, r.count)
	case item == fileTitleHit:
		return r.c.FileUp(e.ID)
	}
	return nil
}
