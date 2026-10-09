# Board

Last updated: 2026-10-09

**Active phase:** 1 — Cells parity ([roadmap](../roadmap.md))
**Waiting on the maintainer:** [Q-0001](../questions/Q-0001-catalog-naming.md),
the naming rule. It blocks KIT-00 and the names of the phase-1 components.
KIT-REF and KIT-LOOK don't wait for it.

## Legend

- `KIT-NN` is a gap ([gap-analysis](../knowledge/gap-analysis.md)). The
  number is its priority.
- `c` is its cells leg (phase 1), and `h` its HOTTY leg (phase 2).
- The sections below list the legs in the order they're built. The first
  open leg that nothing blocks is next.

## Phase 1 — Cells parity

- [ ] **KIT-00** — Names per Q-0001. Rename the existing components and
  functions to the rule, and write the rule into profile §6 and the
  catalog's `instructions`. Blocked: Q-0001.
- [ ] **KIT-REF** — Reference shots.
  - A `ref/` directory, as its own Go module so bubbles, huh and lipgloss
    stay out of the kit's go.mod. It holds one small program per gap, with
    the same content as the gap's story.
  - A script that shoots each reference and the kit's story side by side in
    tmux at 80×24.
  - OpenTUI references only where Bubble Tea has no equivalent: Code, Diff,
    QR and big text.
- [ ] **KIT-LOOK** — The existing components' cells look, held against huh
  and bubbles.
  - What to compare: focus marker, prompt, placeholder, caret, selection and
    error colours, borders, Tabs, the select's list and the help line.
  - It sets the shared style tokens the new components use.
- [ ] **KIT-02c** — Progress: a bar in eighths, a percentage, a gradient, and
  an indeterminate segment.
- [ ] **KIT-03c** — Spinner: bubbles' frame sets on the renderer's tick, and
  A2UI's pending state for async function values.
- [ ] **KIT-01c** — Table: header, column widths, alignment, a highlighted
  row, a sticky header while the body scrolls, and `…` cuts. Data-driven
  (L3, L5), with select-then-act (L2).
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

- [ ] **KIT-02h** — Progress: a smooth bar.
- [ ] **KIT-03h** — Spinner: frames or a CSS animation.
- [ ] **KIT-01h** — Table: a real table, a sticky header, hover, sorting by
  header click.
- [ ] **KIT-04h** — Rich list: two-line rows in proportional type, hover.
- [ ] **KIT-08h** — Key hints as keycaps.
- [ ] **KIT-07h** — Scroll view on the host's scrolling (SPEC §8).
- [ ] **KIT-05h** — Code: highlight spans, selectable text, a copy button.
- [ ] **KIT-06h** — Diff: split view in columns, acting on hunks.
- [ ] **KIT-09h** — Tree: disclosure triangles, guides.
- [ ] **KIT-10h** — Chart as SVG.
- [ ] **KIT-11h** — Suggestions in a popover surface.
- [ ] **KIT-12h** — Selection and copy: the host's own.
- [ ] **KIT-13h** — Toast as a higher-Z surface, tooltip as a popover
  surface.
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
- [ ] **KIT-FIT** — Re-check `html.Rendition.SetFit` (`k-fit`) on hottyterm
  26.10.09. That build ships hotty-blitz 050d4ba with the `100vh` fit fix.
  Keep `k-fit` unless it is now dead weight.

## Parked

- **KIT-21** — 3D and audio. These are not for a terminal UI kit. Media is
  gov R-2's research.

## Done

(none yet)
