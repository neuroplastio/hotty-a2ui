# Board

Last updated: 2026-10-09

**Active phase:** 1 — Cells parity ([roadmap](../roadmap.md))
**Waiting on the maintainer:** a look at KIT-LOOK (journal 2026-10-09.4):
fields drawn as huh draws them; at KIT-02c and KIT-03c (journal
2026-10-09.5): progress bars and spinners; and at KIT-01c (journal
2026-10-09.7): the table. Phase 1 is a go (2026-10-09). New names
follow [catalog-naming](../knowledge/catalog-naming.md).

## Legend

- `KIT-NN` is a gap ([gap-analysis](../knowledge/gap-analysis.md)). The
  number is its priority.
- `c` is its cells leg (phase 1), and `h` its HOTTY leg (phase 2).
- The sections below list the legs in the order they're built. The first
  open leg that nothing blocks is next.

## Phase 1 — Cells parity

- [ ] **KIT-04c** — Rich list: title and description rows, the `/` filter,
  pagination dots, a status line, an empty state.
- [ ] **KIT-08c** — Key hints from the Shortcuts and the focused field's
  keymap, short and full (`?`).
- [ ] **KIT-07c** — Scroll view: a scrollbar column, following the tail, the
  wheel, PgUp and PgDn, and a `scrollTo` renderer function (L8). Lines
  arrive as chunks (L4).
- [ ] **KIT-05c** — Code: syntax highlighting, a line-number gutter, marked
  lines, wrapping. Text's fenced code blocks use it too.
- [ ] **KIT-06c** — Diff: unified and split views, gutters, word marks, hunk
  headers, folded runs. Select-then-act on a hunk (L2).
- [ ] **KIT-09c** — Tree: guides, folding, a selected node.
- [ ] **KIT-10c** — Chart: sparkline, braille line chart and bars, using
  hotty-go's chart, braille and series.
- [ ] **KIT-11c** — Suggestions: ghost text and a list under the field.
- [ ] **KIT-12c** — Selection and copy in cells (OSC 52).
- [ ] **KIT-13c** — Toast and tooltip (tooltip from
  `accessibility.description`).
- [ ] **KIT-14c** — Timer and stopwatch, with `onTimeout`.
- [ ] **KIT-15c** — Paginator.
- [ ] **KIT-16c** — Confirm, or a story showing it as a pattern of Buttons.
- [ ] **KIT-19c** — Big text.
- [ ] **KIT-20c** — QR code.
- [ ] **KIT-17c** — File picker, with the profile's rule for granted roots
  (L10).
- [ ] **KIT-18c** — Terminal, with the profile's rule for registered
  commands (L10).

## Phase 2 — The HOTTY layer

### The host

As of 2026-10-09, from the hotty agent:
- **Higher z.** Placement `z` is in the spec (§5.2). Surfaces with equal z
  go by creation order. Toasts, tooltips and suggestion lists use it.
- **No popover.** A dedicated popover is parked by the maintainer, so
  don't design for one. A separate surface at a higher z is the tool,
  as the select's list already is.
- **Scrolling is live.** §5.3 and §8 work in hotty-blitz, hottyterm and the
  addon: `scroll` on a=doc, overlay scrollbars, and `area`.
- **No clipboard yet.** This is the parked clipboard problem: hotty-blitz
  has no clipboard, and a paste goes to the program, not to the focused
  field. KIT-12h and KIT-05h's copy button wait for the maintainer.
- **SVG.** Inline SVG and SVG in `<img>` render in hotty-blitz, with two
  gaps:
  - CSS paint inside inline SVG is incomplete, so use presentation
    attributes.
  - Thin rounded borders drop out at fractional scales.

  Check SVG with `hotty render`, headless at 1.6x.

### Legs

- [ ] **KIT-SEL** — The select keeps its arrows by binding them to
  `program` on its button (`data-keys`, SPEC §10.2 "Keys for the program",
  hotty 3c9b169; hotty-go 1dc9bea), instead of handing over with `a=blur`
  and `a=focus`. That also fixes the closed select, whose arrows a
  scrolling host takes. Decide Tab first: without the handover the kit
  doesn't hear it. Either the list stays open until a pick, Escape, or a
  click or keyboard loss elsewhere, or the handover stays just for Tab.
  Hosts need hotty-blitz 33b2f9c or later (from the hotty agent,
  2026-10-09).
- [ ] **KIT-02h** — Progress: a smooth bar.
- [ ] **KIT-03h** — Spinner: frames or a CSS animation.
- [ ] **KIT-01h** — Table: a real table, a sticky header, hover, sorting by
  header click.
