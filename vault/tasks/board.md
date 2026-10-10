# Board

Last updated: 2026-10-10

**Active phase:** 1 — Cells parity ([roadmap](../roadmap.md))
**Waiting on the maintainer:** a look at KIT-LOOK (journal 2026-10-09.4):
fields drawn as huh draws them; at KIT-02c and KIT-03c (journal
2026-10-09.5): progress bars and spinners; at KIT-01c (journal
2026-10-09.7): the table; at KIT-04c (journal 2026-10-09.9): the list;
at KIT-08c (journal 2026-10-09.10): the key hints; at KIT-07c
(journal 2026-10-09.11): the scroll view; at KIT-05c (journal
2026-10-09.12): code and its colours; and at KIT-06c (journal
2026-10-09.13): the diff; at the icons (journal 2026-10-10.1,
story `basic/icons`); at KIT-09c (journal 2026-10-10.2): the tree; at
KIT-BOOK (journal 2026-10-10.4): the storybook's nav and tabs; at
KIT-22c (journal 2026-10-10.5): the switch; at KIT-10c (journal
2026-10-10.6): the charts.
Phase 1 is a go
(2026-10-09). New names
follow [catalog-naming](../knowledge/catalog-naming.md).

## Legend

- `KIT-NN` is a gap ([gap-analysis](../knowledge/gap-analysis.md)). The
  number is its priority.
- `c` is its cells leg (phase 1), and `h` its HOTTY leg (phase 2).
- The sections below list the legs in the order they're built. The first
  open leg that nothing blocks is next.

## Phase 1 — Cells parity

- [ ] **KIT-23c** — Drag and drop: a list, a table, a tree and a
  templated List reorder and move by a mouse, with a line where the item
  will land, and by Alt with the arrows; any component can be a drag
  source or a drop target for the agent
  ([drag-and-drop](../knowledge/drag-and-drop.md), model C with P2,
  picked 2026-10-10).
- [ ] **KIT-SLIDE** — Sliders: a Slider that fills from its end
  (`io_neuroplast_hotty.fill`), `HottyRangeSlider` with two knobs, and
  a story of their own; Progress keeps one slider. The rail stays as it
  is (`━━━━■⎯⎯⎯⎯`), approved 2026-10-10.
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
  - `currentColor` is black where `color` is a `color-mix()` (muted
    text): a fork bug, with the hotty agent (journal 2026-10-10.4).

  Check SVG with `hotty render`, headless at 1.6x.

### Legs

- [x] **KIT-ICON** — Icons on a host are Material Symbols, Sharp, filled,
  as inline SVG in the text's colour, and basic's `{svgPath}` draws (gov
  R-4; journal 2026-10-10.1). The whole pack (`icons/materialsymbols/`)
  comes later.
- [x] **KIT-HICON** — `HottyIcon`: any Material Symbols name a program
  registers (`make icons` from its `icons.txt`; the storybook's 37), the
  59 by Material's names, and an `svgPath` with a `strokeWidth` (journal
  2026-10-10.3). HottyTree's node icons take the same names.
- [x] **KIT-BOOK** — The storybook's nav is a HottyTree of branches
  (Components, Behaviours, A2UI examples, Fallbacks), opened by its
  selection and filtered from pick; panel heads the tabs with the story;
  tabs and options have icons on a host (`io_neuroplast_hotty.icons`,
  profile §6.16; journal 2026-10-10.4). The site's e2e tests owe a change
  at its next pin bump.
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
  the host's. Done 2026-10-09: bubbles' letter keys scroll the box on a host
  too, bound to the scroll actions the maintainer approved after reporting
  that j and k did nothing there (SPEC §10.2, *Scrolling keys*, hotty
  c8bc64f). Hosts have them since hottyterm 26.10.09-dev.0aeeed0, hotty-blitz
  f22faf4 and xterm-addon-hotty 83febf2.
- [ ] **KIT-05h** — Code: highlight spans. Selectable text and a copy
  button wait for the clipboard.
