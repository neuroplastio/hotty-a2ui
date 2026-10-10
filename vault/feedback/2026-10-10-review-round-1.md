# Review, round 1 (2026-10-10)

The maintainer's notes from the blind review in `cmd/review`
(`.shots/blind/feedback.md`), unblinded with `.shots/blind/.key`. In the
reference shots the kit is the right side; in the look-* shots, "before" is
the kit before KIT-LOOK and "after" is the kit with it. Their words are
kept as written. The shots are cells, in VHS, in the Terminal theme.

## code — kit preferred

- Kit: "User-agent line is clipped (but this is good, I expect horizontal
  scroll here)"
- Bubble Tea: "Whitespace is wrong, line break in User-Agent leads to two
  lines, gutter markers and line numbers are missing"

## diff — kit preferred

- Kit: "Functionally, this looks better. Supports in-line diffs. Has
  highlighting. I would change the colors a bit though - need colored tint
  on modified lines (keeping syntax highlighting) and less toxic background
  on the in-line diff"
- OpenTUI: "I bet this is not our a2ui, and it looks too basic."

## form

- huh: "The form is functionally different here with yes/no buttons. Not
  a2ui"
- Kit: "I would like have either: A: labels on the same row as the input,
  including plan select; B: left padding for inputs (one cell). Select
  input looks inconsistent comparing to text inputs"

## list — kit preferred

- bubbles: "Too much space above the pager"
- Kit: "This one is better because of the pager spacing. Other than this
  and the color - pixel perfect"

## look-invitation

- After KIT-LOOK: "This doesn't read like a form to me. Location picker is
  ugly in both options though, I would use either select or a multi-line
  option-style list of item"
- Before: "The undeline for text inputs looks nice, we should adopt it"

## look-task — before preferred

- After KIT-LOOK: "B wins"
- Before: "Focus indicator is better here: it clearly says which input is
  selected. Also, underline and vertical alignment looks good. I would add
  gray tone to the second line with a padding though (input subtext)."

## look-validator

- After KIT-LOOK: "I like the gutter indicator here. I think if we combine
  gutter + underline + blinking line cursor that would be the best combo.
  Underline also allows to save vertical space"
- Before: "This looks more like a form"

## progress

- bubbles: "The gradient here looks really nice, eye candy"
- Kit: "No gradient and no gap after "Progress" title"

## suggest

- Kit: "This is nice, but I would need to make sure that this dropdown
  doesn't affect the layout and rendered as an overlay (may need borders).
  I think we should support both styles"

## switch — kit preferred

- Kit: "So much more compact and feels better"

## table — kit preferred

- Kit: "Cleaner, nice use of the heading line with the counter. One thing to
  add: an indicator (small arrow) that there are more items in either
  direction. Or a scrollbar (similar to viewport)"
- bubbles: "Bad: unclear how many items"

## textarea — kit preferred

- Kit: "Looks fine"
- bubbles: "This doesn't read like an input at all. It has line numbers and
  line 4 is fully highlighted"

## toast — kit preferred

- Kit: "Win. More compact, uses icons, has actionable item inside"

## tree — kit preferred

- Kit: "Win. Prints full path, has item counts"
- Bubble Tea: "The arrow on the left is annoying"

## viewport — kit preferred

- bubbles: "No scrollbars, impossible to read"
- Kit: "Looks good, but some background separation could improve it"

## spinner

No notes.
