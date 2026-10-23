# Board

Last updated: 2026-10-23

**Active phase:** 1 — Cells parity ([roadmap](../roadmap.md))
**Waiting on the maintainer:** nothing from rounds 7–9
([round 9](../feedback/2026-10-11-review-round-9.md)): they settled the
timer's clock format, plain dots, page mode's live controls, the kit's
big letters and paginator, the title's shadow, the question's blank row,
and a docs page as the surface itself, with no box (journal
2026-10-11.5). Every pair asks one question about the one difference its
sides show. Still to look at from earlier rounds: round 1 left out the
icons (journal 2026-10-10.1, story `basic/icons`); KIT-BOOK (2026-10-10.4):
the storybook's nav and tabs; KIT-10c (2026-10-10.6): the charts; KIT-SLIDE
(2026-10-10.7): the sliders, and its four open points; KIT-23c
(2026-10-10.8): drag and drop in cells, and its three proposals; and the
proposals of KIT-11c (2026-10-10.9) and KIT-13c (2026-10-10.10).
KIT-24's look question left (2026-10-10.29): GitHub's colours as the
default palette. From the dev-server leg (2026-10-11): the hover look's
strength in cells and on a host, the grip's `◢` (a full-cell wedge in
hottyterm, a small triangle in xterm.js), and the selection's colour.
From hotty-web's tour form (2026-10-17.2): chips that pick one on their
label's row when they fit (`Strategy   (•) Canary   ( ) Blue/green`),
against a row each. (Live-looking controls, round 7 kept: "should look
usable". A heading level for tab titles and the page's origin, the docs
answered: `PageOptions`, 2026-10-11.1.)
Phase 1 is a go
(2026-10-09). New names
follow [catalog-naming](../knowledge/catalog-naming.md).

## Legend

- `KIT-NN` is a gap ([gap-analysis](../knowledge/gap-analysis.md)). The
  number is its priority.
- `c` is its cells leg (phase 1), and `h` its hotty leg (phase 2).
- The sections below list the legs in the order they're built. The first
  open leg that nothing blocks is next.

## Phase 1 — Cells parity

- [x] **KIT-12c** — Selection and copy in cells (OSC 52): a drag on a
  Text's or a HottyMarkdown's text selects it within that element, in
  reading order, and the release copies it (`Copied`, `cells.OSC52`;
  profile §3.7). Asked by the maintainer through the docs (2026-10-11).
  Word and line clicks (2026-10-13, the maintainer): a double click
  selects the word, a triple its line, in text controls and in a Text or
  a HottyMarkdown (its paragraph, copied at once). The rendition counts
  presses, since a terminal's mouse reports carry no count and none is
  planned. On a host, the HTML rendition reads SPEC §9's `count` (hotty
  7eaa22d, hotty-go 1d58e1a). On items, a run of clicks acts once
  (`view.Controller.ClickItem`, journal 2026-10-13.4).
