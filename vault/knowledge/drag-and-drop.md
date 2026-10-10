# Drag and drop: what A2UI and HOTTY give, the stories, the models

As of 2026-10-10. The maintainer asked for drag and drop to be first-class
behaviour of lists, tables, trees and the like (KIT-23). This note lays out
the design they picked from, and what the cells leg settled (below,
"Built").

**Picked (the maintainer, 2026-10-10):** C, both (the kit moves an item
within and between its own lists, and a drop onto anything else is the
agent's), and P2, a line. Keys are Alt with the arrows; cut and paste is
not the kit's. P2 needs G1, and drop-only targets need G2, so both went
to hotty the same day. The cells leg (KIT-23c) doesn't wait for them; the
host leg (KIT-23h) does.

**A2UI v1.0 has none.** It has no drag event, no drop target and no
reorder in the basic catalog. A renderer tells the agent things through an
action (an event with its context) and, with `sendDataModel`, through the
data model that goes along with it. Two-way binding is local to the
renderer (protocol, "Two-way binding & input components"): a Slider writes
its value as it is dragged. So a drop the kit handles is a write to bound
data, an action, or both. All of it is the kit's, under
`io_neuroplast_hotty`, as the icons are.

**HOTTY has drags (SPEC §9.1).** `dragstart`, `drag` and `dragend` on an
element with `drag` in its `data-on`. A `drag` names the element under the
pointer each time it changes, and every event carries the surface's cell
and the keys held. The pointer is the surface's until the release. That
is enough to lift a row and to know which row it is over. What it lacks is
under "Gaps" below.

**Cells** reads the terminal's mouse reports with the button held, as a
Slider's drag already does (`cells.go`, `Drag` and `Release`).

## Stories

- **S1 Reorder.** I reorder a playlist or a to-do list in a HottyList,
  with a mouse on a host, a finger on a phone, or keys in a terminal.
- **S2 Tree.** I move a file into a folder, or between two files. A
  closed folder opens when I hold the file over it. A folder can't go
  into itself.
- **S3 Table.** I rank a backlog by dragging rows of a HottyTable.
- **S4 Kanban.** I move a card from one column to another. Each column
  is a basic List of cards from a template.
- **S5 Drop onto a thing.** I drop a ticket onto an assignee, or a file
  onto "Attach". The agent decides what that means.
- **S6 Keys only.** I make every move above without a mouse.
- **S7 Copy.** I copy instead of moving, with a modifier held.
- **S8 Several.** I drag the whole selection. This waits on multiple
  selection, which the kit doesn't have yet.

## Models

Where the move happens:
- **A, local.** The renderer moves the item in the bound array itself,
  as a Slider writes its value, then sends `onMove` with the item, where
  it was and where it went. The move is instant and works with no agent.
  An agent that disagrees writes the data back.
- **B, agent.** The renderer only shows the drag. On the drop it sends
  `onDrop` with the item, the target and the position, and the agent
  rewrites the data. A drop can mean anything (assign, attach, copy), at
  one round trip per drop, and nothing moves until the agent answers.
- **C, both (recommended).** List, table, tree and a templated basic List
  reorder locally (A) and send `onMove`. Any component can also be a drag
  source or a drop target, through `io_neuroplast_hotty.drag` (a type and
  a value) and `io_neuroplast_hotty.drop` (the types it accepts and an
  action). That is B, matched by type as HTML's `dataTransfer` matches
  sources to targets. S4 is two lists with the same type in one surface,
  so its move is local too.

What the user sees during the drag:
- **P1, live.** The lifted item takes the place of the row under the
  pointer as it goes, and the other rows shift, as in iOS and SortableJS.
  It needs only "which row", which §9.1 gives. It can't tell "into a
  folder" from "before it".
- **P2, a line.** The rows stay put, and a line shows where the item will
  land, between rows, or a ring shows a folder it will go into. The move
  happens on the drop, as in Finder and VS Code's explorer. It needs to
  know where in the row the pointer is (G1). In cells a row is one cell
  high: where the terminal reports the pointer in pixels (SGR-Pixels,
  mode 1016) the kit knows the half or the third of the row; elsewhere
  the item takes the place of the row it is over (see "Built": "the side
  the item comes from" read literally can't move an item one place
  down).

**Keys (S6).**
- Alt+↑ and Alt+↓ move the selected item one place, as in VS Code and
  Org's M-up and M-down.
- In a tree, Alt+← and Alt+→ move it out of its branch or into the one
  above, as Org's M-left and M-right do.
- Not cut and paste (Ctrl+X, then Ctrl+V): the maintainer ruled it out of
  the kit. So a move by keys stays within one component. A move by keys
  to another list, or onto a thing (S4, S5), is the app's, for instance a
  Button or a HottyShortcut that sends an action.
- On a host these are data-keys programs (SPEC §10.2).

**Touch.** A row in a list that scrolls has `pan-y`, so under §9.1 a
touch scrolls the list rather than dragging the row. One way is a grip
(`⠿`) at the start of the row with `touch-action: none`, which works
today. The other is a long press and then the drag, which is how phones
do it and which §9.1 doesn't have (G3).

## Gaps, for hotty once a model is picked

- **G1, where in the target.** Steps are measured only against the
  element being dragged. Sketch: when `t` has `data-steps`, a `drag` and
  a `dragend` also say where in `t` the pointer is. P2 needs it, and so
  does telling "into a folder" from "between" in a tree.
- **G2, a target that isn't draggable.** An element is named in `t` only
  when it has `drag` in its `data-on`. That also lets a press drag it, and
  stops text selection in it. Sketch: `data-on~=drop` names an element in
  `drag` and `dragend` without making it draggable. An empty list, a
  kanban column and an assignee all need it.
- **G3, a long press drags** on touch. Only needed if a grip isn't
  enough.
- **Not a gap: holding still at a scroll edge.** Drags are reported only
  when the element changes, so scrolling while the pointer is held at the
  edge needs a timer in the program.
- **G4, another surface** (later). The pointer stays the dragging
  surface's, and `t` is empty over another surface, so two A2UI surfaces
  can't drop into each other.

## Built (KIT-23c, 2026-10-10, journal 2026-10-10.8)

The cells leg and the keys on a host. The API is profile §6.21's:
- **Props** on HottyList, HottyTable and HottyTree: `reorderable`
  (boolean), `dragType` (string), `moved` (bound: where the move is
  written), `onMove` (an Action). The items must be bound to be written;
  literal ones move in the renderer's state.
- **Extensions** under `io_neuroplast_hotty`: `reorder` on a templated
  List (`type`, `moved`, `onMove`), `drag` on any component (`type`,
  `value`), `drop` on any component (`accepts`, `value`, `action`).
  Their schema is in the catalog's `$defs`, instruction 21 tells an
  agent.
- **The move** written to `moved`: `{"item", "from": {"index", "path"},
  "to": {"index", "path"}}`, a tree's places with `parent` too. `item`
  is the item's id as `selected` names it in a kit list, the item's data
  in a List. `to`'s index is after the item left.
- **A drop** writes what is dropped to `drop.value`'s path (a source's
  value, a list item's data) and runs `drop.action`; the item stays.

**Settled while building:**
- **The line without pixels.** "The side the item comes from" read
  literally (from above, the line above the row) means an item over the
  next row lands where it was, so a drag can't move it one place down.
  Instead the item takes the place of the row it is over: after it when
  it comes from above, before it from below, and into a node that takes
  children; from another list, before it. With pixels, halves and
  thirds.
- **The line's look.** An underline in the accent (SGR 4 with SGR 58)
  along the row above the place, which leaves rows where they are and
  text in its colour; on a blank row between items it is a clean rule.
  An overline (SGR 53) only at the frame's top, as it can't take the
  accent (SGR 58 colours underlines), so that row's text goes accent.
  Into a node, and onto a drop target, the accent reversed. Checked in
  vhs (xterm.js), kitty and hottyterm.
- **Pixels.** The storybook asks for mode 1016 (DECRQM) and the cell's
  size (XTWINOPS 16) and sets the mode once both answer; kitty and
  hottyterm (the AUR package, 26.10.10.r72) answer, and a real drag
  there goes by halves.
- **Grips.** None in cells. On a host, always (KIT-23h): a document
  can't tell a finger from a mouse.
- **An empty folder** is `children: []`; cells draws it `▶ name 0`.
- **The press.** What a click does past selecting waits for the
  release on what can be dragged, so a drag never runs a card's Button.
- **Innermost wins**, for the press and the drop.

**Not built** (later, or KIT-23h):
- A closed folder opening while held over it (S2's spring-loading).
- Scrolling at an edge while held: a HottyList's page, a tree with a
  `height`, a scroll view.
- Choosing a tree's level by the pointer's column after a branch's last
  child (it lands at the child's level).
- S7 copy and S8 several, as planned.
- On a host: drags, the line, the grip (G1, G2).

## Cost

- Local reorder in a list, a table and a templated List, P1 in both
  renditions, and the keys: no wire change.
- P2, and "into" in a tree by pointer: G1.
- S4's empty column and S5: G2. Building them without it would be the
  kind of kit-side workaround the maintainer has asked not to ship.
