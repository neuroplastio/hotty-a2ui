# Board

Last updated: 2026-10-09

**Active phase:** 1 — Cells parity ([roadmap](../roadmap.md))
**Waiting on the maintainer:** a look at KIT-LOOK (journal 2026-10-09.4):
fields drawn as huh draws them; at KIT-02c and KIT-03c (journal
2026-10-09.5): progress bars and spinners; at KIT-01c (journal
2026-10-09.7): the table; at KIT-04c (journal 2026-10-09.9): the list;
at KIT-08c (journal 2026-10-09.10): the key hints; and at KIT-07c
(journal 2026-10-09.11): the scroll view. Phase 1 is a go
(2026-10-09). New names
follow [catalog-naming](../knowledge/catalog-naming.md).

## Legend

- `KIT-NN` is a gap ([gap-analysis](../knowledge/gap-analysis.md)). The
  number is its priority.
- `c` is its cells leg (phase 1), and `h` its HOTTY leg (phase 2).
- The sections below list the legs in the order they're built. The first
  open leg that nothing blocks is next.

## Phase 1 — Cells parity

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
- [ ] **KIT-01h** — Table: a real table, a sticky header, hover, sorting by
  header click.
- [ ] **KIT-04h** — Rich list: two-line rows in proportional type, hover.
- [ ] **KIT-08h** — Key hints as keycaps, following the host's focus. On
  a host the hints leave out what the host works itself (fields, boxes,
  Buttons), since it moves focus without telling the renderer (journal
  2026-10-09.10). Probe whether Blitz restyles `:has(:focus)` on a focus
  change; the other way, a host that names the focused element, is a
  SPEC change and so the maintainer's call.
- [ ] **KIT-07h** — Scroll view on the host's scrolling (SPEC §8). The
  baseline box scrolls natively but starts at its top: a program cannot
  set a host's scroll offset (SPEC §5.3), so `follow` and `hottyScrollTo`
  do nothing there (journal 2026-10-09.11). Probe a `column-reverse` box:
  Blitz draws it at its end, but whether it then scrolls and stays at the
  end as lines are appended is unchecked. Otherwise following the tail
  is a SPEC change, the maintainer's call. The wheel over a host's box is
  the host's. bubbles' letter keys (j, k, b, f, u, d, g, G) do nothing on
  a host either: the arrows work because the host scrolls with them, and a
  keymap outside a text field gives only `program` (SPEC §10.2). Reported
  by the maintainer 2026-10-09; the fix proposed is scroll actions in the
  keymap, a SPEC change.
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

- [x] **KIT-07c** — `HottyScrollView` in cells, as bubbles' viewport
  (journal 2026-10-09.11, awaiting the maintainer's look).
  - A box `height` rows tall of a `child` or of `lines` (L4: an agent
    writes each line to the next index), cut and scrolled sideways, or
    wrapped. `follow` keeps the tail in view until the user scrolls up.
  - A scrollbar column, which bubbles lacks; the keys and the wheel as
    bubbles' viewport takes them; the hints are its help.
  - `hottyScrollTo(id, start|end)` (L8).
  - A child shows through the box: hidden parts take no click, and the
    focused element is scrolled into sight.
  - The HTML baseline is a box the host scrolls. A new line is an
    `append` delta. `follow` and `hottyScrollTo` wait on KIT-07h.
- [x] **KIT-08c** — `HottyKeyHints` in cells, as bubbles' help (journal
  2026-10-09.10, awaiting the maintainer's look).
  - The keys of what has the keyboard, then the labelled HottyShortcuts,
    then `? more`; `?` shows the full view in columns, a text field's
    editing keys among them, from its keymap.
  - It follows the keyboard; the agent places it once. `toggle` turns
    `?` off where the surface needs it.
  - The list story's help line is one now. A new story, hotty/keyhints,
    tabs through a note form.
  - The HTML baseline draws the same line with `kbd`s. On a host only a
    Table's, a list's, a Slider's and a select's keys show (KIT-08h).
  - The Table takes k, j, g and G, as bubbles' does.
- [x] **KIT-04c** — `HottyList` in cells, as bubbles' list (journal
  2026-10-09.9, awaiting the maintainer's look).
  - A title, a status line, and items of a label and a description,
    literal or bound. The selection is an item's `value`, written to the
    bound `selected`; Enter or a second click runs `onActivate` (L2).
  - `height` gives pages, with dots; the arrows (and j, k, h, l, g, G)
    move and turn pages as bubbles' do.
  - `filterable`: `/` types a filter, ranked as bubbles ranks it
    (sahilm/fuzzy), with matched characters underlined; Enter applies it,
    Escape drops it. The query is the renderer's state, not the data
    model's.
  - The HTML baseline shows the same page. The filter is typed through
    the program's keys there too, since a focused box uses none. A key
    vector covers both renditions.
- [x] **KIT-02h** and **KIT-03h** — Progress and Spinner move on a host,
  taken ahead of phase 2 because the maintainer saw the indexing bar stand
  still there (journal 2026-10-09.8).
  - Both renditions keep the same clock (`view.ProgressStep`,
    `SpinnerSet.Frame`), so cells and HTML show the same frame, side by
    side.
  - A tick is one delta: the Progress's `--k-at` attribute, the
    Spinner's frame's text (`~f`). `html.Rendition` has `Clock` and
    `Animating`, and `Book.Tick` counts the surfaces it placed.
  - The bar steps 2.5% of the track at 10 ticks a second, as cells
    does. A CSS animation would be smoother, but hotty-blitz can't run
    one yet (journal 2026-10-09.8, knowledge/host-motion.md).
  - hotty-demo returns `Book.Tick` since a4afa21, so the site's
    storybook moves too.
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
  - HTML and text baselines. HTML moves since KIT-02h and KIT-03h; text
    stays still.
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