- [ ] **KIT-06h** — Diff: split view in columns, acting on hunks.
- [ ] **KIT-09h** — Tree: disclosure triangles, guides.
- [ ] **KIT-10h** — Chart on a host past the baseline, whose line is
  already an SVG with presentation attributes (journal 2026-10-10.6): its
  box at the plot's real size, not a 480-wide guess that thickens a steep
  stroke on a wide plot; the value under the pointer.
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
- [ ] **KIT-22h** — Toggle: a pill whose knob slides by deltas.
- [ ] **KIT-23h** — Drag and drop on a host: SPEC §9.1 drags, with a line
  where the item lands. Blocked on G1 (where in the target) and G2 (a
  target that isn't draggable), with hotty since 2026-10-10.

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

- [x] **KIT-10c** — `HottyChart` and `HottySparkline`, held against
  ntcharts (journal 2026-10-10.6, awaiting the maintainer's look).
  - A chart is a line (`kind` line, braille) or bars (eighths of a row,
    grouped a point), of `values` or `series`, literal or bound, with
    `labels`, `window`, `min`, `max` and `height`. The agent adds a point
    by writing the next index (L4).
  - The axis is the view's, fitted by hotty-go's `series.Scale` with its
    ticks a free row apart; bars hang below 0, which may fall inside a
    row. Ticks `┤`, a legend, the series in info, warning, success, error.
  - A sparkline is a column a value (hotty-go's `blocks`), from 0 or its
    lowest value; a `window` keeps its width as values arrive.
  - On a host the line is an SVG path from hotty-go's `chart.Line`, a new
    point one attribute's delta; bars and sparklines are boxes. Text is
    a table of the values.
- [x] **KIT-22c** — `HottySwitch`, an on/off switch for a setting, held
  against huh's Confirm, inline (journal 2026-10-10.5, awaiting the
  maintainer's look).
  - A `label`, a bound `value` (a boolean), `checks` and `disabled`;
    these take literals or paths, not basic's functions (L1).
  - Cells: `▬▬■` in the accent when on, `□⎯⎯` muted when off (the
    maintainer's pick, 2026-10-10), in a
    field's frame; faint when disabled.
  - Space, Enter or a click flips it, Enter in a HottyForm too; a
    disabled one is skipped by Tab.
  - On a host a `role=switch` button, a pill with a knob, Material's
    ringed off state. The knob's slide is KIT-22h.
- [x] **KIT-09c** — `HottyTree` in cells, held against bubbles' tree
  (journal 2026-10-10.2, awaiting the maintainer's look).
  - Nodes of a label, an icon (a host's), a value and children, literal
    or bound. `selected` and `expanded` are written where bound;
    unbound, the folding is the renderer's. The selection's branches
    show open.
  - lipgloss's guides, bubbles' `▶`/`▼` folds, a closed branch's count;
    the list's `│` bar on the selected node, bold in the accent.
  - ↑ ↓ (k j) move, → (l) opens then goes in, ← (h) closes then goes
    up; Enter and Space fold a branch or act on a leaf. A click selects,
    folds a branch, and acts on the selected leaf.
  - `filter`, bound to a field's path: what matches, the branches to it
    and what it holds, all open, matches underlined.
  - On a host a `role=tree` box, indented, with folds and icons.
- [x] **KIT-06c** — `HottyDiff` in cells, held against OpenTUI's Diff
  (journal 2026-10-09.13, awaiting the maintainer's look).
  - A patch of one or more files, or two texts compared (package
    `diff`, go-udiff's LCS). Unified, or split where each side has 16
    columns of code.
  - File names with `+N -M`, hunk headers, both line numbers, unchanged
    runs folded into a `⋯ N unchanged lines` row.
  - Changed lines are tinted a sixth toward their role, and changed words
    a third, with delta's distance rule. Without a theme background, the
    line's text takes the role colour and its words are reversed, as
    git's diff-highlight does.
  - Select-then-act on a hunk (L2): `selected` is `path:line`, and
    `onActivate` runs on Enter or a second click. On a host the selected
    hunk takes the host's focus, so that the host scrolls it into view.
  - It costs the site's storybook module 27 KB brotli.
- [x] **KIT-05c** — `HottyCode` in cells, held against glamour's code
  blocks (journal 2026-10-09.12, awaiting the maintainer's look).
  - chroma lexes; each token kind takes a role, so every theme and the
    terminal's own palette colour code. 37 languages come with the kit
    (`highlight/lexers`). A program that wants all of chroma's imports
    `highlight/all`.
  - Line numbers, `startLine`, marks (highlight, added, removed, error,
    warning) as a sign and a tinted row, and wrapping or cutting.
  - Text's fenced code blocks are highlighted too, in both renditions.
  - The HTML baseline is the same rows, with token spans and tints.
  - It costs the site's storybook module 185 KB brotli.
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
