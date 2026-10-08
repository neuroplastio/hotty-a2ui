package basic

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// digits reads a count of fraction digits: nil when absent or not a
// number, else clamped to 0–20.
func digits(v any) (int, bool) {
	if v == nil {
		return 0, false
	}
	f := a2ui.ToNumber(v)
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return min(max(int(f), 0), 20), true
}

func groupingOff(v any) bool { return v != nil && !a2ui.Truthy(v) }

// FormatNumber formats a number for en-US: decimals fraction digits (up to
// three, trailing zeros dropped, when not given), grouped by thousands
// unless grouping is falsy. Not a number is "".
func FormatNumber(value, decimals, grouping any) string {
	n := a2ui.ToNumber(value)
	if math.IsNaN(n) {
		return ""
	}
	d, fixed := digits(decimals)
	if !fixed {
		d = 3
	}
	return decimal(n, d, fixed, !groupingOff(grouping))
}

// currencySymbols are en-US's symbols for the currencies it writes with
// one, as Intl does.
var currencySymbols = map[string]string{
	"USD": "$", "EUR": "€", "GBP": "£", "JPY": "¥", "INR": "₹", "KRW": "₩",
	"CNY": "CN¥", "CAD": "CA$", "AUD": "A$", "MXN": "MX$", "BRL": "R$",
	"ILS": "₪", "VND": "₫", "NZD": "NZ$", "HKD": "HK$", "TWD": "NT$", "PHP": "₱",
}

// FormatCurrency formats an amount for en-US: the currency's symbol, or
// its code and a space, then the amount with decimals digits (two when not
// given). Not a number is "".
func FormatCurrency(value, currency, decimals, grouping any) string {
	n := a2ui.ToNumber(value)
	if math.IsNaN(n) {
		return ""
	}
	d, ok := digits(decimals)
	if !ok {
		d = 2
	}
	code := strings.ToUpper(a2ui.ToString(currency))
	sym, ok := currencySymbols[code]
	if !ok {
		sym = code + " "
	}
	s := decimal(math.Abs(n), d, true, !groupingOff(grouping))
	if n < 0 && strings.Trim(s, "0.,") != "" {
		return "-" + sym + s
	}
	return sym + s
}

// decimal writes n with d fraction digits, rounding half away from zero on
// its exact value, as Intl does; without fixed, trailing zeros go.
func decimal(n float64, d int, fixed, grouping bool) string {
	neg := n < 0
	exact := strconv.FormatFloat(math.Abs(n), 'f', 1100, 64)
	intPart, frac, _ := strings.Cut(exact, ".")
	frac += strings.Repeat("0", d+1)
	digits := []byte(intPart + frac[:d])
	if frac[d] >= '5' {
		i := len(digits) - 1
		for ; i >= 0 && digits[i] == '9'; i-- {
			digits[i] = '0'
		}
		if i < 0 {
			digits = append([]byte{'1'}, digits...)
		} else {
			digits[i]++
		}
	}
	ip := string(digits[:len(digits)-d])
	fp := string(digits[len(digits)-d:])
	if !fixed {
		fp = strings.TrimRight(fp, "0")
	}
	if grouping {
		ip = group(ip)
	}
	s := ip
	if fp != "" {
		s += "." + fp
	}
	if neg && strings.Trim(s, "0.,") != "" {
		s = "-" + s
	}
	return s
}

func group(ip string) string {
	if len(ip) <= 3 {
		return ip
	}
	var b strings.Builder
	head := len(ip) % 3
	if head > 0 {
		b.WriteString(ip[:head])
	}
	for i := head; i < len(ip); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(ip[i : i+3])
	}
	return b.String()
}

var (
	dateTokens = regexp.MustCompile(`yyyy|yy|MMMM|MMM|MM|M|EEEE|E|dd|d|HH|H|hh|h|mm|ss|a`)
	// timeOffset is the offset ending a timestamp's time part: Z, ±hh,
	// ±hhmm or ±hh:mm, after the T or space that starts the time.
	timeOffset = regexp.MustCompile(`[T ]\d{2}(?::?\d{2}(?::?\d{2}(?:[.,]\d+)?)?)?(?:([zZ])|([+-])(\d{2})(?::?(\d{2}))?)$`)
	// written are the fields as written: year, month, day, and maybe hour,
	// minute and second.
	written = regexp.MustCompile(`^([+-]?\d{4,6})-?(\d{2})-?(\d{2})(?:[T ](\d{2})(?::?(\d{2})(?::?(\d{2}))?)?)?`)
)

var timeLayouts = []string{
	"2006-01-02T15:04:05.999999999Z07:00",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04Z07:00",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05.999999999Z07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04:05.999999999Z0700",
	"2006-01-02T15:04:05.999999999-07",
	"2006-01-02",
	"20060102",
	"20060102T150405",
}

