package main

import tea "charm.land/bubbletea/v2"

// The terminal's mouse beyond its reports. Shift with a click comes to the
// program (XTSHIFTESCAPE), where a terminal keeps it for its own selection
// otherwise, so that a field extends its selection with it. The pointer
// has the shape the cells under it ask for (OSC 22): an I-beam over a
// field's value, as a browser has it.
const (
	shiftCaptureOn  = "\x1b[>1s"
	shiftCaptureOff = "\x1b[>0s"
)

// pointer is the pointer's shape the program last set over the cells, ""
// for the terminal's own.
type pointer struct{ shape string }

// set has the terminal show shape over the cells, "" its own, when that is
// not what it shows.
func (p *pointer) set(shape string) tea.Cmd {
	if shape == p.shape {
		return nil
	}
	p.shape = shape
	return tea.Raw(osc22(shape))
}

// off gives the terminal its pointer and its Shift with a click back.
func (p *pointer) off() tea.Cmd {
	seq := shiftCaptureOff
	if p.shape != "" {
		seq += osc22("")
		p.shape = ""
	}
	return tea.Raw(seq)
}

func osc22(shape string) string {
	if shape == "" {
		shape = "default"
	}
	return "\x1b]22;" + shape + "\x1b\\"
}
