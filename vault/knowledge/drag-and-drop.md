# Drag and drop: what A2UI and HOTTY give, the stories, the models

As of 2026-10-10. The maintainer asked for drag and drop to be first-class
behaviour of lists, tables, trees and the like (KIT-23). This note lays out
the design for them to pick from. Nothing is built yet.

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
  high, so the line is an underline on the row above, or a mark in the
  gutter.

**Keys (S6).**
- Alt+↑ and Alt+↓ move the selected item one place, as in VS Code and
  Org's M-up and M-down.
- In a tree, Alt+← and Alt+→ move it out of its branch or into the one
  above, as Org's M-left and M-right do.
- For a distant move, Ctrl+X marks the item and Ctrl+V puts it after the
  selection, as file managers do (and lf's and ranger's `dd` then `p`).
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

## Cost

- Local reorder in a list, a table and a templated List, P1 in both
  renditions, and the keys: no wire change.
- P2, and "into" in a tree by pointer: G1.
- S4's empty column and S5: G2. Building them without it would be the
  kind of kit-side workaround the maintainer has asked not to ship.
