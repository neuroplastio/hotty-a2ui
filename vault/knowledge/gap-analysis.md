# Gap analysis: the kit against Bubble Tea and OpenTUI

As of 2026-10-09: hotty-a2ui 18b5b87, A2UI at `third_party/a2ui/REV` (ae466ff).

## What was compared

- **Bubble Tea's ecosystem.**
  - bubbles: spinner, textinput, textarea, table, progress, paginator,
    viewport, list, filepicker, timer, stopwatch, help, key, cursor, tree.
  - huh's fields: confirm, filepicker, input, multiselect, note, select, text.
  - lipgloss: list, table, tree.
- **OpenTUI.**
  - Components: Text, Box, Input, Textarea, Select, TabSelect, Slider,
    ScrollBox, ScrollBar, Code, Markdown, LineNumbers, Diff, TextTable,
    ASCIIFont, FrameBuffer, Image, QR code, EmbeddedTerminal.
  - Modules: keymap, animation, console, selection, clipboard, audio,
    notifications, and three.js.
- **The kit.**
  - A2UI's basic catalog: Text, Image, Icon, Video, AudioPlayer, Row, Column,
    List, Card, Tabs, Modal, Divider, Button, TextField, CheckBox,
    ChoicePicker, Slider, DateTimeInput.
  - The hotty catalog: Shortcut, Form, the focus and blur functions, and the
    autofocus and keys extensions.

## What the kit already covers

| Theirs | The kit's |
| --- | --- |
| textinput, huh input, OpenTUI Input | TextField |
| textarea, huh text, OpenTUI Textarea | TextField `longText` |
| huh select and multiselect, OpenTUI Select (plain) | ChoicePicker: a select with its popover list, checkboxes, chips |
| OpenTUI TabSelect | Tabs |
| OpenTUI Slider | Slider |
| OpenTUI Box, lipgloss layout | Card, Row, Column |
| OpenTUI Markdown, huh note | Text (Markdown) |
| lipgloss list | List, and Markdown lists in Text |
| OpenTUI Image | Image (a picture on a host, `[image: …]` in cells) |
| bubbles key, OpenTUI keymap | Shortcut, the keys extension, SPEC §10.2 keymaps |
| bubbles cursor | the text field's caret |
| bubbles viewport, OpenTUI ScrollBox (in part) | List scrolls; the rest is KIT-07 |

## The gaps

Each gap has a ticket. Its two legs are on the board:
- `c`, the cells rendition (phase 1);
- `h`, the HOTTY layer (phase 2).

The number gives the gap's priority. The board gives the order in which
they're built.