- [ ] **KIT-04h** — Rich list: two-line rows in proportional type, hover.
- [ ] **KIT-08h** — Key hints as keycaps.
- [ ] **KIT-07h** — Scroll view on the host's scrolling (SPEC §8).
- [ ] **KIT-05h** — Code: highlight spans. Selectable text and a copy
  button wait for the clipboard.
- [ ] **KIT-06h** — Diff: split view in columns, acting on hunks.
- [ ] **KIT-09h** — Tree: disclosure triangles, guides.
- [ ] **KIT-10h** — Chart as SVG, with presentation attributes, not CSS.
- [ ] **KIT-11h** — Suggestions in a surface at a higher z.
- [ ] **KIT-12h** — Selection and copy: the host's own. Blocked: the
  clipboard is parked by the maintainer.
- [ ] **KIT-13h** — Toast and tooltip, each as a surface at a higher z.
- [ ] **KIT-14h** — Timer and stopwatch in display type.
- [ ] **KIT-15h** — Paginator.
- [ ] **KIT-16h** — Confirm.
- [ ] **KIT-19h** — Big text as display type.
- [ ] **KIT-20h** — QR code as SVG.
- [ ] **KIT-17h** — File picker.
- [ ] **KIT-18h** — Terminal.

## Later

- [ ] **KIT-CAT** — A composite catalog: basic's definitions verbatim plus
  the hotty catalog, generated, so an agent can use it as the surface
  default (L1). Do it when the agent integration starts, which the
  maintainer has on hold.
- [ ] **KIT-WAIT** — A spinner where a value is still to come: a component
  whose prop is an agent function's value not yet returned. Split from
  KIT-03c: the core resolves function calls at once and has no pending
  state to draw, so it needs one first. Do it with the agent integration,
  where such calls come from.
- [ ] **KIT-FIT** — Re-check `html.Rendition.SetFit` (`k-fit`) on hottyterm
  26.10.09. That build ships hotty-blitz 050d4ba with the `100vh` fit fix.
  Keep `k-fit` unless it is now dead weight.

## Parked

- **KIT-21** — 3D and audio. These are not for a terminal UI kit. Media is
  gov R-2's research.

## Done

- [x] **KIT-01c** — `HottyTable` in cells (journal 2026-10-09.7, awaiting
  the maintainer's look).
  - Columns with widths, alignment and `…` cuts. Rows come as data,
    bound or literal.
  - The selection is a row's `rowKey`, written to the bound `selected`.
    The arrows, Page Up, Page Down, Home and End move it. Enter or a
    second click runs `onActivate` (L2).
  - `height` gives a fixed body, which scrolls under the header to keep
    the selection in view. The rule says which rows show.
  - The HTML baseline is a real table with the same window of rows. Keys
    and clicks work on a host, and a key vector covers both renditions.
- [x] **KIT-02c** and **KIT-03c** — `HottyProgress` and `HottySpinner` in
  cells, on a frame clock (journal 2026-10-09.5, awaiting the maintainer's
  look).
  - Progress: a label, a bar in eighths blending from info into accent,
    the percentage, success once done, and a sliding segment without a
    value.
  - Spinner: bubbles' twelve frame sets at bubbles' rates. The label
    stays put, and the spinner stops when `active` is false.
  - The clock: `Rendition.Animating` and `Clock`, `Book.Tick`, and ticks
    in `storybook -bare`.
  - HTML and text baselines, which don't move (KIT-02h, KIT-03h).
  - A2UI's pending state moved to KIT-WAIT.
- [x] **KIT-LOOK** — The cells look of fields, held against huh and
  bubbles' textinputs (journal 2026-10-09.4, awaiting the maintainer's
  look).
  - What changed: huh's gutter and its `┃` focus bar, bold titles, the
    `> ` prompt, the textarea's bar, `[•]`, one option to a row, the
    select's `▾`, and a blank row between fields.
  - Left for the legs that need them: background roles (KIT-01, KIT-04)
    and the help line (KIT-08).
- [x] **KIT-REF** — Reference shots: `ref/` (its own module), `storybook
  -bare`, and `make shot NAME=…` (journal 2026-10-09.3). It has references
  for form, textarea, progress, spinner, table and list. Each later leg adds
  its own reference to `ref/gaps.go` with its story. OpenTUI references are
  used only where Bubble Tea has no equivalent (Code, Diff, QR, big text),
  and that leg adds them.
- [x] **KIT-00** — The `Hotty` prefix: HottyShortcut, HottyForm, hottyFocus
  and hottyBlur, with the rule in profile §6 and the catalog's
  `instructions` (journal 2026-10-09.2).
