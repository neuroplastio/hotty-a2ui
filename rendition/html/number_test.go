package html

import (
	"strings"
	"testing"
)

// TestGroupBreaks: a Text that is one number breaks after its group
// separators only; prose and a decimal point keep their numbers whole.
func TestGroupBreaks(t *testing.T) {
	for _, c := range []struct{ plain, html, want string }{
		{"$850,000,000,000.00", "<p>$850,000,000,000.00</p>", "<p>$850,|000,|000,|000.00</p>"},
		{"$43,500.25", "<p>$43,500.25</p>", "<p>$43,|500.25</p>"},
		{"1.000.000,50 €", "<p>1.000.000,50 €</p>", "<p>1.000.000,50 €</p>"}, // a space: not one number
		{"1.000.000,50€", "<p>1.000.000,50€</p>", "<p>1.|000.|000,50€</p>"},
		{"-0.5%", "<p>-0.5%</p>", "<p>-0.5%</p>"},
		{"1,000 people", "<p>1,000 people</p>", "<p>1,000 people</p>"},
		{"1,0000", "<p>1,0000</p>", "<p>1,0000</p>"},
		{"$1,234", `<p><strong>$1,234</strong></p>`, `<p><strong>$1,|234</strong></p>`},
	} {
		got := strings.ReplaceAll(groupBreaks(c.plain, c.html), "\u200b", "|")
		if got != c.want {
			t.Errorf("%q: %q, want %q", c.plain, got, c.want)
		}
	}
}
