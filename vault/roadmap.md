# Roadmap

The goal: the kit looks nice in a plain terminal first, and awesome on a HOTTY
host. Gaps are closed in two phases. Every component goes through phase 1
before any goes through phase 2. The gaps and their tickets are in
[knowledge/gap-analysis.md](knowledge/gap-analysis.md), and the build order is
on [the board](tasks/board.md).

## Phase 1 — Cells parity

Each gap gets its cells rendition, and it should look and feel like its Bubble
Tea or OpenTUI reference: the same density, glyphs, keys and states. The
existing components are restyled first (KIT-LOOK), so the new ones share
their style.

A phase-1 leg (`KIT-NNc`) is done when it has all of these:
- **Catalog.** The component (or function) is in the hotty catalog under the
  names Q-0001 settles, with its description and a line in the catalog's
  `instructions`. Any L2, L4 or L5 rule that applies is in that line
  ([a2ui-limits](knowledge/a2ui-limits.md)).
- **Renditions.**
  - It draws in cells.
  - It draws in text.
  - On a HOTTY host, the html rendition draws a plain but correct baseline,
    so no story shows `! Type`.
- **A story** under `story/stories/hotty/` that shows its states: focused,
  disabled, empty, loading and error, where it has them.
- **Vectors** for anything it does with keys or edits, played on cells and
  html alike.
- **A reference shot**, side by side with Bubble Tea's or OpenTUI's
  equivalent at the same size (KIT-REF).
- **A section in `docs/profile.md`.**
- **Checks.** `make check` passes, the change is pushed, and the storybook
  has been relaunched and looked at.

Exit: every phase-1 leg on the board is ticked, and the maintainer has seen
each side-by-side shot.

## Phase 2 — The HOTTY layer

The same components on a HOTTY host, using what HTML gives:
- proportional type;
- SVG;
- real tables;
- popover and higher-Z surfaces;
- the host's own scrolling and selection;
- hover.

A phase-2 leg (`KIT-NNh`) is done when its html rendition has been judged in
hottyterm (26.10.09 or later, which has the §10.2 field keymaps) and in the
xterm.js addon, and hotty-demo's storybook has moved to it.

Exit: every phase-2 leg on the board is ticked, and
<https://hotty.neuroplast.io/storybook> shows them.
