package basic

import "testing"

func TestFormatNumber(t *testing.T) {
	for _, c := range []struct {
		v, d, g any
		want    string
	}{
		{1234.567, nil, nil, "1,234.567"},
		{1234.5678, nil, nil, "1,234.568"},
		{2.5, 0.0, nil, "3"},
		{-2.5, 0.0, nil, "-3"},
		{0.125, 2.0, nil, "0.13"},
		{1.005, 2.0, nil, "1.00"},
		{999.999, 2.0, nil, "1,000.00"},
		{-0.001, 2.0, nil, "0.00"},
		{1e6, nil, false, "1000000"},
		{"x", nil, nil, ""},
	} {
		if got := FormatNumber(c.v, c.d, c.g); got != c.want {
			t.Errorf("FormatNumber(%v, %v, %v) = %q, want %q", c.v, c.d, c.g, got, c.want)
		}
	}
}

func TestFormatCurrency(t *testing.T) {
	for _, c := range []struct {
		v    any
		cur  string
		want string
	}{
		{1234.5, "USD", "$1,234.50"},
		{-3.0, "eur", "-€3.00"},
		{10.0, "CHF", "CHF 10.00"},
	} {
		if got := FormatCurrency(c.v, c.cur, nil, nil); got != c.want {
			t.Errorf("FormatCurrency(%v, %s) = %q, want %q", c.v, c.cur, got, c.want)
		}
	}
}

func TestFormatDate(t *testing.T) {
	for _, c := range []struct{ v, f, want string }{
		{"2026-09-04T23:30:00-05:00", "yyyy-MM-dd HH:mm", "2026-09-04 23:30"},
		{"2026-02-02T15:17:00Z", "E MMM d, yyyy h:mm a", "Mon Feb 2, 2026 3:17 PM"},
		{"2026-09-04", "EEEE, MMMM d", "Friday, September 4"},
		{"2026-09-04T23:30:00-05:00", "ISO", "2026-09-05T04:30:00.000Z"},
		{"2026-02-30", "yyyy", ""},
		{"nope", "yyyy", ""},
	} {
		if got := FormatDate(c.v, c.f); got != c.want {
			t.Errorf("FormatDate(%q, %q) = %q, want %q", c.v, c.f, got, c.want)
		}
	}
}
