# The HOTTY profile of A2UI

    Status:   Draft, proved by this repository (gov NEIO-11).
    A2UI:     v1.0 (candidate), pinned at third_party/a2ui/REV.
    HOTTY:    SPEC.md 0.1, unchanged.

A2UI describes a UI as a flat list of components and a data model. This
profile says how a renderer shows an A2UI surface in a terminal, in the three
renditions of HOTTY's SDK.md §2.5, and what it adds to A2UI to do so. It adds
only through A2UI's own extension points: a catalog of its own (the hotty
catalog), mixed with the basic catalog component by component; renderer
functions; and one extension key prefix, `io_neuroplast_hotty`.

Nothing here changes A2UI or HOTTY. Where this draft and either of them
disagree, they are right.

## 1. Renditions

One A2UI surface has three renditions, and a renderer chooses one as SDK.md
§2.5 does:

- **Surfaces**, on a HOTTY host: the surface is one HOTTY surface, an HTML
  document, changed by deltas (§2).
- **Cells**, on a terminal that is not a host: the renderer lays the surface
  out in cells and paints it, identically in every implementation (§3).
- **Text**, with no terminal: what the surface says, as plain text (§4).

The surface's state is the renderer's, not the rendition's: the data model,
which tab is selected, whether a Modal is open, which component has the
keyboard, the text being edited. A renderer that changes rendition while a
surface is shown keeps it.

## 2. Surfaces

*To be written as the renderer settles: the element each component becomes,
ids, the kit's stylesheet as a resource, how updateComponents and
updateDataModel become deltas, which events become which A2UI writes and
actions, and Modal as a second surface above the first.*

## 3. Cells

*To be written: the layout (a flexbox subset), the character-width table,
the glyphs and roles each component paints with, and the vectors: stories
with the grids they must produce at given sizes.*

## 4. Text

*To be written.*

## 5. Keys and focus

Keys and focus are HOTTY SPEC.md §10 in both renditions. On a host the
terminal implements them. In cells the renderer does, by the same rules:

- **Who has the keyboard.** A surface has it when the user clicks a
  component that takes focus, or something gives it: the `focus` function
  (§6.3), or `autofocus` (§6.4). A click elsewhere, Escape handled by the
  program, or `blur` gives it back.
- **Keys** go to the focused component as SPEC.md §10.2's table says. Tab
  and Shift+Tab move focus in tree order; past the last component or before
  the first, the surface loses the keyboard. Escape, and every key the
  focused component does not use, reach the program.
- **Shortcuts** (§6.1) take the keys that reach the program, for the
  surface that is active: the one that has the keyboard, or when none has
  it, the one the program says is current.

Within a surface, which component has the keyboard is invisible to a
program on a host (SPEC.md §9: `focus` and `blur` concern the whole
surface). So shortcuts apply to a whole surface, in both renditions. Whether
they should apply per component is NEIO-11's open question 4, to be decided
after the storybook.

**Inputs are current when an action runs.** On a host, a field commits its
value only when focus leaves it (SPEC.md §9, `change`). Before resolving the
context of an action that a key started while a field has the keyboard, a
renderer sends `a=blur`, waits for `change` (if the value changed) and then
`blur`, writes the value to the data model, and then resolves the context.

## 6. The hotty catalog

`https://neuroplast.io/hotty/a2ui/v1/catalog.json`, in
[catalog/hotty/catalog.json](../catalog/hotty/catalog.json).

### 6.1 Shortcut

*To be written: key syntax, `press` and `action`, printable keys while a
text control has the keyboard.*

### 6.2 Form

*To be written.*

### 6.3 focus and blur

*To be written.*

### 6.4 autofocus

*To be written.*

## 7. Fallbacks

A component never fails the surface it is in.

*To be written: Video and AudioPlayer as a labelled link, Image without a
network, an unknown component, a missing child.*