| Ticket | Gap | Reference | Priority | Cells (phase 1) | HOTTY (phase 2) |
| --- | --- | --- | --- | --- | --- |
| KIT-01 | **Table** | bubbles table, lipgloss table, OpenTUI TextTable | High | Header row, column widths from the content and the hints, alignment, a highlighted row moved by ↑ ↓ PgUp PgDn Home End, a sticky header while the body scrolls, cut text ending in `…` | A real table in proportional type, a sticky header, the row under the mouse, sorting by a click on a header |
| KIT-02 | **Progress** | bubbles progress | High | A bar filled in eighths (`▏`…`█`), a percentage, lipgloss's gradient, and an indeterminate segment that moves | A smooth bar with its label, animated where the host animates |
| KIT-03 | **Spinner** | bubbles spinner, OpenTUI animation | High | bubbles' frame sets (dots, line, minidot, pulse, points, meter…) ticked by the renderer, and a label. It also draws A2UI's pending state for async function values | The same frames, or a CSS animation |
| KIT-04 | **Rich list**: items with descriptions, a filter, a status line | bubbles list, huh select and multiselect, OpenTUI Select | High | Title and description rows, a highlight bar, a filter as you type (`/`), pagination dots, a status line, an empty state | Two-line rows in proportional type, the row under the mouse |
| KIT-05 | **Code** | OpenTUI Code and LineNumbers | High for coding agents | Syntax highlighting, a line-number gutter, marked lines, wrapping or horizontal scrolling. Text's fenced code blocks use it too | `pre` with highlight spans. Selectable text and a copy button wait for the host's clipboard (board, *The host*) |
| KIT-06 | **Diff** | OpenTUI Diff | High for coding agents | Unified and split views from a unified diff or from two texts: `+`/`-` gutters, line numbers, changed words marked, hunk headers, unchanged runs folded | A split view in real columns, word-level marks, acting on a hunk (select-then-act, see a2ui-limits L2) |
| KIT-07 | **Scroll view** | bubbles viewport, OpenTUI ScrollBox and ScrollBar | Medium–High | A box of fixed height with a scrollbar column. It follows the tail while you're at the bottom, and scrolls with the wheel, PgUp and PgDn | The host's own scrolling (SPEC §8 area and scroll) with a native scrollbar |
| KIT-08 | **Key hints** | bubbles help and key | Medium, cheap | A line built from the surface's Shortcuts and the focused field's keymap, short and full (`?`) | Keys drawn as keycaps |
| KIT-09 | **Tree** | bubbles tree, lipgloss tree | Medium | `├──` `└──` guides, folding with ← →, a selected node | Disclosure triangles and indentation guides |
| KIT-10 | **Chart**: sparkline, line, bar (`HottyChart`, `HottySparkline`) | ntcharts, Bubble Tea's chart library (journal 2026-10-10.6); OpenTUI FrameBuffer. hotty-go has chart, braille and series | Medium | Sparklines (`▁▂▃▄▅▆▇█`), line charts in braille, bars in eighths, axes and labels | SVG, checked with `hotty render` at 1.6x |
| KIT-11 | **Suggestions** (autocomplete) | bubbles textinput's suggestions | Medium | Ghost text after the caret that Tab or → accepts, and a list under the field | The list as a surface at a higher z, as the select's is |
| KIT-12 | **Selection and copy** | OpenTUI selection and clipboard | Medium | Dragging the mouse selects text in cells. Copy goes through OSC 52 | The host's own selection and copy. This waits for the host's clipboard (board, *The host*) |
| KIT-13 | **Toast and tooltip** | OpenTUI notifications, console | Low–Medium | A toast is a box in a corner, over the surface, that times out. A tooltip is a component's `accessibility.description`, shown in a status line | A toast is a surface at a higher Z. A tooltip is a surface at a higher z, at the hovered area |
| KIT-14 | **Timer and stopwatch** | bubbles timer and stopwatch | Low | A countdown or elapsed time that ticks, and an action when time is up | The same, set in display type |
| KIT-15 | **Paginator** | bubbles paginator | Low | Dots (`• ○`) or `3/10`, moved with ← → | The same |
| KIT-16 | **Confirm** | huh confirm | Low | Yes and No inline, with the y and n keys | The same. It may turn out to be a pattern made of Buttons, not a component |
| KIT-17 | **File picker** | bubbles filepicker, huh filepicker | Low | A directory listing walked with ↑ ↓ ← →. It shows only the roots the program grants (a2ui-limits L10) | The same, with icons |
| KIT-18 | **Terminal** | OpenTUI EmbeddedTerminal | Low, large | A VT pane that runs only commands the program has registered, never a command line from the agent (L10) | To be decided |
| KIT-19 | **Big text** | OpenTUI ASCIIFont | Low | figlet-style block letters | Large display type |
| KIT-20 | **QR code** | OpenTUI QR | Low | Half blocks (`▀▄█`) | SVG, checked with `hotty render` at 1.6x |
| KIT-21 | **3D and audio** | OpenTUI three.js and audio | Parked | Not for a terminal UI kit. AudioPlayer is already a labelled link. Media is gov R-2's research | — |
| KIT-22 | **Toggle** (`HottySwitch`) | iOS's switch, Material's Switch; the maintainer asked for it (2026-10-10) | Medium | A boolean as CheckBox's (label, bound value, checks), drawn as a track with a knob at its end: `accent` when on, `muted` when off. Space, Enter or a click flips it | A pill with a knob, `role=switch` and `aria-checked`. The knob slides by deltas, since Blitz runs no CSS transitions (host-motion) |
| KIT-23 | **Drag and drop**: reorder and move as first-class behaviour of a list, a table, a tree and a templated List | iOS and SortableJS (live), Finder and VS Code (a line); the maintainer asked for it (2026-10-10) | High | The terminal's mouse reports with the button held, plus keys: Alt+arrows, Ctrl+X then Ctrl+V | SPEC §9.1 drags, with gaps G1 (where in the target) and G2 (a drop-only target); the design is [drag-and-drop](drag-and-drop.md) |

## What "parity" means here

A phase-1 leg is judged against its reference, side by side, at the same size
(KIT-REF). It should look and behave like the reference:
- the same density;
- the same glyphs;
- the same keys;
- the same states (focused, disabled, empty, loading, error).

It doesn't have to copy the reference pixel for pixel. Where Bubble Tea and
OpenTUI differ, Bubble Tea's look leads, because the kit's programs are
Bubble Tea programs (the storybook is one), so its components should look at
home next to bubbles. OpenTUI is the reference only where Bubble Tea has no
equivalent: Code, Diff, QR and big text.