- [ ] **KIT-12d** — Copy as a page copies (the maintainer, 2026-10-13:
  "I want both hotty and cell modes to feel browser-like", with Ctrl+C
  and Cmd+C overridden, and an option to change them). Cells is done
  (journal 2026-10-13.6): a selection goes to the primary selection, and
  a copy key puts it on the clipboard (`SetCopyKeys`, the storybook's
  `-copy`). On a host the selection is the host's, so hotty-blitz,
  hottyterm and the addon carry it; that went to hotty. web wires the
  site's storybook. Cut and paste followed in cells (journal 2026-10-14.1:
  "Copy is working well, nice. Now, we need Cut and Paste"): a cut key
  cuts a text control's selection, a paste key reads the clipboard with
  OSC 52's query, and a bracketed paste types too (`SetCutKeys`,
  `SetPasteKeys`, `-cut`, `-paste`). The host's cut and paste went to
  hotty with its copy. Undo and redo followed (journal 2026-10-14.2):
  each text control has its own history, as on a page (`SetUndoKeys`,
  `SetRedoKeys`, `-undo`, `-redo`). Then the keys moved onto hotty's
  keymap (journal 2026-10-14.4, hotty 9f534fc): the edit actions are
  keymap actions bound to `Mod`, the controller's `Platform` says what
  `Mod` is (Control's where no host says), a host's edit commands arrive
  as `Rendition.Command`, and `-keys` replaces the five key flags (gone
  since, with the Keys picker: journal 2026-10-15.2). Left:
  a middle click pasting the primary selection; key sequences and the
  which-key menu (§10.5, §10.6's `a=menu`), the next leg.
- [x] **KIT-17c** — File picker, with the profile's rule for granted roots
  (L10).
- [x] **KIT-DEF** — A cue for the default button (the maintainer, round
  7): a HottyForm's default is the Button whose action is its onSubmit;
  cells draws a bar at its left edge (`▏Yes `, `▎` while it has the
  keyboard), a host a line inside its left edge, whatever has the
  keyboard (profile §3.4, §6.2; journal 2026-10-13.2).

- [x] **KIT-CYCLE** — A select stepped in place, `‹ staging ›`, as the
  tour's form draws it (the maintainer, 2026-10-17), the value centred
  between the arrows: `io_neuroplast_hotty.cycle` on a ChoicePicker, in
  cells and on a host; story `hotty/choice` (journal 2026-10-17.1). Its
  look points wait for the next feedback round.
- [x] **KIT-SLIDERUN** — A Slider's and a HottyRangeSlider's labels join
  a run of one-line fields' in cells, their tracks starting where the
  inputs do (hotty-web, 2026-10-17; journal 2026-10-17.2).
- [x] **KIT-PICKER** — A command palette's picker, `HottyPicker`, for
  plexos's palette (the maintainer, 2026-10-18, plexos d/palette-picker:
  "I don't think palette fits into any existing components except text
  input with auto complete"). A query, the rows the agent ranks for it
  (the picker never filters), actions on the selected row by keys that
  type nothing (`enter`, `mod+enter`, `alt+enter`, `shift+enter`,
  `shift+delete`) and an actions panel (`panelKey`). Cells, text and a
  host baseline; story `hotty/picker` with the palette's agent behind it;
  vectors on both; `ref picker`; profile §6.32 (journal 2026-10-18.1).
  Awaiting the maintainer's look.

### plx's settings page (2026-10-23)

What plx's settings page (plx 709b743c) ran into, which the maintainer
had plx-settings send here; the design calls are the kit's. A leg each,
in this order; plx bumps its pin as they land.

- [x] **KIT-HIDE** — `io_neuroplast_hotty.hidden`, a DynamicBoolean that
  shows or hides a component as the data model changes, in all three
  renditions; the keyboard on it goes on to what follows it.
  `accessibility.hidden` is a screen reader's alone again: cells hid it,
  a host drew it. Profile §6.34, instruction 35 (journal 2026-10-23.1).
- [x] **KIT-MUTED** — Muted at 4.5:1 on the terminal's ground:
  `--k-muted` mixes 78% of the fg, not 62% (hottyterm's default palette:
  3.5 → 4.6:1); a palette with dim text stays under it. Profile §2
  (journal 2026-10-23.3); the shade awaits the maintainer's look.
- [x] **KIT-WEIGHT** — A Row's weights the same in cells as on a host and
  in A2UI's own renderer: weighted children share what the others leave,
  from nothing, each no narrower than its least. Profile §3.3 (journal
  2026-10-23.2); the icons grid and the kanban await the maintainer's look.
- [x] **KIT-FILL** — `io_neuroplast_hotty.stretch` on the root: as tall
  as the surface (cells: `Rendition.SetRows`), the rows left to its
  weighted children; a weighted HottyScrollView pages at what it gets.
  plx drops its computed height. A weight no longer drops a box's
  `--k-rows` on a host. Profile §6.35, instruction 36 (journal
  2026-10-23.4); hotty/stretch awaits the maintainer's look.
- [x] **KIT-PLAIN** — HottyScrollView `variant: "plain"`, a screen's
  body: no fill, corners, padding or ring. Its focus is on its
  scrollbar: on a host a 2px line down its right edge and the thumb in
  the focus colour; in cells the accent thumb, the whole bar where all of
  the content shows (any box's, which showed no focus there). Profile §2,
  §3.4, §6.11, instruction 11 (journal 2026-10-23.5). hotty/stretch's
  sections are plain. Look question: an overlay scrollbar (Blitz's)
  shows nothing until the box scrolls, so on a host only the cut at its
  foot says the body scrolls; rules above and below it are the option.
- [x] **KIT-GAP** — `io_neuroplast_hotty.gap` on a Column, a Row or a
  List: whole terminal rows (or a Row's columns), 0 to 8, in place of
  the kit's own spacing in both renditions; on a host `k-gap` and
  `--k-gap` in `--hotty-cell-h`/`-w`. Profile §6.36, instruction 37
  (journal 2026-10-23.6). Story hotty/gap: settings a row apart, no
  Dividers.
- [x] **KIT-WIDTH** — `io_neuroplast_hotty.width` on a TextField or a
  DateTimeInput: its value's columns, 1 to 256, in place of a text
  field's 32 or a DateTime's form; a longText starts that wide. On a host
  `k-cols` and `--k-cols` in `ch`, at most the field's. Profile §6.37,
  instruction 38 (journal 2026-10-23.7). Story hotty/width. A field
  without one still fills its field on a host and stops at 32 in cells,
  as before. Shared columns come of KIT-WEIGHT. A number field has no
  spin buttons on a host, which cut a port's 5 columns in Chromium
  (journal 2026-10-23.9).
- [x] **KIT-NAV** — HottyList `variant: "compact"`, a nav's: an item a
  row, its description beside its label (cut first), its title bold in
  muted, no status line, no blank rows, its dots under its items; keys,
  filter, pages, moves and clicks the list's. On a host `k-compact`.
  Profile §6.9, instruction 9 (journal 2026-10-23.8). Story hotty/nav;
  the title's look awaits the maintainer.

Two are hotty-blitz's, passed to the hotty agent: `::placeholder` is not
drawn, and an empty `<input type=number>` is a sliver. Both fixed on
hotty-blitz main at 2a5904a (2026-10-23); hottyterm has them once it
re-pins.

### plx's onboarding (2026-10-23)

The maintainer, through plx-settings, on plx's onboarding (plx 807e2e3a)
on a hotty host: "Buttons are too thin".

- [x] **KIT-SIZE** — `io_neuroplast_hotty.size` on a Button: `large`, a
  page's call to action, or `medium`, the default. On a host `k-large`:
  at least 3rem tall where a medium one is 2.25rem (a field's), 1.5rem
  each side; borderless ones as tall, so a Back/Skip/Continue row lines
  up; a List's row keeps its size. In cells a row, as before. Profile
  §6.38, instruction 39 (journal 2026-10-23.10). Story hotty/size; how
  large awaits the maintainer's look.

## Phase 2 — The hotty layer

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
- [ ] **KIT-01h** — Table: a real table, a sticky header, sorting by
  header click, the wheel on a host (the surface scrolls itself today, so
  the wheel never reaches the program's window of rows). Hover landed
  2026-10-11 (kit.css `:hover`).
- [ ] **KIT-04h** — Rich list: two-line rows in proportional type. Hover
  landed 2026-10-11.
- [ ] **KIT-08h** — Key hints as keycaps, following the host's focus.
  Keycaps done 2026-10-22 (profile §6.10, *Keycaps*): per-system labels,
  glyphs or text in either rendition, the user's own labels, in the hints
  and the picker. plx sets `Controller.Keycaps` from its config; in the
  storybook, the story hotty/keycaps's controls do (2026-10-23, no flags).
  Blitz breaks a nowrap line after an inline-block, so an overflowing
  hints line wraps its first caps on a host: a hotty-blitz fix. Open: on
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
- [ ] **KIT-11h** — Suggestions in a surface at a higher z, with the
  ghost text. The baseline is a `datalist` (journal 2026-10-10.9), which
  Blitz doesn't draw; check the addon's. Taking one by Tab and picking by
  Enter need the blur and focus handover or a SPEC change: a host keeps
  Tab (§10.2) and leaves a focused field's value alone (§6.2). Decide
  with KIT-SEL.
- [ ] **KIT-12h** — Selection and copy: the host's own. Blocked: the
  clipboard is parked by the maintainer.
- [ ] **KIT-13h** — Toast and tooltip, each as a surface at a higher z.
  The baseline (KIT-13c) is a region cut at the surface's edges and the
  tooltip in the key hints by SPEC §9.4's hover; a tooltip beside its
  element needs the element's area, which hover does not give.
- [ ] **KIT-14h** — Timer and stopwatch in display type.
- [ ] **KIT-15h** — Paginator.
- [ ] **KIT-16h** — Confirm.
- [ ] **KIT-19h** — Big text as display type.
- [ ] **KIT-20h** — QR code as SVG.
- [ ] **KIT-17h** — File picker.
- [ ] **KIT-22h** — Toggle: a pill whose knob slides by deltas.
- [ ] **KIT-23h** — Drag and drop on a host: SPEC §9.1 drags, with a line
  where the item lands. The baseline landed on 2026-10-13, at the
  maintainer's ask (journal 2026-10-13.3): a mouse or a pen drags items,
  rows, nodes and a List's children, and the line works as in cells
  without pixels. G1 and G2 landed the same day (hotty 155b24a, caps
  `drop`; hotty-go 649c60a; journal 2026-10-13.5): halves, a folder's
  thirds, an empty list and drop-only targets on a host that has `drop`.
  Hosts: blitz ea72902 and addon 0be7c8f are on main, while hottyterm
  7d38a45 and npm 0.1.0-next.13 wait on the maintainer. Still to do:
  the grips, for touch.

## For the docs (gov NEIO-14)

Asked by web on 2026-10-10, on the maintainer's go-ahead. The maintainer
OK'd them for the kit "as long as it fits the framework": A2UI's extension
points only (no forked basic component), SPEC changes through hotty first,
hotty-go protocol-only (AGENTS.md). Every project's docs site draws its
components with the kit, as a plain web page, as hotty surfaces (xterm.js
with the addon) and in cells (NEIO-14, a draft, *On the kit*; web's pilot
is its branch `docs-pilot`). None of these blocks the
docs: until the kit has a piece, the docs module draws it and switches when
it lands. The docs also wait on KIT-09h, KIT-05h and KIT-01h. They wanted
Terminal to play an asciicast, which is parked (KIT-18): until it comes
back, their own player draws screencasts and one-frame hotty snapshots
through hotty-go's `hottyvt` (web's `shared/screencast`).

- [ ] **KIT-24p** — A page host, split from KIT-24: the addon's surface
  code (DOM, events, deltas) without the terminal, so that a wasm program
  running the kit makes page mode's HTML live. Progressive enhancement,
  never needed to read.
## Later

- [ ] **KIT-SIZE** — A program carries only the components it uses: the
  renditions split into nested Go modules (the maintainer, 2026-10-11:
  "optimize later with nested go.mod modules"). Not before the kit's
  components settle. Measured by hotty-web, 2026-10-17: a form in the
  site's TinyGo main module (cells, html, view) took it from 6.9 MB to
  8.9 MB raw, with chroma and regexp2, icons, qr, diff, fuzzy and uax29
  linked though the form uses few of them (a HottyCode, so chroma). Served
  (TinyGo, wasm-opt -Oz, brotli -q 11): 1,645,197 bytes without the kit,
  2,166,970 with it, +0.52 MB (+32%), for a cycle, a Slider, chips, two
  Buttons, Text and a HottyScrollView of HottyCode.
- [ ] **KIT-STACK** — A form's first key or click (cells `Key`, or html
  `Event` then `Update`) overflows TinyGo's 64 KB goroutine stack (hotty-web,
  2026-10-17); the site's main module runs it at `-stack-size=128KB`.
  Find the deep path (the html diff or the cells layout, hotty-web
  guesses) and make it lighter, so the site can drop the flag.
- [ ] **KIT-REDRAW** — cells `Draw` lays a long HottyMarkdown out whole on
  every draw, a hover's too (docs, 2026-10-17): on hotty's SPEC page at
  ~100 columns a link's hover draw is 93 ms of wasm in Chromium, natively
  ~100 ms and 180 MB allocated, 85% in layout.measureHeight/lines
  (runGlyphs, wrapConts, glyphs, splitLines), about half of the wasm time
  Go's GC. Keep each element's lines at a width across draws while the
  document, the width and its non-hover state hold, so a hover only
  repaints (paintHover over the last base frame); reuse the frame, canvas
  and cell buffers. Repro: web docs-pilot on hotty-0.4, `make serve
  SITE=hotty CHECKOUT='../../../{name}/hotty-0.4'`, Text mode, pointer on
  and off a link.
- [x] **KIT-TOUCH** — A finger drags a Slider in cells (the maintainer,
  via hotty-web, 2026-10-17: "in cell mode slider touch dragging doesn't
  work in xterm and in hottyterm"). The mouse drag worked
  (cmd/storybook TestSliderDragInCells, SGR reports), but a touch drag
  over the cells is the terminal's scroll (SPEC §9). hotty 0.4's V4-02,
  SPEC §9.6 "Touch in the cells", lets a program name cells that take a
  finger's drag (`a=touch` areas). The kit names its enabled Sliders' and
  HottyRangeSliders' tracks (`cells.TouchAreas`, profile §3.7), and the
  storybook sends them (`storybook.Options.Mouse`; 4b1f6d3, by the hotty
  agent). On main with hotty 0.4, 2026-10-18.
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
- [ ] **KIT-28** — Context menu (right click): application-level context menu
  support on secondary click (right click). Available by default in text inputs
  (`TextField`, `longText`) and wherever text selection is supported (KIT-12;
  `Text`, `HottyMarkdown`, `Code`). Menu items follow standard browser text input
  context menus: Undo, Redo, separator, Cut, Copy, Paste, Delete, Select All,
  with mnemonic accelerators (`_U_ndo`, `_R_edo`, `Cu_t_`, `_C_opy`, `_P_aste`,
  `_D_elete`, `Select _A_ll`) and dynamic disabled/enabled states based on
  selection, clipboard content, and field mutability. Must work both in cells
  (drawn popup overlay box) and on a hotty host (surface/overlay at a higher z)
  without relying on native browser context menus when running in an xterm.js /
  web terminal addon. Reference shot: `vault/knowledge/context-menu.png`.
- [ ] **KIT-29** — Horizontal scroll support: lateral scrolling and indicator
  support for components that overflow available width or contain wide tabular,
  hierarchical, or preformatted content. Targets components that benefit from
  horizontal viewport panning:
  - `HottyTable`: wide tables with many or wide columns keep minimum readable
    widths without aggressive crushing (`fitWidths` down to 0-width dropped
    columns); optional frozen key column.
  - `HottyTree`: deeply nested branches and long labels pan sideways without
    clipping or wrapping tree hierarchy.
  - `HottyFilePicker`: multi-column file listings (name, size, permissions,
    modified) or long paths in narrow panes.
  - `HottyScrollView`: true two-axis scroll container; scrolls child components
    horizontally (`scrollX` / `overflow-x`) with horizontal scrollbar (`─` track
    with `━` thumb or `◄` `►` overflow markers) along the bottom.
  - `HottyCode` & `HottyDiff`: unwrapped long lines with bottom scroll indicator
    and keyboard panning (Shift+arrows or h/l).
  - Overflowing `Row` and `Tabs`: horizontal strips that pan when items exceed
    container width.
  Supports Shift+Wheel, horizontal mouse wheel / trackpad gestures, keyboard
  navigation, and drawn horizontal scrollbars across cells and hotty host renditions.
- [ ] **KIT-17a** — The file picker for apps: what bubbles' filepicker and
  huh's field give an app that HottyFilePicker doesn't (the gap analysis of
  2026-10-13, against bubbles v2.2.1 and huh v2.0.3). All of it is props
  and the program's API, with nothing new on the wire. In order:
  1. Grants per app, not per `view.Controller`: one place where a program
     grants roots for every surface it shows (only `story.Run` grants
     today), and a helper that turns a root-relative pick into the
     program's own path.
  2. What can be picked: allowed types (others muted, and a pick of one
     shows the field's error, as huh's "X files only"); files, directories
     or both (`FileAllowed`, `DirAllowed`: a "choose a folder"); hidden
     files left out unless asked for (`ShowHidden`).
  3. A form field: `label`, `checks` and their error, the picked file
     marked in the listing, and a closed state that shows the picked path
     or a placeholder and opens to pick, as huh's does.
  4. The entry the cursor is on, bound (`HighlightedPath`), so that an app
     can preview it as the user moves.
  5. Symbolic links: `name → target`, and a link to a directory opens it
     (Go 1.25's `fs.ReadLinkFS`, which `os.DirFS` and `os.OpenRoot`'s FS
     have). Today one is listed and picked as a file.

  Then: reading without blocking a build (bubbles reads in a `tea.Cmd`; a
  slow or remote root stalls the UI), the permissions column, a height
  that fills the room (`AutoHeight`), sizes as bubbles writes them (SI,
  `1.2kB`, where ours are `1.2K` in binary units), and a cap or a window for
  very large directories on a host (deferred in journal 2026-10-13.1).
  Beyond both references, the maintainer's call: picking several files, a
  save dialog (a name, a new folder, an overwrite question), sorting by
  date, typing a path.

## Parked

- **KIT-18** — Terminal, cells and host (KIT-18c, KIT-18h), shelved by the
  maintainer on 2026-10-13. What was sketched, for when it comes back: a
  complete VT, not a log view, on charmbracelet's `x/vt` (web's
  `internal/vterm` runs its panes on it) or hotty-go's `hottyvt`: colours,
  the cursor, the alternate screen, the child's mouse modes, a scrollback
  the kit keeps. Input while focused: keys encoded for the child (web's
  vterm fixes, kitty flag 1), the mouse only when the child asks, the size
  following the box, bracketed paste (the clipboard on a host is parked),
  and one reserved key that hands the keyboard back, the maintainer's
  choice and the first decision. A "registered command" (L10) as a stream
  each way plus resize, which the program provides: a process in a PTY, an
  in-process Go program, a wasm worker, or an asciicast player (the docs'
  ask; wasm has no processes). The agent names one by id, never a command
  line or keystrokes, and sees the exit status, the output only if the
  program shares it. Open for the host leg: a VT grid in an HTML surface
  may be heavy for full-screen apps, and drawing it natively is a SPEC
  question; a hotty app inside the pane would need its surfaces relayed, as
  vt-apps does. Pick up the docs' frame format (web docs-pilot
  `internal/view`: a Cast, then a Frame per change) for the player.
- **KIT-21** — 3D and audio. These are not for a terminal UI kit. Media is
  gov R-2's research.

## Done

- [x] **KIT-25c/25h**, **KIT-26c/26h** — `HottyMarkdown` and links in place
  (journal 2026-10-10.28, profile §6.27, instruction 27): long-form
  Markdown as GitHub reads it: headings with GitHub's anchors
  (github-slugger, numbered across the surface), alerts as callouts,
  tables as a HottyTable's, code as a Text's, images as alt text or
  `<img>`, HTML and comments (the docs' markers) as nothing. A link with no
  scheme goes in place: a Tab stop that Enter or a click follows; a
  `#fragment` the renderer follows to its heading (cells scrolls it to the
  top of its HottyScrollView; a host gets `a=focus`), any other href is
  written to `link` and `onLink` runs. HottyTree nodes take an `href`.
  On a page they are `<a href>` (2026-10-11.3). Story `hotty/markdown`.
  Open (for the feedback loop): IMPORTANT in the accent, no fill for
  code in a HottyScrollView,
  nothing marking the document focused after a jump. The two hottytest
  findings were hottytest bugs, fixed in hotty-go a4025c0, and with
  hotty 0.2 (4b74875, journal 2026-10-11.6) the key vector that was
  cells-only runs on hottytest too. Since 2026-10-11, for the docs' own contents: every
  heading's box in cells (`view.HeadingID`) and `Controller.GoTo`. For
  the maintainer: in cells a link with a scheme is no Tab stop (the
  terminal owns OSC 8), so the keyboard can't reach outside links such as
  "Edit this page". Making them Tab stops whose Enter writes `link` and
  runs `onLink` would change KIT-26's model. The docs keep the model
  meanwhile (2026-10-11).
- [x] **KIT-20c** — `HottyQRCode` (journal 2026-10-10.31, profile §6.29,
  instruction 29): `value`, `errorCorrection` (a minimum, raised when the
  version holds more for free), `label`. Package `qr/` on rsc.io/qr's
  coding (BSD-3), with the version, mode, level and mask chosen as ISO
  18004 has them. Cells draws half blocks with a quiet zone of 4 on paper
  (`Cell.Paper`: black on white whatever the theme, §3.6), or the label
  and value as text when the box is too narrow; a host gets an svg path.
  34 codes decode back from cells (gozxing), and zbar reads every picture.
  Story `hotty/qr`. Open for the maintainer: paper under NO_COLOR (drawn
  reversed today), 256-colour paper at the floor, the raised level, text
  as the narrow fallback.
- [x] **KIT-19c** — `HottyBigText` (journal 2026-10-10.30, profile §6.28,
  instruction 28): `text`, `size` (small, medium, large: 3, 4 or 5 rows in
  cells), `align`. The kit's own pixel fonts (rendition/cells/bigfont.go;
  OpenTUI's come from cfonts, GPL-3.0, not taken), in half blocks; wraps
  between words. A host gets display type, the text as written. Story
  `hotty/bigtext`. Open for the maintainer: `size` over `font`, capitals
  in cells but the text's case on a host, a finer large font, the blank
  row after each.
- [x] **KIT-24** — Page mode (journal 2026-10-10.29, profile §2.1):
  `html.PageCSS()` once in a page's head, `Rendition.Page(PageOptions)` a
  surface's fragment for its body (ids prefixed by the surface's name;
  tab titles as headings of a level the page gives, and its own origin's
  links in the tab, journal 2026-10-11.1), and
  `storybook -page`. The page sets SPEC §8's `--hotty-*` (bg, fg, accent,
  ansi-8, and -9/-3/-2/-6 for the signals); what it leaves out is GitHub's
  light or dark. No program behind it: Tabs are sections, a HottyTree's
  branches are `<details>` open or closed as given, one `pageLink` decides
  every link (relative same tab, absolute new tab, other schemes
  dropped), and controls show their state and take no input. A tree
  node's `href` and a HottyMarkdown's links in place are `<a href>`, its
  headings' ids GitHub's anchors (journal 2026-10-11.3).
- [x] **KIT-15c** — `HottyPaginator` (journal 2026-10-10.25, profile §6.26,
  instruction 26): with a `child` (a List, Column or Row) it pages the
  child's items itself, `perPage` a page, so a turn needs no round trip;
  bare, with `pages`, it writes `page` and runs `onChange` for the
  agent's data. Dots (the page shown's bright, the others faint, as
  bubbles' example) or `3/10`; ← → h l PageUp PageDown Home End; a click
  on a dot; one Tab stop before its page's controls; cells keeps the
  tallest page's height. Story `hotty/paginator`. HottyList's inactive
  dots are faint now too (they share the drawing). For the maintainer's
  look: the faint dots, the focus bar down the whole paginator, Home and
  End beyond bubbles.
- [x] **KIT-14c** — `HottyTimer` and `HottyStopwatch` (journal 2026-10-10.24,
  profile §6.25, instruction 25): a label and the time, bubbles'
  `4m59s` or `clock` `4:59`, `fg` while it counts and `muted` while still;
  `duration` (ms), `running` (bindable; a timer that runs out writes
  false), `interval`, `format`, `onTimeout` (once, at 0). The time lives in
  the surface's state and counts on the rendition's clock
  (`TickTimers`), a hidden one too; hottyStartTimer, hottyStopTimer,
  hottyToggleTimer, hottyResetTimer for Buttons and HottyShortcuts. Story
  `hotty/timer`. A stopwatch's bound `elapsed` (the maintainer's ok,
  2026-10-11; journal 2026-10-11.2) is written before any action runs,
  when it stops and at a reset. A focused Button that becomes disabled
  keeps the keyboard: ok (the maintainer). Open for the maintainer's
  look: the default format.
- [x] **KIT-16c** — Confirm, as a pattern of two Buttons and HottyShortcuts,
  not a component (journal 2026-10-10.26, profile §6.24, instruction 24):
  y and n press, ← → move (`hottyFocus`), Enter picks, the focused one
  filled; stories `hotty/confirm` and `hotty/confirm-form`. Key hints
  merge shortcuts of one label and leave out keys a text field takes; on
  a host a focused Button is filled with the focus colour (every Button).
  For the maintainer's look: the host's focus fill, a blank row before
  Buttons after Text (huh has one). No HottyConfirm component (the
  maintainer, 2026-10-11): it stays a pattern.
- [x] **KIT-27** — Schema validation is optional (journal 2026-10-10.22):
  `a2ui.Validator`, nil by default, and `a2ui/schema` with A2UI's
  (`schema.NewProcessor`), which the conformance suites and the story
  package (the storybook, `-stream`) use. A program that is its own agent
  leaves jsonschema and x/text out: the docs-like probe, 11.4 MB → 9.7 MB
  js/wasm, 2.89 MB → 2.46 MB gzipped. Splitting further is KIT-SIZE.
- [x] **KIT-13c** — Toast and tooltip, story `hotty/toast` (journal
  2026-10-10.10, awaiting the maintainer's look).
  - `hottyToast({message, kind, timeout, id, actionLabel, action,
    surfaceId})` and `hottyDismissToast({id})`, renderer functions for a
    Button and the agent (instruction 23, profile §6.23); an id replaces
    in place; the action is an event whose context is read as it shows.
  - Cells: rounded boxes stacked at the top right, a column in, the
    newest first, the border and mark (`ⓘ ✓ ! ✗`) in the kind's role;
    time on the clock, held under the pointer or the keyboard.
  - Tooltip: `accessibility.description`, a row over the HottyKeyHints'
    keys, for the element under the pointer (mode 1003; SPEC §9.4
    `hover` on a host) or with the keyboard.
  - Host: a region over the surface and the layer, alerts and statuses.
    Text: a line a toast. Vectors in both renditions. The surface at a
    higher z is KIT-13h.
- [x] **KIT-11c** — Suggestions in cells, as bubbles' text input has them
  (journal 2026-10-10.9, awaiting the maintainer's look). Story
  `hotty/suggest`.
  - `io_neuroplast_hotty.suggestions` on a one-line TextField: `options`
    (a DynamicStringList, best bound, which the agent rewrites) and
    `onInput` (an Action at every change), instruction 22, profile §6.22.
  - The renderer narrows to the options that start with the value, case
    aside; the rest of one faint after the caret, a list of up to 5 under
    the field, the highlighted one reversed in the accent.
  - Tab and → take, ↓ ↑ Control+n Control+p move round, Enter picks the
    highlighted one, Escape shuts; nothing highlighted, Enter submits and
    Tab moves on. A taken one is the value as written.
  - A host has a `datalist` (Blitz draws none); text the field alone.
    Vectors in both renditions for the editing, cells only for the keys
    (`renditions:`).
- [x] **KIT-23c** — Drag and drop in cells, model C with P2
  ([drag-and-drop](../knowledge/drag-and-drop.md), "Built"; journal
  2026-10-10.8, awaiting the maintainer's look). Stories under Behaviours
  › Drag and drop: a to-do list, a backlog, a file tree, a kanban board.
  - `reorderable`, `dragType`, `moved`, `onMove` on HottyList,
    HottyTable and HottyTree; `reorder` on a templated List, `drag` and
    `drop` on any component (`io_neuroplast_hotty`, instruction 21).
  - A press lifts once the pointer leaves the item; the line is an
    underline in the accent along the row above the place, into a node
    or onto a target the accent reversed; halves and thirds by
    SGR-Pixels where the terminal has them (kitty, hottyterm).
  - Alt with the arrows in both renditions, vectors in both. The host's
    drags are KIT-23h.
- [x] **KIT-SLIDE** — Sliders: a Slider that fills from its end and
  `HottyRangeSlider`, story `hotty/slider` (journal 2026-10-10.7,
  awaiting the maintainer's look).
  - `io_neuroplast_hotty.fill` `"end"` swaps a Slider's sides,
    `⎯⎯⎯⎯■━━━━`, the rail's glyphs unchanged; on a host the fill.
  - A range: `start` and `end` bound each to a path, `min`, `max`,
    `steps` (the basic Slider's name), `checks`, `disabled`; cells
    `⎯⎯■━━━━■⎯⎯ 20–70`, the range and the focused knob in the accent,
    its number bold. Text `Price: 20–70`.
  - Each knob a Tab stop that stops where it meets the other; a click or
    a drag moves the nearer knob, and where they meet the first move
    picks.
  - On a host two `role=slider` buttons on one track; a drag by §9.1
    steps, taps by the notches. A tap after a drag let go off a slider
    is a tap again (a Slider's too).
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
