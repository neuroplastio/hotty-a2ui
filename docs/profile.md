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

In cells the renderer lays the surface out on a grid a given number of
columns wide and as many rows tall as it takes (HOTTY's `r=auto`), and
paints every cell. The rules here make that grid the same in every
implementation. The reference is `rendition/cells`, and the vectors (§3.9)
are the test.

### 3.1 Widths

- **A cell holds one grapheme cluster** (UAX #29, Unicode 17.0.0). A
  cluster of width 2 takes two cells, the second empty.
- **A code point's width** comes from the Unicode Character Database
  17.0.0. The first rule that matches wins:
  - 0 for General_Category Mn, Mc, Me and Cf: marks, ZWJ and the other
    format controls, and the variation selectors;
  - 2 for East_Asian_Width W or F, and for Extended_Pictographic code
    points with Emoji_Presentation;
  - 1 for everything else, East Asian Ambiguous included.
- **A cluster's width** is its first non-zero code point's, with three
  exceptions that make it 2:
  - it holds VS16 (U+FE0F);
  - it is an emoji ZWJ sequence: it holds U+200D and starts with an
    Extended_Pictographic code point;
  - it is a flag: two regional indicators.

  A cluster of width 0 is not drawn.
- **Controls.** A tab is a space. Other C0 and C1 controls are dropped.
- **The table** is `rendition/cells/widths_gen.go`. `go generate` writes it
  from EastAsianWidth.txt, emoji/emoji-data.txt and
  extracted/DerivedGeneralCategory.txt at 17.0.0, and the generator pins
  each file's SHA-256. The grapheme segmenter (clipperhouse/uax29 v2.7.0)
  is at the same version.
- **Wide clusters at an edge.** A wide cluster that would cross the right
  edge of its box is not drawn, and neither is the rest of that line. At
  the frame's own edge it becomes a space. Painting over half of a wide
  cluster blanks the other half.

### 3.2 Lines

- **Wrapping** applies to Text and to errors. Lines break:
  - at hard breaks;
  - at spaces (U+0020 only).

  A word longer than the line starts a line of its own and breaks at the
  cluster where that line ends. The spaces at a soft break are dropped; a
  hard line's leading spaces stay. Trailing spaces are not drawn.
- **Cutting.** What takes one row and does not fit is cut instead, and ends
  in `…`. That covers a control's label or value, an Image's text and a
  Placeholder.

### 3.3 Layout

The layout is a flexbox subset. Every element has three measures:

- its **natural width**: its width when nothing wraps (max-content);
- its **minimum**: the narrowest it gets before its words break
  (min-content);
- its **height at a width**.

Natural widths:

| element | natural width |
| --- | --- |
| Text | its widest line unwrapped, a list item's indent and marker included; a rule counts 0 |
| Button | the label + 4 (`[ ` and ` ]`); borderless, the label |
| TextField, DateTime | the label's width, at least 20 (an HTML input's size) |
| CheckBox | 3 for the box, plus 1 and the label when it has one |
| a select | `label: `, the widest option's label, ` ▸` |
| a Choice's options | the label's width, or all the options in one row two columns apart, whichever is wider |
| Slider | the label + 1 if it has one, a track of 10, 1, and the value's width (the widest of min, max and the value) |
| Image, Icon, Media, Placeholder | what they paint (§3.4) |
| Divider | 1 |
| Card | its content + 4 |
| Row | its children's, plus one column between each two |
| Column, Form, Modal | its widest child's |
| Tabs | its titles two columns apart, or its content, whichever is wider |

Minimums:

- A Text's minimum is its longest word. A list item's indent and marker
  count with its first word.
- A field's is its label's longest word, and at least 1.
- A Slider's is 4 plus its value's width.
- A Choice's options' is its widest option, or its label's longest word if
  that is wider.
- Any other control's is its natural width.
- For containers, a Row adds its children's minimums and the columns
  between them; a Column, a Form, a Modal and a Tabs take their widest
  child's; a Card takes its content's plus 4.

How the containers lay their children out:

- **Column** covers a Column, a vertical List, a Form, a Modal's trigger,
  and the content of a Card or a tab.
  - Its children are stacked with no rows between them.
  - Across: with `align` stretch, a child gets the full width. With start,
    center or end, it gets its natural width (at most the full width),
    placed at the start, the centre (rounded down) or the end.
  - A Column has spare rows only when it is given more than it needs (a
    Row's stretch). Weighted children take them by weight: each takes the
    floor of its share, then the weighted children take one row each, in
    order, until none is left. Without weights, `justify` places them:
    - center puts half the spare (rounded down) before the first child;
    - end puts all of it there;
    - spaceBetween, spaceAround and spaceEvenly split the spare into equal
      parts, the earlier parts one larger;
    - stretch weighs every child 1.
- **Row** covers a Row and a horizontal List. One column separates each two
  children.
  - **Widths, with `justify` stretch:** the columns are shared out by
    weight, each unweighted child weighing 1, whatever the children hold
    (CSS `flex: 1 1 0`).
  - **Widths, otherwise:** each child starts at its natural width (at most
    the room).
    - While the children are too wide, one column comes off the widest
      child that is still wider than its minimum (the first of equals).
      Once none is, columns come off the widest child.
    - Spare columns go to the weighted children by weight, as in a Column,
      since a weight is `flex-grow`. Without weights, `justify` places the
      children as in a Column.
  - **Height:** the Row is as tall as its tallest child.
  - **Placement down** follows `align`: start, center (rounded down), end
    or stretch. With stretch, a child's box is the Row's height, which
    makes a Card's border, a vertical Divider and a Column's spare rows
    reach down.
  - A child left with no columns is not drawn.
- **Card** is a rounded box (`╭─╮`, `│`, `╰─╯`) in `border`. Its content
  sits inside with one column of padding on each side and no padding rows.
- **Divider** is `─` across its box, or `│` down it, in `border`.
- **Tabs** has three parts, top to bottom:
  - a bar of its titles, two columns apart, wrapping;
  - a `─` rule across, in `border`;
  - the shown tab's content.
- **accessibility.hidden** elements take no room.

### 3.4 Paint

| element | cells |
| --- | --- |
| Text | the Markdown's blocks (`view.Markdown`), one after another with no blank lines; see below |
| Image | `[image: alt]`, or `[image]` with no description, in `muted` |
| Icon | its glyph (`view.IconGlyph`, width 1; `◇` for an unknown name) |
| Video, AudioPlayer | `▶ Video`, or `▶ ` and its description, underlined and linked to its URL (OSC 8) |
| Divider, Card, Tabs | §3.3 |
| Button | `[ label ]` on one row. The label is the plain text of the Button's Texts and the glyphs of its Icons, a space apart. Primary is bold; borderless drops the brackets and is underlined; disabled is `muted` and faint. |
| TextField, DateTime | a label line in `muted`, then the value on a row underlined across the box (§3.5), then the error |
| CheckBox | `[x] label` or `[ ] label`, then the error |
| a select (one value, `checkbox` display) | `label: value ▸`: the label in `muted`, then the picked option's label, or `…` in `muted` when none is picked. While its list is open (§3.7), the options follow one a row: `  ● label` for the picked one and `  ○ label` for the others. Then the error. |
| a Choice's options (several values, or `chips`) | a label line in `muted`, then the options two columns apart, wrapping: chips `( label )` and `(● label)` when picked, else boxes `[ ] label` and `[x] label`; then the error |
| Slider | `label ━━━━●──── 50` on one row: the label in `muted` and a space, then the track, a space and the value. The track fills the columns left. Up to the knob it is `━`, the knob is `●` at round((value − min) / (max − min) × (track − 1)), and after it the track is `─` in `border`. When there is no label, or the track would be shorter than 3, the label is dropped. Then the error. |
| Placeholder | `…` in `muted` while pending; `! Type` in `warning` when the type is unknown or the component contains itself |
| an error | `✗ message` in `error` under its control, wrapped, its continuation lines indented 2 |

**Text blocks:**
- A heading is bold, and an h1 is also underlined.
- A list item is indented two columns a level. Its marker is `• `, `1. `,
  or `☐ ` / `✓ ` for a task. Continuation lines align with the text.
- A quote starts every line with `▎ `, once per level, in `border`.
- A code block is `muted`, its lines broken by cluster at the width rather
  than wrapped.
- A rule is `─` across, in `border`.
- Inline: bold, italic and strikethrough are attributes, and code is
  `muted`. A link is underlined and its cells carry the URL (OSC 8).
- A caption is `muted` throughout.

**Focus** (the element with the keyboard). Each focused element is in
`accent`, with an attribute so that the focus still shows without colour:
- A Button, a chip, a Tabs' title, a select's value, a Media link and a
  Modal trigger that is not a control are reversed whole.
- A CheckBox and a box option reverse their box.
- A Slider's track is in `accent` up to the knob, and the knob is
  reversed.
- A text field's label line is in `accent`, and its cursor cell is
  reversed.

### 3.5 Text fields

- **The value row.** A text field's value sits on a row underlined across
  its box. The placeholder shows while the value is empty, in `muted` and
  faint. A DateTime's placeholder is the form its value takes:
  `YYYY-MM-DD`, `HH:MM` or `YYYY-MM-DDTHH:MM`.
- **obscured** shows `•` for each cluster.
- **longText** shows its hard lines, at least 3 rows and at most 8, with no
  soft wrap.
- **Without the keyboard** a field shows its value from the start, each
  line cut with `…`.
- **With the keyboard** a field shows the cursor:
  - The cursor is an index into the value's clusters; it starts at the
    end.
  - The cell under it is reversed, and `Frame.Cursor` reports it.
  - The field scrolls as little as keeps the cursor in it: across, by
    columns (it shows from the start whenever the cursor's line fits),
    and for a longText, down by lines.

### 3.6 Colour and attributes

Cells name NEIO-4's roles (`fg`, `bg`, `muted`, `accent`, `selection`,
`surface`, `border`, `success`, `warning`, `error`, `info`), never colours.
`accent` goes only on focus. Every state that has a colour also has a glyph
or an attribute: `✗`, `!` and `…`, reverse for focus, faint for disabled
and for placeholders.

At the ANSI-16 floor, a role is a foreground colour:

| role | SGR |
| --- | --- |
| `fg`, `bg`, `surface` | the terminal's (none) |
| `muted`, `border` | 90 (bright black, the host's `ansi-8`) |
| `accent`, `selection` | 94 (bright blue, the host's dark-scheme accent, SPEC §8) |
| `success` | 32 |
| `warning` | 33 |
| `error` | 31 |
| `info` | 36 |

The attributes are bold (1), faint (2), italic (3), underline (4), reverse
(7) and strikethrough (9). Under NO_COLOR, the attributes are written and
the role colours are not, and every meaning above still reads.

In the output:
- Each run of cells in one style starts with `ESC [ 0 ; … m`, or `ESC [ 0
  m` back to plain.
- A link's cells are wrapped in OSC 8 (`ESC ] 8 ; ; URL ESC \`).
- Every row ends reset, without its trailing plain blank cells.

### 3.7 Modal, keys and clicks

**An open Modal's content** is a floating panel:
- It is a rounded box in `border`, filled blank, with one column of
  padding.
- Its width is the content's natural width + 4, at most the columns
  less 4 (all the columns below 20). Its height is the content's height
  + 2.
- It is centred, rounded down. The frame grows to the panel's height if
  the surface is shorter.
- The surface under it is faint and takes no clicks.

**Keys** are §5's, with these rules for cells:

- **Who takes a key.** While the surface has the keyboard, the focused
  element takes SPEC §10.2's keys for its kind, unmodified or with Shift
  only.
- **Text fields** take printable characters, Space, Backspace, Delete,
  ArrowLeft, ArrowRight, Home and End.
  - A number field takes every printable character but keeps only
    `0-9 . , - + e E`.
  - Enter submits the field's Form, if it is in one.
  - A longText also takes ArrowUp, ArrowDown, PageUp and PageDown, which
    move by a line and by the rows it shows. Enter inserts a line break
    there.
  - Every edit writes the value at once.
- **A select** works in two states:
  - Closed, ArrowUp and ArrowDown pick the previous and next option.
    Home and PageUp pick the first, End and PageDown the last. A printable
    character picks the next option whose label starts with it (case
    aside), wrapping.
  - Space or Enter opens its list, with the value highlighted. The same
    keys then move the highlight, and Space or Enter picks it and closes
    the list.
  - The list closes when the select loses the keyboard.
- **A Slider** takes no keys: SPEC §10.2 has no row for a range. A click
  sets it.
- **Space and Enter** activate a Button, a Tabs' title, an option, a
  CheckBox, a Media link, or a Modal trigger that is not a control.
- **Tab and Shift+Tab** move the keyboard in tree order.
- **Any other key**, Escape among them, goes to the surface's Shortcuts.
  - Without the keyboard, only the Shortcuts are tried.
  - A printable key, alone or with Shift, never reaches a Shortcut while a
    text field has the keyboard.
- **Escape** that no Shortcut takes closes an open Modal, as on a host.

**A click** lands on the topmost thing drawn at its cell:
- **Something that takes focus** gets the keyboard, and then:
  - a field puts its cursor before the cluster clicked (a click on its
    label line only focuses it);
  - a select opens or closes its list, and a click on a row of the open
    list picks that option and closes the list;
  - a click on a Slider's track sets the value at that column:
    min + (max − min) × column / (track − 1), stepped and clamped;
  - anything else is activated.
- **A disabled Button** does not take the keyboard, but the click still
  activates it, as in rendition/html; the controller decides what that
  does.
- **A click on nothing that takes focus** gives the keyboard back.
- **A click outside an open Modal's panel** closes the Modal.

### 3.8 What cells keep

The cursor, the scroll offsets and which select's list is open belong to
the rendition, by element id. Everything else is the surface's state (§1).

### 3.9 Vectors

`rendition/cells/testdata/<example>.<columns>.golden` are the grids, as
plain text (rows joined by newlines, trailing spaces trimmed), that A2UI's
basic examples produce at 60 and 40 columns. The examples are
`00_simple-login-form`, `00_row-layout`, `07_task-card`,
`34_child-list-template`, and `36_modal` with its Modal open. An
implementation of this profile produces the same grids. `TestWidth` in
`rendition/cells/width_test.go` lists the width vectors. They cover CJK,
combining marks, VS16 and VS15, emoji ZWJ sequences, flags (tag sequences
included), keycaps and a lone ZWJ.

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