// parseTimestamp reads an ISO 8601 timestamp: the instant, and the
// wall-clock time it was written in. Without an offset it is UTC, never
// the host's zone. Fields that do not survive the round trip (2026-02-30)
// make it no timestamp.
func parseTimestamp(s string) (instant, wall time.Time, ok bool) {
	var t time.Time
	var err error
	for _, l := range timeLayouts {
		if t, err = time.Parse(l, s); err == nil {
			break
		}
	}
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	instant = t.UTC()
	wall = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	if m := timeOffset.FindStringSubmatch(s); m == nil || (m[1] == "" && m[2] == "") {
		instant = wall
	}
	w := written.FindStringSubmatch(s)
	if w == nil {
		return time.Time{}, time.Time{}, false
	}
	field := func(i int, have int) bool {
		if w[i] == "" {
			return true
		}
		n, _ := strconv.Atoi(w[i])
		return n == have
	}
	if !field(1, wall.Year()) || !field(2, int(wall.Month())) || !field(3, wall.Day()) ||
		!field(4, wall.Hour()) || !field(5, wall.Minute()) || !field(6, wall.Second()) {
		return time.Time{}, time.Time{}, false
	}
	return instant, wall, true
}

// FormatDate formats an ISO 8601 timestamp with a TR35 pattern (the tokens
// yyyy yy MMMM MMM MM M EEEE E dd d HH H hh h mm ss a; text in single
// quotes is copied as it is, ” is a quote, and other text is copied), in
// the wall-clock time it was written in. The pattern "ISO" is
// the UTC instant as JavaScript's toISOString writes it. A value that is
// no timestamp is "".
func FormatDate(value, pattern any) string {
	if !a2ui.Truthy(value) {
		return ""
	}
	instant, t, ok := parseTimestamp(a2ui.ToString(value))
	if !ok {
		return ""
	}
	format := ""
	if a2ui.Truthy(pattern) {
		format = a2ui.ToString(pattern)
	}
	if format == "ISO" {
		return instant.Format("2006-01-02T15:04:05.000Z")
	}
	if format == "" {
		format = "yyyy-MM-dd"
	}
	h12 := t.Hour() % 12
	if h12 == 0 {
		h12 = 12
	}
	pad := func(n int) string {
		if n < 10 {
			return "0" + strconv.Itoa(n)
		}
		return strconv.Itoa(n)
	}
	return unquoted(format, func(part string) string {
		return dateTokens.ReplaceAllStringFunc(part, func(tok string) string {
			switch tok {
			case "yyyy":
				return strconv.Itoa(t.Year())
			case "yy":
				return pad(t.Year() % 100)
			case "MMMM":
				return t.Month().String()
			case "MMM":
				return t.Month().String()[:3]
			case "MM":
				return pad(int(t.Month()))
			case "M":
				return strconv.Itoa(int(t.Month()))
			case "EEEE":
				return t.Weekday().String()
			case "E":
				return t.Weekday().String()[:3]
			case "dd":
				return pad(t.Day())
			case "d":
				return strconv.Itoa(t.Day())
			case "HH":
				return pad(t.Hour())
			case "H":
				return strconv.Itoa(t.Hour())
			case "hh":
				return pad(h12)
			case "h":
				return strconv.Itoa(h12)
			case "mm":
				return pad(t.Minute())
			case "ss":
				return pad(t.Second())
			case "a":
				if t.Hour() < 12 {
					return "AM"
				}
				return "PM"
			}
			return tok
		})
	})
}

// unquoted is a TR35 pattern with f applied to its text outside quotes:
// quoted text is copied as it is, without its quotes, and ” is a quote,
// in quotes or out. A quote left open runs to the end.
func unquoted(pattern string, f func(string) string) string {
	var b, run strings.Builder
	quoted := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		if c != '\'' {
			if quoted {
				b.WriteByte(c)
			} else {
				run.WriteByte(c)
			}
			continue
		}
		if i+1 < len(pattern) && pattern[i+1] == '\'' {
			if quoted {
				b.WriteByte('\'')
			} else {
				run.WriteByte('\'')
			}
			i++
			continue
		}
		if !quoted {
			b.WriteString(f(run.String()))
			run.Reset()
		}
		quoted = !quoted
	}
	b.WriteString(f(run.String()))
	return b.String()
}

// Pluralize picks the form for a count: an explicit zero, one or two for
// exactly 0, 1 or 2, else en-US's plural category (one for 1, other
// otherwise), falling back to other. An empty form is a form.
func Pluralize(value any, forms map[string]any) string {
	n := a2ui.ToNumber(value)
	has := func(k string) bool { _, ok := forms[k]; return ok }
	cat := "other"
	switch {
	case n == 0 && has("zero"):
		cat = "zero"
	case n == 1 && has("one"):
		cat = "one"
	case n == 2 && has("two"):
		cat = "two"
	case n == 1:
		cat = "one"
	}
	if f, ok := forms[cat]; ok {
		return a2ui.ToString(f)
	}
	return a2ui.ToString(forms["other"])
}
