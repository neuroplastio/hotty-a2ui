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

On a host, an A2UI surface is one HOTTY surface. The reference is
`rendition/html`.

**The document.** It holds the kit's stylesheet and two elements: the
surface (`~s`), and the layer an open Modal shows in (`~o`), empty while
none is open. The document asks the network for images over HTTPS only
(`<meta name="hotty-network" content="img-src https:">`); the host's own
policy decides (SPEC.md §7.2).

**The look.** A surface is a web page in the terminal, so it is set in web
type, not in the grid: a proportional face (Inter, else the system's
sans-serif), a type scale for Text's variants (`h1` to `h5`, `caption`,
`body`), line height 1.5, and spacing in `rem`, the host's root size,
which follows the terminal's font. Cells (§3) lay out in the grid; a
surface lays out as a page, the same components in the same order.
- **Colour** is the host's palette (SPEC.md §8) unless a theme sets it on
  the surface's two elements as the kit's variables (`--k-fg`, `--k-bg`,
  `--k-accent`, …). The colours made from those (a card's raised fill, a
  button's tonal fill, rules) follow. Without a theme the kit keeps to
  NEIO-4: accent only on focus and links, and a primary Button filled
  with the foreground.
- **Shape**: Cards, fields and Buttons are rounded, chips are pills; a
  theme may set its own radii (Material: pill Buttons, 12px Cards, 4px
  fields).
- **Fills before borders.** A Card, a field and a chip each have a fill,
  so they keep their shape where a thin rounded border does not draw
  (Blitz at a fractional scale).
- **Emoji** go in a span that names a colour emoji font, in a Text's HTML
  and in labels. A host's font fallback may otherwise draw one from a text
  font, in outline. A pictograph that defaults to text (☀ without VS16)
  stays text.

**Elements.** Each component is one element, whose `id` is the
component's node key encoded for HOTTY. A node key is the component's id,
with a template item's path in brackets (`row[/items/0]`). Letters,
digits and `-_./` stay; every other byte is `~` and two upper-case hex
digits, so an id is always a valid control value (SPEC.md §3.2). The
parts a component draws besides itself add `~` and a lower-case letter:
its wrapper `~w`, label `~l`, error `~e`, a Slider's notches `~k0`,
`~k1`, …. A control comes wrapped with its
label and its error, which is always there (empty while there is none),
so that an error comes and goes as a text delta.

| component | element |
| --- | --- |
| Row, Column, List | `div`, a flex row or column; List scrolls. A Row wraps, as a page's inline content does, when it holds a field or only inline things (texts, icons, buttons, small pictures); a Row of columns or cards shrinks them instead. A weighted child of a Row takes its share of it, whatever it holds (`flex: weight`, as A2UI's Lit renderer has it), so Rows weighted alike line up as a grid's; in a Column, a weight is a share of the rows to spare |
| Card | `div`, a raised fill and a rounded border |
| Text | `div` with the Markdown as HTML, its emoji in spans; links open in the terminal (`target=_blank`, SPEC.md §9). A word breaks only when it fills a line alone. A Text that is one number (`$850,000,000.00`, a grid's cell) may also break after its group separators, a zero-width space after each, so a narrow column breaks it between groups, not between any two digits. An ordered list item's number is written out, a `span.k-n` the sheet sets in the indent with a gap, ending at the same place for `9.` and `10.`, because Blitz draws a list's own numbers flush against the text and has no counters to draw them otherwise |
| Image | `img`, sized by its variant in `rem` |
| Icon | `span` with a colour emoji where one says the name plainly, else the glyph cells draw |
| Video, AudioPlayer | `a` with `href`: a link the program opens (§7) |
| Divider | `hr`, or a vertical rule |
| Button | `button type=button`, `disabled` while its checks fail. A vertical List's Button is its row (`k-item`), as a menu's: as wide as the List, its label at the start, its variant only a fill, so that focusing or picking a row moves nothing |
| TextField | `input` (`text`, `password`, `number`) or `textarea`, with `data-on=input` |
| CheckBox | `input type=checkbox` |
| ChoicePicker | one of several shown as checkboxes is a select: a `button aria-haspopup=listbox` with the picked option's label and a caret, whose list opens in a surface of its own (below); else its options, as checkboxes (several) or chips (`button aria-pressed`). Hosts draw `select` unevenly (Blitz not at all), and a select's list is the program's to place (SPEC.md §9) |
| Slider | `button role=slider` drawing the track (its rail, fill and knob), between `−` and `+` buttons out of the Tab order, then an `output` as wide as the widest value (§3.3's), in `ch`. The track is cut into notches with `data-on=drag`, one a step and one each end (at most 41; twenty steps without a `step`), so a drag sets the value of the notch under the pointer. Hosts draw `input type=range` unevenly (Blitz not at all) |
| DateTimeInput | `input type=text` with the ISO 8601 value, its form as the placeholder, as in cells (§3.5). Hosts draw date and time inputs unevenly (Blitz not at all), and none takes an offset such as `Z` |
| Tabs | a `tablist` of `button role=tab`, then the tab shown |
| Modal | its trigger; while open, its content in the layer, over a backdrop, the surface `inert` |
| HottyForm | `form` with a hidden submit button out of the Tab order, so Enter submits |

**Updates are deltas.** The document goes once. After it, the renderer
diffs the elements it sent against the elements the view makes now, and
sends the smallest deltas the ids allow (SPEC.md §6): `attr` and `unattr`
for an attribute, `text` for an element whose content is one text, a
recursion into children that keep their ids and order, and `inner`
(which morphs) for any other change below an element. `updateComponents`
and `updateDataModel` change a surface this way, and so does the
renderer's own state: a tab shown, a Modal opened.

**Events.** Each acts on the surface as the user's act does in cells:

| event | does |
| --- | --- |
| `input`, `change` | writes the control's value: to its bound path, else as the renderer's |
| `click` | a Button runs its action; a tab is shown; a chip toggles; a link opens; a Modal's trigger opens it; the backdrop closes it; a Slider's `−` or `+` steps it; a select opens or closes its list, an option in the list is picked, and a click beside the list closes it |
| `submit` | the HottyForm submits (§6.2) |
| `dragstart`, `drag` | a Slider takes the value of the notch the pointer is on (SPEC.md §9.1); `dragend` and a drag off the track leave it |
| `focus`, `blur` | the surface has the keyboard, or not |

Typing reaches the data model at every key (`data-on=input`), as A2UI's
own renderers write a bound field. The renderer does not send the value back
while the user edits the field: the host has it, and an echo arrives a key
late. Once the field is left, the view's value goes out, and the
program's wins (SPEC.md §6.2).

**A select's list** is the rendition's, as cells' is (§3.8). It is a
surface of its own, `<surface>-list`, which the program places above the
others: under the select by the `area` the click reports, as wide as the
select at least, and as high as the host's fit says, within the program's
screen (above the select when there is more room there). Inside the
select's surface, the list would be cut at its edges, and a surface grown
to hold it would hide what it covers. In the select's surface, the layer
takes a click anywhere else, which closes the list; so does the keyboard
going to another surface. Its options are `button role=option`, out of
the Tab order, the picked one filled. While it is open the program
has the keyboard: the renderer sends `a=blur` and the controller keeps
the select focused, since a host scrolls with the arrows a focused button
leaves (SPEC.md §5.3), and they would never reach the program. The keys
then work the list as in cells (§3.7), its highlight drawn as focus is;
Tab closes it and goes on. When it closes, the select has the keyboard
again (`a=focus`). Closed, a select is a focused button, which leaves the
arrows, Home, End, Page Up, Page Down and letters to the program, which
picks with them as cells does; a host that scrolls takes the arrows for
that first.

**The keyboard.** The renderer gives the host the keyboard it has in mind
when they differ: `a=focus` at the element `autofocus` (§6.4) or `hottyFocus`
(§6.3) names, `a=blur` for `hottyBlur`. Within the surface, Tab moves focus
where the program does not see it (SPEC.md §9), so the renderer knows the
element last clicked or edited, not always the one focused.

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
| Button | the label + 4 (`[ ` and ` ]`); borderless, the label; a List's, the label + 2 |
| TextField, DateTime | 2 for the gutter, plus the label's width or 22 (the prompt and an HTML input's size of 20), whichever is wider; a longText, the label's or 20 |
| CheckBox | 2 for the gutter, 3 for the box, plus 1 and the label when it has one |
| a select | 2 for the gutter, plus the label's width or the widest option's label and ` ▾`, whichever is wider |
| a Choice's options | 2 for the gutter, plus the label's width or the widest option row (`> [ ] label`), whichever is wider; chips, all the options in one row two columns apart |
| Slider | 2 for the gutter, the label + 1 if it has one, a track of 10, 1, and the value's width: the widest of min, max, the value, and a value on a step (its ends' whole part, a point and the step's decimals), so that the track keeps its length as the value moves |
| Image, Icon, Media, Placeholder | what they paint (§3.4) |
| Divider | 1 |
| Card | its content + 4 |
| Row | its children's, plus one column between each two |
| Column, HottyForm, Modal | its widest child's |
| Tabs | its titles two columns apart, or its content, whichever is wider |

Minimums:

- A Text's minimum is its longest word. A list item's indent and marker
  count with its first word.
- A field's is the gutter and its label's longest word, and at least 3.
- A Slider's is the gutter, 4 and its value's width.
- A Choice's options' is the gutter and its widest option, or its label's
  longest word if that is wider.
- Any other control's is its natural width.
- For containers, a Row adds its children's minimums and the columns
  between them; a Column, a HottyForm, a Modal and a Tabs take their widest
  child's; a Card takes its content's plus 4.

How the containers lay their children out:

- **Column** covers a Column, a vertical List, a HottyForm, a Modal's trigger,
  and the content of a Card or a tab.
  - Its children are stacked with no rows between them, except for one
    blank row between two controls (fields and Buttons that are not a
    List's rows) when either has a title row (§3.4), or when one is a
    field and the other a Button: huh's space between its fields, and
    bubbles' before a form's button. A stack of CheckBoxes, or of
    Buttons, stays tight. Through a Row, a Column or a HottyForm, the
    rule sees its first child (or, before it, its last).
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
| Button | `[ label ]` on one row. The label is the plain text of the Button's Texts and the glyphs of its Icons, a space apart. Primary is bold; borderless drops the brackets and is underlined; disabled is `muted` and faint. A vertical List's Button is its row: ` label `, its style (focus's reverse, say) across the List; bold unless borderless, and never underlined. |
| TextField, DateTime | in the field's box past the gutter: a title row, then the value (§3.5), then the error |
| CheckBox | past the gutter, `[•] label` or `[ ] label`, then the error |
| a select (one value, `checkbox` display) | past the gutter, a title row, then the picked option's label, or `…` in `muted` when none is picked, and ` ▾` in `muted`. While its list is open (§3.7), the options follow one a row: `● label` for the picked one and `○ label` for the others, each after `  `, or after `> ` in `accent` on the highlighted row, whose label is `accent` and bold. Then the error. |
| a Choice's options (several values, or `chips`) | past the gutter, a title row, then the options: one a row, `[•] label` or `[ ] label` after `  `, or after `> ` in `accent` for the option with the keyboard (huh's multiselect); chips two columns apart, wrapping, `( label )` and `(● label)` when picked. Then the error. |
| Slider | past the gutter, `label ━━━━●──── 50` on one row: the label as a title and a space, then the track, a space and the value. The track fills the columns left. Up to the knob it is `━`, the knob is `●` at round((value − min) / (max − min) × (track − 1)), and after it the track is `─` in `border`. When there is no label, or the track would be shorter than 3, the label is dropped. Then the error. |
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

**Fields** are the text fields, DateTime, CheckBox, Choice and Slider:
the controls huh calls fields, drawn as huh draws them.
- **The gutter.** A field's first two columns are its gutter. While the
  field has the keyboard (for a Choice's options, one of them), the
  gutter is `┃ ` in `accent` down the field's rows; otherwise it is blank.
- **The title.** A text field, a DateTime or a Choice with a label has a
  title row: the label, bold, in `accent` while the field has the
  keyboard. A CheckBox's and a Slider's labels are on their one row.

**Focus** (the element with the keyboard). Each focused element is in
`accent`, with a glyph or an attribute so that the focus still shows
without colour:
- A field shows its gutter's bar, and its title is in `accent`.
- A Button, a chip, a Tabs' title, a Media link and a Modal trigger that
  is not a control are reversed whole.
- A CheckBox's box is in `accent`; a box option's row starts with `> `.
- A Slider's track is in `accent` up to the knob, and the knob is
  reversed.
- A text field's prompt is in `accent`, and its cursor cell is reversed.

### 3.5 Text fields

- **The value row.** A one-line field's value follows a prompt, `> `, in
  `muted` (`accent` while it has the keyboard), as bubbles' text input
  has it. A longText has no prompt: its rows carry `┃` in the gutter, in
  `border`, as bubbles' textarea does. The placeholder shows while the
  value is empty, in `muted` and faint. A DateTime's placeholder is the
  form its value takes: `YYYY-MM-DD`, `HH:MM` or `YYYY-MM-DDTHH:MM`.
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

In a theme (`rendition/theme`), a role the theme colours is that colour
in truecolor (`38;2;r;g;b`), and the theme's `bg` is under every cell
(`48;2;r;g;b`), so a row keeps its blank cells to the end. A role the
theme leaves to the terminal keeps the floor above.

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

- **Who takes a key.** While the surface has the keyboard, a focused text
  field takes the keys its keymap binds and the characters it types (§5,
  SPEC §10.2); another focused element takes SPEC §10.2's keys for its
  kind, unmodified or with Shift only.
- **Text fields** edit as a host's do, with hotty-go's `hottyedit`: the
  same actions, words, lines and caret, so the same keys make the same
  edits in both renditions. A field in cells does not wrap: its rows are
  its lines, and a longText's page is the rows it shows.
  - A number field takes every printable character but keeps only
    `0-9 . , - + e E`.
  - `submit` (Enter, in a single-line field) submits the field's
    HottyForm, if it is in one. In a longText, Enter types a line break.
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
- **A Slider** steps by Left and Down (back) and Right and Up, and goes
  to its ends by Home and End. A step is its `step`, else a twentieth of
  its range, clamped. Whatever sets it, a Slider's value lands on a step
  from its min (a hundredth of its range without a `step`), rounded to
  the step's decimals: 0.45 less 0.05 is 0.4, not 0.39999999999999997.
  On a host the slider is a button, which leaves these keys to the
  program (SPEC.md §10.2), so the renderer steps it there too.
- **Space and Enter** activate a Button, a Tabs' title, an option, a
  CheckBox, a Media link, or a Modal trigger that is not a control.
- **Tab and Shift+Tab** move the keyboard in tree order.
- **Any other key**, Escape among them, goes to the surface's
  HottyShortcuts.
  - Without the keyboard, only the HottyShortcuts are tried.
  - A key a focused text field uses, a character it types or a key its
    keymap binds to an edit, never reaches a HottyShortcut.
- **Escape** that no HottyShortcut takes closes an open Modal, as on a host.

**A click** lands on the topmost thing drawn at its cell:
- **Something that takes focus** gets the keyboard, and then:
  - a field puts its cursor before the cluster clicked (a click on its
    label line only focuses it);
  - a select opens or closes its list, and a click on a row of the open
    list picks that option and closes the list;
  - a click on a Slider's track sets the value at that column:
    min + (max − min) × column / (track − 1), stepped and clamped, and
    the value follows the pointer while the button stays down;
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

With no terminal, a surface is plain text: what it says and holds, in tree
order, one line per component, for a pipe or a screen reader's log. The
reference is `rendition/text`.

- A Row's parts share one line, two spaces apart, when each is one line.
- Text is its Markdown as plain text: paragraphs and list items a line
  each, markers kept.
- A field is `Label: value`; an obscured one shows `•` for each
  character, and an empty one its placeholder in parentheses.
- A CheckBox is `[x] Label` or `[ ] Label`; a ChoicePicker `Label: ` and
  the labels picked; a Slider `Label: value (min–max)`.
- A Button is `[ label ]`, followed by `(disabled)` while its checks fail.
- Tabs are their titles, the one shown in brackets, then its content.
- An error is `✗ message`, on the line after its control.
- Image, Icon, Video, AudioPlayer and placeholders are as §7 has them.
- An open Modal's content follows the surface, after a line `───`.
- A component with `accessibility.hidden` says nothing.

## 5. Keys and focus

Keys and focus are HOTTY SPEC.md §10 in both renditions. On a host the
terminal implements them. In cells the renderer does, by the same rules:

- **Who has the keyboard.** A surface has it when the user clicks a
  component that takes focus, or something gives it: the `hottyFocus`
  function (§6.3), or `autofocus` (§6.4). A click elsewhere, Escape handled
  by the program, or `hottyBlur` gives it back.
- **Keys** go to the focused component as SPEC.md §10.2's table says. Tab
  and Shift+Tab move focus in tree order; past the last component or before
  the first, the surface loses the keyboard. Escape, and every key the
  focused component does not use, reach the program. Keys are named as
  SPEC.md §10.4 has them: `Control+s`, `Alt+b`, `A`, `Space`.
- **A text field's keys are its keymap** (SPEC.md §10.2). The renderer
  sets the surface's on its top elements as `data-keys`: hotty-go's
  `TerminalKeys`, the keys of Bubble Tea's text input (Control+a and
  Control+e, Alt+b and Alt+f, Control+w, Control+k, Control+u…), unless
  the program sets another (`SetKeys`; empty for SPEC.md's default
  keymap alone). A component's `keys` (§6.5) overrides it key by key for
  the fields inside. The cells rendition resolves the same keymap
  (`hotty.Resolve`), so a key does the same in both.
- **HottyShortcuts** (§6.1) take the keys that reach the program, for the
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
It then gives the keyboard back with `a=focus` and no `t`: to the element
the host had focused, wherever Tab took it. A renderer whose fields report
`input` is current anyway; the blur keeps it so for a field that does not.

**Vectors.** `vectors/keys.yaml` checks all of this against both
renditions: a story played by keys, focus, blur and clicks, with who has
the keyboard, the actions the agent got and the data model expected after
each step. The host side runs on hottytest's host, whose `Key` takes keys
as SPEC.md §10.2 has a host take them.

## 6. The hotty catalog

`https://neuroplast.io/hotty/a2ui/v1/catalog.json`, in
[catalog/hotty/catalog.json](../catalog/hotty/catalog.json).

**Names.** Every component's name starts with `Hotty` and every function's
with `hotty`: the plain noun or verb A2UI would use, after the prefix
(`HottyForm`, `hottyFocus`). A2UI scopes a name by its catalog, so a
`Form` in another catalog would be legal, but a prompt that holds both
catalogs, or a composite catalog made of them, would mix the two up. A2UI
reserves no component prefix, so this one can never be taken. When A2UI
adds a component that does what one of these does, the renderer maps it
onto the same view, and this one is deprecated (`deprecated`,
`x-deprecated-reason`). Extension keys stay under `io_neuroplast_hotty`.

### 6.1 HottyShortcut

Binds a key to its surface, and draws nothing. `key` is a key as HOTTY
SPEC.md §10.4 names it: a W3C UI Events key value after its modifiers
(`Control`, `Alt`, `Meta`, `Shift`), joined by `+`: `Control+s`,
`Escape`, `Alt+ArrowUp`, `?`, `Space`. Modifiers match in any order; a
character carries Shift in itself, so `Control+S` is Control and Shift
with s. Tab and Shift+Tab move focus and cannot be bound.

When a key reaches the program (§5) while the HottyShortcut's surface is
the one keys apply to, the HottyShortcut either presses the Button `press` names,
as a click would and only if the Button's checks pass, or runs its
`action`. `press` is a plain id, not a ComponentId: a HottyShortcut does
not contain its Button; in a template, the first instance in tree order is
pressed. A key a focused text field uses never reaches a HottyShortcut: a
character it types, or a key its keymap binds to an edit (§5, SPEC.md
§10.2). Give HottyShortcuts keys the keymap leaves alone, such as Control+s,
or bind the key to `program` for the fields (§6.5).

### 6.2 HottyForm

A container whose `onSubmit` runs on Enter in a text field inside it (on a
host, HOTTY's `submit`). It runs only when the HottyForm's own checks and those
of every control inside pass; otherwise each failing control shows its
error, and nothing goes to the agent. A control's error shows once the
user has changed it, or tried its HottyForm. Fields commit first (§5), so
`onSubmit`'s context has what was typed. HottyForms do not nest on a host:
an inner HottyForm is a group, and Enter submits the outer one.

### 6.3 hottyFocus and hottyBlur

Renderer functions that a Button or the agent may call
(`allowedCallers: rendererOrAgent`). `hottyFocus({id})` gives the keyboard to
a component of the surface: a template's component, to the instance in
the caller's scope. Called by the agent (`callRendererFunction` names no
surface), it focuses the first surface that has the component. `hottyBlur()`
gives the keyboard back: from the caller's surface, or from every surface
when the agent calls it. On a host they are `a=focus` and `a=blur`; a
field that had the keyboard commits first.

### 6.4 autofocus

`metadata.extensions.io_neuroplast_hotty.autofocus: true` on a component
that takes focus gives it the keyboard as soon as it is there, until the
surface has given the keyboard to anything. A surface often arrives after
it is made, in later `updateComponents`; autofocus waits for the
component. Other renderers ignore it.

### 6.5 keys

`metadata.extensions.io_neuroplast_hotty.keys` on a component is a
keymap, as SPEC.md §10.2's `data-keys`: bindings `key=action` separated
by white space, such as `Control+s=submit Control+k=program`. It applies
to the text fields inside the component, and to the component itself if
it is one, over the surface's keymap (§5) and those of the components
around it, key by key, the nearest last. On a host it is the element's
`data-keys`; in cells the rendition resolves it the same way. Other
renderers ignore it.

## 7. Fallbacks

A component never fails the surface it is in.

- **Video and AudioPlayer** are a labelled link, `▶ Video` or `▶` and the
  description: on a host a link the program opens (SPEC.md §9), in cells
  an OSC 8 hyperlink, in text the label and the URL. It takes focus, and
  Enter or Space opens it, through the renderer's `openUrl`.
- **Image** is a picture on a host, fetched only as the network policy
  allows (§2). In cells and text it is `[image: description]`.
- **A child still to come.** A component may name one the agent has not
  sent yet, as a stream does. It shows `…` in `muted` until it comes.
- **A component that contains itself.** Where the loop closes, it shows
  `! Type` in `warning`; the renderer reports the cycle to the agent once,
  and the rest of the surface shows.
- **A component this renderer cannot draw**, of a catalog the processor
  knows but the renderer has no rendition for, shows `! Type` too. A
  component of a catalog the surface does not support is the agent's
  error: A2UI refuses the message, and the renderer reports it.
