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
| Icon | `span role=img`, holding an inline `svg` 1em square whose one `path` is filled with `currentColor`, so it takes the text's colour, the theme's or a Button's. One of the 59 names is its Material Symbols shape (Sharp, filled; `favoriteOff` and `starOff` unfilled), generated into `icons/` by `make icons` (gov R-4). An `svgPath` is drawn as A2UI's reference renderers draw it, in a 24 box, and only when it is path data and nothing else, at most 8 KB (`icons.Path`). The kit writes the `svg` itself, so a path is only ever an attribute. A name with no shape, or a path it doesn't accept, is the glyph cells draw. A HottyIcon is drawn the same way (§6.15); its path, with a `strokeWidth`, is stroked with `currentColor` that wide, with round caps and joins, and not filled |
| Video, AudioPlayer | `a` with `href`: a link the program opens (§7) |
| Divider | `hr`, or a vertical rule |
| Button | `button type=button`, `disabled` while its checks fail. A Row as its label stays a row, so an icon is beside its text. A vertical List's Button is its row (`k-item`), as a menu's: as wide as the List, its label at the start, its variant only a fill, so that focusing or picking a row moves nothing |
| TextField | `input` (`text`, `password`, `number`) or `textarea`, with `data-on=input` |
| CheckBox | `input type=checkbox` |
| ChoicePicker | one of several shown as checkboxes is a select: a `button aria-haspopup=listbox` with the picked option's label and a caret, whose list opens in a surface of its own (below); else its options, as checkboxes (several) or chips (`button aria-pressed`). Hosts draw `select` unevenly (Blitz not at all), and a select's list is the program's to place (SPEC.md §9) |
| Slider | `button role=slider` drawing the track (its rail, fill and knob), between `−` and `+` buttons out of the Tab order, then an `output` as wide as the widest value (§3.3's), in `ch`. The track is cut into notches, one a value (at most 41; twenty steps without a `step`), each centred on where the knob stands at its value and half one at either end, so that a tap, a click, sets the value tapped. On a host with `steps` (SPEC.md §4, §9.1) the track is the drag target, `data-on="drag"` and `data-steps` its notches' count less one, and the notches take `click` alone: a drag sets the value of the step under the pointer wherever the pointer goes, off the track and out of the surface too, until it is let go, and the click it ends with changes nothing more. Without `steps` each notch has `data-on="drag click"`: a drag sets the value of the notch under the pointer, and stops where there is none. The track's `touch-action: pan-y` lets a finger drag it: one that moves along the track drags, and one that moves up or down scrolls. Hosts draw `input type=range` unevenly (Blitz not at all) |
| DateTimeInput | `input type=text` with the ISO 8601 value, its form as the placeholder, as in cells (§3.5). Hosts draw date and time inputs unevenly (Blitz not at all), and none takes an offset such as `Z` |
| Tabs | a `tablist` of `button role=tab`, then the tab shown |
| Modal | its trigger; while open, its content in the layer, over a backdrop, the surface `inert` |
| HottyForm | `form` with a hidden submit button out of the Tab order, so Enter submits |
| HottyProgress | `div role=progressbar` with `aria-valuemin`, `aria-valuemax` and, when it has a value, `aria-valuenow`: the label, then a rounded track with a fill as wide as the fraction (in `--k-info`, `--k-success` once full) and an `output` with the percentage. Without a value, a quarter of the track sweeps across it with the clock (§3.4): the element's `--k-at` is where it starts, so that a tick is one attribute's delta |
| HottyTable | `div tabindex=0 role=grid` holding a `table`: the header in `thead`, then in `tbody` the rows the view shows (the same window as cells, §3.4), each a `tr data-on=click` whose id is the table's and `~y` and the row's index, `aria-selected` on the selected one, which is filled (`--k-tonal`, tinted with `--k-focus` while the table has the keyboard). Its `data-keys` give the program the arrows, Page Up, Page Down, Home and End (SPEC.md §10.2, keys for the program), on a host that would scroll with them; Enter reaches the program anyway, a focused box using no keys. While it scrolls, a note under it says which rows show. A column's `width` is for cells: on a host the table lays its columns out. A host's own scrolling, a sticky header and the row under the pointer are vault KIT-01h |
| HottyList | `div tabindex=0 role=listbox`, as a HottyTable's box: its title (or, while its filter is typed, `Filter:`, the text and a caret that shows while it has the keyboard), its status line (`~u`), then the items of the page the view shows (the same page as cells, §3.4), each a `div role=option data-on=click` whose id is the list's and `~i` and the item's index, its label over its description in `--k-muted`, `aria-selected` on the selected one, which is filled with `--k-tonal` and has a 3px bar at its start, in `--k-border`, and in `--k-accent` with its text while the list has the keyboard; a label's characters that matched the filter in `span.k-match`, underlined; then a dot for each page, the page shown's in `--k-fg`. Its `data-keys` give the program what a HottyTable's do, the arrows left and right and Space; the characters its filter types, Backspace, Enter and Escape reach the program anyway, a focused box using no keys, so the filter is typed as in cells. Its empty text shows when it has no items. Two-line rows in proportional type and the item under the pointer are vault KIT-04h |
| HottyKeyHints | `div`: in the short view, a line of hints, each a `kbd` (the key, in `--k-muted`) and what it does (fainter), ` • ` between them, cut where it does not fit; in the full view, its groups (§6.10) side by side, `4ch` apart, each a grid of keys and what they do. The host moves focus among the elements it works itself (fields, boxes, Buttons) without telling the renderer (the keyboard, below), so their keys are left out, where a guess would go wrong at the next Tab; a HottyTable's, a HottyList's, a HottyDiff's, a HottyTree's, a Slider's and a select's show, as the program's keys go there. `?`, which a focused field types, reaches the program from anywhere else and switches the views. Keycaps are vault KIT-08h |
| HottyScrollView | `div tabindex=0`, a box `--k-rows` terminal rows tall (its `height`, by SPEC.md §8's `--hotty-cell-h`), filled with `--k-tonal`, a ring while focused, whose content overflows it: the host scrolls it, with its own scrollbar, the wheel, a touch drag and the keys a browser scrolls with (SPEC.md §5.3; the storybook places its surfaces with `scroll`). Its child goes inside as itself; its lines are a `div` (`~j`) of a row each, in the mono face and `white-space: pre`, so that the box shows `height` of them, cut at the box unless it wraps them. A log's (`follow`) box is `role=log`, which a screen reader reads as lines arrive, and a line written to the next index arrives as an `append` delta. A program cannot set where a host has scrolled, so the box starts at its top, `follow` or not, and `hottyScrollTo` does nothing there. Its `data-keys` binds the keys bubbles' viewport scrolls with to the host's scroll actions (SPEC.md §10.2, *Scrolling keys*), so that they scroll it as in cells (§3.7): j and k, f, b, Space and Shift+Space, u and d (Control+u, Control+d), g and G, and h and l while its lines are cut. A Button in it keeps Space and a field in it types the letters, since a key an element uses stays its own and a field leaves scroll actions out. Following the tail on a host is vault KIT-07h |
| HottyCode | `div` in the mono face, filled with `--k-tonal` as a Text's code block is, a flex row (`k-cl`) for each line: its number, right-aligned as wide as the widest (`--k-ln`), in `--k-muted` and `aria-hidden`; its mark's sign, when the code has marks; then its code (`k-src`), a span for each token that is not plain, of class `k-t-` and its kind, which the sheet colours with the roles cells uses (§3.4). A marked row has `k-m-` and its kind: a highlighted one is filled with `--k-selection`, the others with a sixth of their role's colour (`color-mix`). Lines wrap, `pre-wrap`; with `wrap` false they do not, and the box scrolls sideways, which the host does itself (SPEC.md §5.3), every row as wide as the widest so that a tint reaches the end. A Text's fenced code block is goldmark's `pre` and `code`, its tokens in the same spans |
| HottyDiff | `div` in the mono face, filled with `--k-tonal` as a HottyCode's; with hunks, `tabindex=0 role=listbox`, whose `data-keys` gives the program ArrowUp, ArrowDown, Home, End, k, j, g and G (a HottyScrollView around it binds the letters to scroll actions, SPEC.md §10.2). For each file with a name (or of several), a row (`k-diff-file`): its name, bold, then what the change adds and removes, `+N` in `--k-success` and `-M` in `--k-error`, or new, deleted or binary. A run of unchanged lines left out is a muted row, `⋯ N unchanged lines` (`k-fold`). Each hunk is a `div tabindex=-1 role=option data-on=click` (`k-hunk`), whose id is the diff's, `~b` and the hunk's index: its header in info and its section muted (`k-hh`), then its lines. Unified, a line is a flex row (`k-dl`): its old number and its new (`--k-lo` and `--k-ln` wide, `aria-hidden`), its sign, then its code, its tokens in a HottyCode's spans. Split, the hunk is a grid of two equal columns, the old side and the new (`k-ds`), paired as cells pairs them, a rule between them and the header across both. A removed line (`k-d-del`) is filled with a sixth of `--k-error`, an added one (`k-d-add`) with a sixth of `--k-success`, and its changed words (`k-w`) more. The selected hunk (`k-sel`, `aria-selected`) has a rail on its left and its header reversed, muted, and in `--k-focus` while the diff has the keyboard (`k-on`). It then has the host's focus itself, so that the host scrolls it into view (SPEC.md §5.3) as cells keeps it in sight. A split diff stays split at any width, its sides wrapping: the markup does not know the width, and a host has no container queries |
| HottyTree | `div tabindex=0 role=tree`, as a HottyList's box, holding the nodes the view shows (the same rows as cells, §3.4), each a `div role=treeitem data-on=click` whose id is the tree's and `~q` and the node's index, with `aria-level`, `aria-expanded` on a branch and `aria-selected`. A row is a flex row indented `--k-5` a level (`--k-level`): its fold (`k-node-fold`, a column wide), `▸` or `▾` in `--k-muted` before a branch and blank before a leaf, so that a level's labels line up; its icon, as an Icon draws it, or a blank one as wide where another node has an icon; then its label, the filter's matches in `span.k-match`. The selected node is marked as a HottyList's selected item: filled with `--k-tonal`, a 3px bar at its start in `--k-border`, and in `--k-focus` with its text while the tree has the keyboard (`k-on`). Its `data-keys` are a HottyList's; its letters and Enter reach the program anyway, a focused box using no keys. Each row is `--k-node-h` (2rem) tall. With a `height`, the box is that many rows tall and holds every node that shows, and the host scrolls it (SPEC.md §5.3), with the wheel and a thin scrollbar; the selected node is `tabindex=-1` and has the host's focus while the tree has the keyboard, so that the host scrolls it into view as the selection moves, as a HottyDiff's selected hunk does. Guides are cells' (§3.4): a host has the room to indent instead. Its empty text shows when no node does |
| HottySpinner | `span role=status`: the frame of its set the clock is at (§3.4) while it spins (blank while it does not), `aria-hidden`, in the mono face and as wide as the set's widest frame in cells (in `ch`), then the label, which so stays put as in cells. The frame's id is the Spinner's and `~f`, so that a tick is one text's delta |

**Updates are deltas.** The document goes once. After it, the renderer
diffs the elements it sent against the elements the view makes now, and
sends the smallest deltas the ids allow (SPEC.md §6): `attr` and `unattr`
for an attribute, `text` for an element whose content is one text, a
recursion into children that keep their ids and order, `append` for
children added after them, and `inner` (which morphs) for any other
change below an element. `updateComponents`
and `updateDataModel` change a surface this way, and so does the
renderer's own state: a tab shown, a Modal opened.

**Events.** Each acts on the surface as the user's act does in cells:

| event | does |
| --- | --- |
| `input`, `change` | writes the control's value: to its bound path, else as the renderer's |
| `click` | a Button runs its action; a Table's row or a HottyList's item is selected, or acted on once it is; a tab is shown; a chip toggles; a link opens; a Modal's trigger opens it; the backdrop closes it; a Slider's `−` or `+` steps it, and a tap on its track sets it; a select opens or closes its list, an option in the list is picked, and a click beside the list closes it |
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
element last clicked or edited, not always the one focused. Edited means
typed in (`input`), or a control's `change`, which comes at once: a text
field's `change` is its commit as the host's focus leaves it, often for
where the renderer just moved the keyboard, so it says nothing about where
the keyboard is.

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
| HottyProgress | its label's width or 25 (a bar of 20 and ` 100%`), whichever is wider |
| HottySpinner | its set's widest frame, plus 1 and the label when it has one |
| HottyTable | its columns' widths, each a column's `width`, else its header's or its widest cell's, whichever is wider, plus 2 for each column (a column of padding each side) |
| HottyList | 2 for the indent, plus its widest line: its title + 2, its status line, its empty text, `Filter: ` + 1 when it filters, each item's label and description |
| HottyKeyHints | its short line, or in the full view its columns, uncut |
| HottyScrollView | its widest line unwrapped (whether it wraps them or not), or its child's natural width, plus 2 for the scrollbar |
| HottyCode | its gutter (§3.4) and its widest line |
| HottyDiff | 1 for the rail, then its gutter (§3.4) and its widest line; split, twice a side's gutter and that line, and 3; or a file's row, a hunk's header or a fold, if one is wider |
| HottyTree | 2 for the bar, plus its widest node, whether it shows or not, so that it keeps its width as branches open and close: its guides (4 a level below the roots), its fold (2), its label and, on a branch, a space and its count; or its empty text |
| Image, Icon, HottyIcon, Media, Placeholder | what they paint (§3.4) |
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
- A HottyProgress's is its label's longest word, and at least 8: a bar of 3
  and the percentage.
- A HottyTable's is its columns', each at most 3 wide, with their padding.
- A HottyList's is its indent and 4.
- A HottyKeyHints's is 1: it cuts what does not fit (§3.4).
- A HottyScrollView's is its child's, or 1 for its lines (it cuts them or
  wraps them), plus 2 for the scrollbar.
- A HottyCode's is its gutter and 8 columns of code, or its widest line
  when that is narrower.
- A HottyDiff's is its rail, its unified gutter and 8 columns of code.
- A HottyTree's is its bar, a level of guides and 4.
- Any other control's is its natural width.
- For containers, a Row adds its children's minimums and the columns
  between them; a Column, a HottyForm, a Modal and a Tabs take their widest
  child's; a Card takes its content's plus 4.

How the containers lay their children out:

- **Column** covers a Column, a vertical List, a HottyForm, a Modal's trigger,
  and the content of a Card or a tab.
  - Its children are stacked with no rows between them, except for one
    blank row between two controls (fields, Buttons that are not a
    List's rows, HottyProgress, HottySpinner, HottyTable, HottyList and
    HottyScrollView)
    when either has a title row (§3.4; a HottyProgress's label is one, a
    HottyTable's header, and a HottyList's status line), or when one is a field and the other is not: huh's space between its fields, and bubbles'
    before a form's button. A stack of CheckBoxes, of Buttons, or of
    HottySpinners stays tight. A HottyKeyHints has a blank row before it,
    whatever comes before, as bubbles' help sits a row under its list.
    A HottyScrollView has one after it, whatever comes after: its box
    draws no edge but the scrollbar, so what follows would read as its
    content. So do a HottyCode and a HottyDiff, which draw no edge at all. Before it, the rule for controls applies, so a title above
    it stays on the row before its box.
    Through a Row, a Column or a HottyForm, the rule sees its first child
    (or, before it, its last).
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
| Icon, HottyIcon | its glyph (`icons.Glyph`, width 1, by the basic name or Material's for one of the 59; `◇` for any other name and for an `svgPath`) |
| Video, AudioPlayer | `▶ Video`, or `▶ ` and its description, underlined and linked to its URL (OSC 8) |
| Divider, Card, Tabs | §3.3 |
| Button | `[ label ]` on one row. The label is the plain text of the Button's Texts and the glyphs of its Icons, a space apart. Primary is bold; borderless drops the brackets and is underlined; disabled is `muted` and faint. A vertical List's Button is its row: ` label `, its style (focus's reverse, say) across the List; bold unless borderless, and never underlined. |
| TextField, DateTime | in the field's box past the gutter: a title row, then the value (§3.5), then the error |
| CheckBox | past the gutter, `[•] label` or `[ ] label`, then the error |
| a select (one value, `checkbox` display) | past the gutter, a title row, then the picked option's label, or `…` in `muted` when none is picked, and ` ▾` in `muted`. While its list is open (§3.7), the options follow one a row: `● label` for the picked one and `○ label` for the others, each after `  `, or after `> ` in `accent` on the highlighted row, whose label is `accent` and bold. Then the error. |
| a Choice's options (several values, or `chips`) | past the gutter, a title row, then the options: one a row, `[•] label` or `[ ] label` after `  `, or after `> ` in `accent` for the option with the keyboard (huh's multiselect); chips two columns apart, wrapping, `( label )` and `(● label)` when picked. Then the error. |
| Slider | past the gutter, `label ━━━━●──── 50` on one row: the label as a title and a space, then the track, a space and the value. The track fills the columns left. Up to the knob it is `━`, the knob is `●` at round((value − min) / (max − min) × (track − 1)), and after it the track is `─` in `border`. When there is no label, or the track would be shorter than 3, the label is dropped. Then the error. |
| HottyProgress | its label on a row, when it has one, then the bar and the percentage on the next: ` 42%`, five columns (`%3.0f%%` after a space, half to even). The bar takes the columns before them, filled in eighths of a cell: `█` for each full cell, then one of `▏▎▍▌▋▊▉` for a cell part filled, blending from `info` into `accent` across the bar (§3.6), or `success` alone once full. The rest is `░` in `border`. Without a value, a segment of a quarter of the bar (at least one cell) is `█` in `info`, moving with the clock (below), and the percentage's columns are blank. When the bar would be shorter than 3, the percentage is dropped. |
| HottyTable | a header row, bold; a rule of `─` in `border`; then a row for each row the body shows, a cell for each column: its text, padded a column each side, cut with `…` when wider than its column, at the start, the centre (rounded down) or the end as `align` says. The columns take their natural widths; while the table is too wide, a column at a time comes off the widest that is wider than 3 (the first of equals), and once none is, off the widest. The body shows `height` rows, or all of them (one, `No rows` in `muted`, when it has none): from the first it showed, moved as little as brings the selected row into view, never past the last (the view's `Top`, the same in both renditions). While it scrolls, the rule ends with ` 4–10 of 12 ` in `muted` and one more `─`. The selected row is reversed across the table, in `accent` while the table has the keyboard and in `muted` otherwise. |
| HottyList | as bubbles' list with its default delegate, every line 2 columns in: its title, ` Title ` in `accent` reversed, and a blank row (neither without a title); its status line in `muted` and a blank row; then the items of the page that shows, each its label and, under it, its description in `muted` (one row an item, with no blank rows between, when none has a description; else two, a blank row between). The selected item's first two columns are `│ `: in `accent` with its label and description while the list has the keyboard, and in `muted` otherwise, its text plain. A label too wide is cut with `…`; its characters that matched the filter are underlined. With a `height`, a page shows that many items, from the page that holds the selected one, and the body keeps their rows on every page; after it, while the items take more than a page, a blank row and a dot for each page (`•`, the page shown's in `fg`, the others in `border`; `3/10` in `muted` when the dots do not fit), whose two rows stay while a filter leaves one page. The status line is `12 items` (`1 item`, `No items`); while a filter applies, `“query” 7 items`, or `Nothing matched`, then ` • 5 filtered`. While the filter is typed, `Filter: ` in `accent` and the text take the title's row, with the cursor after them. With no items, the empty text shows in `muted` in the body. |
| HottyKeyHints | as bubbles' help draws it: the short view one line, each key in `muted`, a space, what it does in `muted` faint, and ` • ` in `border` faint between them (bubbles' three steps; the Terminal theme's `muted` and `border` are one colour); the hints that do not fit go, and ` …` ends the line where it fits. The full view (after `?`, §3.7) is its groups (§6.10) as columns four apart, each a row a key, the keys padded to the widest, then a space and what they do; a group that does not fit goes, and ` …` follows the first row where it fits. |
| HottyScrollView | a box `height` rows tall, as bubbles' viewport: its content, 2 columns narrower than the box, then a blank column and the scrollbar down the last. Its lines are a row each, cut at the box and scrolled sideways six columns at a time, or with `wrap` broken between characters into the rows they take; a child is laid out at the content's width, and shows through the box at the rows scrolled to: what it hides takes no click, and the element with the keyboard inside it is scrolled into sight. It starts at its top, or with `follow` at its end, where it stays as lines arrive until the user scrolls up, and follows again once back at the end. The scrollbar, while the content is taller than the box, is a track of `│` in `border` faint and on it a thumb of `┃` as long as the share that shows (at least a row), where it shows, in `muted`, and in `accent` while the box has the keyboard. bubbles' viewport draws no scrollbar: without one, a box with no edge gives no sign that it scrolls |
| HottyCode | a row for each line, or with `wrap` (the default) as many as the line takes, broken by cluster under the code: first its number, right-aligned as wide as the widest, in `muted`, and a space, when `lineNumbers` is on; then, when the code has marks, the line's sign and a space (`▎` in `info` for highlight, `+` in `success` for added, `-` and `✗` in `error` for removed and error, `!` in `warning`); then the code. Continuation rows leave the gutter blank. Without `wrap` a long line is cut with `…`. Each token is coloured by its kind (package `highlight`): keywords `info` and bold, types and builtins `info`, functions' names bold, strings `success`, numbers and constants `warning`, as are preprocessor lines, decorators and attributes, comments `muted` and italic, a diff's added and removed lines `success` and `error`, its hunk headers `info` and bold, and the rest `fg`. Not `accent`, which marks only focus (§3.6). A marked line's rows are tinted across the width (§3.6): toward `selection` itself for highlight, a sixth of the way toward `success`, `error` or `warning` for the others. As OpenTUI's Code and LineNumbers; glamour, the reference shot's, draws code blocks with neither numbers nor marks |
| HottyDiff | a column for the rail, then rows. For each file: a blank row before all but the first; its name when it has one (or the diff has several files), bold, then `+N` in `success` and `-M` in `error`, or `new`, `deleted` or `binary`. For each hunk: `⋯ N unchanged lines` in `muted` when it leaves lines out before it (and after the last, where the diff knows, from two texts); its header in `info` and its section in `muted`; then its lines. Unified, a line is its old number and its new (blank on the side that does not have it), right-aligned as wide as the widest, in `muted`, each and a space, when `lineNumbers` is on (the default); its sign (`-` in `error`, `+` in `success`) and a space; then its code, its tokens coloured as a HottyCode's, wrapped under the code or, without `wrap`, cut with `…`. Split, where each side has room for 16 columns of code (unified where not): the old side, ` │ ` in `border`, then the new, each a number, a sign and code; a context line on both, and in a run of removed lines followed by added ones, each removed line beside the added line that replaces it; a row is as tall as its taller side, and a side with no line is blank. A removed line's rows are tinted a sixth of the way toward `error`, an added one's toward `success`, its changed words a third (§3.6). The selected hunk's rows have `▎` in the rail and its header is reversed, both in `muted`, in `accent` while the diff has the keyboard; a scroll view around it keeps that hunk in sight. With no hunks and no name, `No changes` in `muted`. As OpenTUI's Diff, which has no names, folds, word marks or selection |
| HottyTree | as lipgloss's tree draws one, a row a node, every row 2 columns in: for each level below the roots, its guides in `border`, `│   ` down past an ancestor a sibling follows (else four blanks), then `├── ` before a node a sibling follows and `└── ` before the last; a branch's fold, `▶ ` closed or `▼ ` open, in `muted`, as bubbles' tree has it, or a root leaf's two blank columns, so that the roots line up; its label, the filter's matches underlined, cut with `…` when it does not fit; and a closed branch's count, ` 3`, in `muted`. A node's siblings are those that show, so that while a filter applies a line ends at its last match. The selected node's first two columns are `│ `: in `accent` with its label, bold, while the tree has the keyboard, and in `muted` otherwise. With a `height`, that many rows show: moved as little as brings the selected node into view when the selection moves, and three rows a notch by the wheel, which leaves the selection where it is (§3.7). A node's icon is a host's: few have a glyph a column wide, and the guides and folds say what a tree is. With no node shown, its empty text in `muted`. |
| HottySpinner | the frame of its set the clock is at, in `info`, then the label after the widest frame's columns and a space, so that it stays put. One that is not active leaves the frame's columns blank. |
| Placeholder | `…` in `muted` while pending; `! Type` in `warning` when the type is unknown or the component contains itself |
| an error | `✗ message` in `error` under its control, wrapped, its continuation lines indented 2 |

**Text blocks:**
- A heading is bold, and an h1 is also underlined.
- A list item is indented two columns a level. Its marker is `• `, `1. `,
  or `☐ ` / `✓ ` for a task. Continuation lines align with the text.
- A quote starts every line with `▎ `, once per level, in `border`.
- A code block is two columns in, as glamour sets one off from the
  prose, its lines broken by cluster at the width rather than wrapped. It
  is highlighted as its fence says (```` ```go ````), its tokens coloured as a
  HottyCode's, its plain ones in the Text's colour; without a language, or
  with one no lexer knows, it is plain.
- A rule is `─` across, in `border`.
- Inline: bold, italic and strikethrough are attributes, and code is
  `muted`. A link is underlined and its cells carry the URL (OSC 8).
- A caption is `muted` throughout.

**The clock.** What moves, in cells and on a host, moves with the wall
clock, not with the draws. For t the Unix time, a HottySpinner shows
frame ⌊t / interval⌋ mod its frames. An indeterminate HottyProgress is at
step k = ⌊t / 100 ms⌋ mod 50 of a five-second sweep: its segment starts
at column ⌊k × (bar + segment) / 50⌋ − segment in cells, and at
k × 125% / 50 − 25% of the track on a host. Each rendition says how soon
what it last made changes by itself (`Rendition.Animating`: the shortest
interval it showed, 0 while nothing moves): a draw in cells, a Doc or an
Update on a host. The program draws or updates again then: the
storybook's `Book.Tick` and `storybook -bare` do, on Bubble Tea's
`tea.Every`. So two renditions of one surface made at once show the same
frame, and a test sets the clock (`Rendition.Clock`, in both).

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

A cell may blend two roles (a HottyProgress's fill): k/255 of the way
from the first to the second, which is, channel by channel, p + (q − p) ×
k / 255 in integers, truncated. It takes effect where the theme colours
both roles; elsewhere, and at the ANSI-16 floor, the cell is the first
role's.

A cell's background may be tinted (a HottyCode's marked line): k/255 of
the way from the theme's `bg` to a role's colour, as above. It takes
effect where the theme colours both; with the terminal's background, and
at the ANSI-16 floor, there is no tint, and the line's sign says it alone.
A HottyDiff's changed lines and words are tinted so too. Where they cannot
be, a changed line's text takes its role's colour in place of its tokens',
and its changed words are reversed as well, as git's diff-highlight shows
them; under NO_COLOR the reverse still marks them.

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
- **A HottyTable** moves its selection by ArrowUp and ArrowDown (k and j)
  a row, by PageUp and PageDown the rows its body shows, and by Home and
  End (g and G) to its first row and its last, clamped, as bubbles'
  table does; with no row selected, each selects the
  first. Enter acts on the selected row (§6.8). On a host the table is a
  focused box, which leaves every key to the program, so the renderer
  moves it there too, with the same rules (`view.Controller.TableKey`).
- **A HottyDiff** moves its selection by ArrowUp and ArrowDown (k and j)
  a hunk, and by Home and End (g and G) to its first hunk and its last;
  with no hunk selected, each selects the first. Enter acts on the
  selected hunk (§6.13). On a host its box gives those keys to the
  program (§2), so the renderer moves it there too
  (`view.Controller.DiffKey`).
- **A HottyList** takes keys as bubbles' list does
  (`view.Controller.ListKey`, in both renditions): ArrowUp and ArrowDown
  (k and j) move its selection an item; ArrowLeft and ArrowRight (h and l,
  PageUp and PageDown) turn a page, to the same place on it, clamped to
  its items, and not past the first page or the last; Home and End (g and
  G) go to its first item and its last. With none selected, each selects
  the first. Enter acts on the selected item, while it shows (§6.9). With
  `filterable`, `/` starts the filter: then a character, Space among them,
  adds to it, Backspace takes the last off, ArrowUp and ArrowDown still
  move, Enter applies it (and drops it when it leaves nothing), and Escape
  drops it. Escape also drops a filter that applies. Each change to the
  filter selects the first item it leaves, as bubbles' does.
- **A HottyTree** moves its selection by ArrowUp and ArrowDown (k and j)
  to the node above or below that shows, by PageUp and PageDown its
  height's rows, and by Home and End (g and G) to its first node and its
  last; with none selected, each selects the first. ArrowRight (l) opens
  a closed branch, else goes to its first child; ArrowLeft (h) closes an
  open branch, else goes to the parent. Enter and Space open or close the
  selected branch, and act on the selected leaf (§6.14). While a filter
  applies, every branch shows open and stays so. On a host its box gives
  those keys to the program (§2), so the renderer moves it there too
  (`view.Controller.TreeKey`).
- **A HottyScrollView** scrolls as bubbles' viewport does: ArrowUp and
  ArrowDown (k and j) a row; PageUp (b) and PageDown (f, Space) a page;
  u and d (Control+u, Control+d) half a page; Home and End (g and G) to
  its ends; and ArrowLeft and ArrowRight (h and l) six columns, while its
  lines are wider than the box and do not wrap (else they fall through).
  Each is clamped, and the box takes the others even at an end, as
  bubbles' does. Its child's controls take keys as they would anywhere;
  the box is a stop of its own in the Tab order. On a host it is a box the
  host scrolls itself (§2), with the keys a browser scrolls with and the
  same letters, bound to the host's scroll actions. A host older than
  those (SPEC.md §10.2, *Scrolling keys*) gives the letters to the program,
  which cannot scroll the box, so they do nothing there.
- **Space and Enter** activate a Button, a Tabs' title, a chip, a Media
  link, or a Modal trigger that is not a control.
- **A CheckBox, and an option shown as a box,** is a checkbox on a host,
  and takes keys as one (SPEC §10.2): Space toggles it, and Enter submits
  its HottyForm, or does nothing outside one, as huh's Enter moves on.
- **Tab and Shift+Tab** move the keyboard in tree order.
- **Any other key**, Escape among them, goes to the surface's
  HottyShortcuts.
- **?**, when none of these took it, switches a HottyKeyHints between
  its views (§6.10). A field and a HottyList's filter type it.
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
  - a click on a HottyTable's row selects it, and a click on its selected
    row acts on it (§6.8); a click on its header only focuses it; so too
    a HottyList's items (§6.9), and its title and status line, and a
    HottyDiff's hunks (§6.13), its files' names and folds;
  - a click on a HottyTree's node selects it, and opens or closes it
    when it is a branch (while no filter applies), or acts on it when it
    is the selected leaf (§6.14, `view.Controller.ClickNode`);
  - a click on a Slider's track sets the value at that column:
    min + (max − min) × column / (track − 1), stepped and clamped, and
    the value follows the pointer while the button stays down;
  - anything else is activated.
- **A disabled Button** does not take the keyboard, but the click still
  activates it, as in rendition/html; the controller decides what that
  does.
- **A click on nothing that takes focus** gives the keyboard back.
- **A click outside an open Modal's panel** closes the Modal.

**The wheel** over a HottyTree with a `height` moves its rows, three a
notch, while they can move that way. Else it scrolls the innermost
HottyScrollView under the pointer that can still move that way: three rows a notch, or with Shift six
columns sideways, as bubbles' viewport takes it. A notch past an end goes
to the box around it, if any. The storybook passes the wheel to the story
under the pointer; a program without mouse reports has no wheel.

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
- A HottyProgress is `Label: 42%`, or `Label: …` without a value; a
  HottySpinner is `Label: …` while it spins, else its label alone.
- A HottyTable is its header and every row, whatever its `height`: the
  columns two spaces apart, each as wide as its widest cell and aligned as
  it says, after `> ` for the selected row and two spaces for the others.
- A HottyList is its title, its status line, and every item its filter
  leaves, whatever its `height`: its label, ` — ` and its description,
  after `> ` for the selected item and two spaces for the others; its
  empty text when it has no items.
- A HottyTree is every node it shows, whatever its `height`, drawn as
  cells draws it (its guides, folds and closed branches' counts), after
  `> ` for the selected node and two spaces for the others; its empty
  text when it shows none.
- A HottyScrollView is all of its content, whatever its `height`: its
  lines, a line each, or its child.
- A HottyCode is its code as it is, each line after its number when the
  numbers show, and its mark's sign when it has marks (`>` for highlight,
  `+`, `-`, `✗`, `!`), as a pipe reads a diff.
- A HottyDiff is a unified diff that patch reads, split or not: each
  named file's `---` and `+++` lines (git's `a/` and `b/`, `/dev/null`
  for a side it does not have), then its hunks, each its header and its
  lines, signed; a binary file is git's line for one.
- Tabs are their titles, the one shown in brackets, then its content.
- A HottyKeyHints says nothing: a pipe takes no keys.
- An error is `✗ message`, on the line after its control.
- Image, Icon, Video, AudioPlayer and placeholders are as §7 has them; a
  HottyIcon is an Icon's glyph.
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

### 6.6 HottyProgress

How far a task has come: a bar and a percentage under its `label`, as
bubbles' progress. `value` (a DynamicNumber) runs from 0 to `max` (1 by
default, so a fraction; 100 takes a percentage) and is held to that range.
While it is absent, or its path holds nothing, the bar is indeterminate:
it moves without saying how far. It is not a control: it takes no focus
and sends nothing. The agent moves it with `updateDataModel` as the task
goes.

### 6.7 HottySpinner

That something is under way, with no measure: a spinner and its `label`.
`spinner` names its set, bubbles' spinner's frames at bubbles' rates
(`view.Spinners`, MIT): `dot` (the default), `line`, `miniDot`, `jump`,
`pulse`, `points`, `globe`, `moon`, `monkey`, `meter`, `hamburger` and
`ellipsis`. `active` (a DynamicBoolean, true by default) says whether it
spins; bound to a path that holds nothing, it does not. One that stops
keeps its label in place, so the agent can stop it, or replace it with
the result, when the work is done. It takes no focus and sends nothing.

### 6.8 HottyTable

Rows of data under a header, one selected at a time, as bubbles' table.
`columns` are its columns in order, each the `key` of the row field it
shows, a `header`, an optional `width` in cells, and `align` (`start`,
`center`, `end`). `rows` are objects whose fields the columns name, best
bound to the data model, so that the agent changes them with
`updateDataModel`. A cell is its field as text (A2UI's string conversion:
a number as JavaScript writes it, a missing field empty).

A row is identified by its `rowKey` field, as text, or by its index
without one. `selected` (a DynamicString) is the selected row's
identity: bound to the data model, the user's selection is written
there, and the agent sets it there too; the body scrolls to it either
way. It is identity rather than an index, so that it stays on its row
as rows come and go.

A2UI's actions carry nothing from the renderer (vault a2ui-limits L2), so
a row is selected, then acted on: Enter, or a click on the selected row,
runs `onActivate`, whose context reads the selection from where
`selected` is bound. With no row selected it does nothing. `height`
fixes the body's rows; the rest scroll under the header.

### 6.9 HottyList

Items to pick from, a label and a line of description each, one selected
at a time, as bubbles' list with its default delegate: an optional
`title` above, and a status line that counts them. `items` are objects
with a `label`, an optional `description` and a `value`, literal or best
bound to the data model, so that the agent changes them with
`updateDataModel`. An item is identified by its `value`, as text, or by
its index without one; `selected` (a DynamicString) is the selected
item's, as a HottyTable's is (§6.8), and the page that shows follows it.
Enter, or a click on the selected item, runs `onActivate`, whose context
reads the selection from where `selected` is bound (vault a2ui-limits
L2); with no item selected, or the selected one filtered out, it does
nothing.

`height` is the items a page shows; the rest go on pages, which the
arrows left and right turn. With `filterable`, `/` starts a filter: the
items narrow, as the user types, to those whose label holds its
characters in order, ignoring case, ranked as bubbles' list ranks them
(sahilm/fuzzy, MIT: matches at the start, after a separator, at a capital
and next to each other first; fewer unmatched characters first). The
filter is the renderer's state, not the data model's: the agent sees the
selection, not the query. `emptyText` shows when there are no items
(`No items.` by default). The help line bubbles draws under its list is a
HottyKeyHints (§6.10).

### 6.10 HottyKeyHints

A help line, as bubbles' help draws one at the bottom of an app: the keys
the user can press now, each with what it does. It follows the keyboard,
so the agent places it once and never updates it. The renderer builds it
(`view.Controller.KeyHints`) from, in order:

1. The keys of the component with the keyboard, in bubbles' words:
   a HottyList's `↑/k up`, `↓/j down`, `/ filter`, `enter choose` (and,
   while its filter is typed, `enter apply filter`, `esc cancel`), a
   HottyTable's the same without the filter, a HottyDiff's `↑/k prev
   hunk`, `↓/j next hunk`, `enter choose`, a field's `enter submit` in
   a HottyForm, a CheckBox's `space toggle`, a Button's `enter press`, a
   Slider's `←/→ adjust`, a select's `enter open`, a HottyScrollView's
   `↑/k up`, `↓/j down`, `f/pgdn page down`, `b/pgup page up`; and
   `esc close` while a Modal is open.
2. The surface's HottyShortcuts that have a `label` (§6.1), each its key
   as bubbles writes keys (`ctrl+s`, `alt+←`, `pgdn`) and the label.
3. In the short view, `? more` while `toggle` is on.

The full view, which `?` switches to and back from (`toggle`, on by
default; turn it off where `?` is the surface's own key), shows those as
groups in columns, as bubbles' full help: the component's keys, more of
them, in a column or two as bubbles splits them (a list's moves, its
pages, `g/home` and `G/end`, then its filter and Enter; a table's rows,
then its pages; a diff's hunks and ends; a scroll view's rows and ends, then its pages and half
pages, then `←/h move left` and `→/l move right` for lines that do not
wrap, as bubbles' viewport's; a text field's moves, then its edits, from
its keymap, §5, two keys an action at most); the HottyShortcuts; then `tab next`,
`shift+tab back` and `? close help`.
Which view shows is the renderer's state, shared by the surface's
renditions. It takes no focus and sends nothing.

On a host, the component's keys are those of a HottyTable, a HottyList,
a HottyDiff, a HottyTree, a Slider or a select, whose keys the program works; the host moves focus
among the others without telling the renderer (§2, the keyboard). A
HottyScrollView's keys are the host's own there, so it has none.

### 6.11 HottyScrollView

A box `height` rows tall (10 by default) whose content scrolls, as
bubbles' viewport: a `child`, laid out at the box's width, or `lines`, a
list of strings, a row each (a null is an empty line). Lines are cut at
the box and scroll sideways, or with `wrap` break into the rows they
take. With `follow` it starts at its end and stays there as lines arrive,
until the user scrolls up; back at the end, it follows again. That is
the way to show a log or a command's output: bind `lines` to a list in
the data model and write each line to the next index with
`updateDataModel` (`/log/0`, `/log/1`, …); A2UI's data model has no
append, and on a host the line arrives as one `append` delta (§2). It
takes focus, as a stop of its own in the Tab order, and its keys,
scrollbar and wheel are §3.4 and §3.7. Where it scrolls is the
renderer's state, not the data model's.

`hottyScrollTo({id, to})`, a renderer function the agent or a Button may
call (`allowedCallers: rendererOrAgent`), scrolls it to its `start` or
its `end` (the default), the id resolved as `hottyFocus`'s is (§6.3). At
its end a view with `follow` follows the tail again; one without is at
its end once, and stays where it is as lines arrive. On a host it does
nothing: a program cannot set where a host has scrolled (SPEC.md §5.3),
so there a box starts at its top, `follow` or not (§2, vault KIT-07h).

### 6.12 HottyCode

Source code, as a code viewer shows it: `code`, highlighted for its
`language`, a language's name or alias (`go`, `Python`, `ts`) or a file
name (`main.go`, `Dockerfile`); without one, or for one no lexer knows,
it is plain. `lineNumbers` shows the lines' numbers, counted from
`startLine` (1 by default) for an excerpt of a longer file. `marks` point
at lines, each `{line, kind}`, or `{line, end, kind}` for a range, in the
same numbers: `highlight` (a line to look at), `added` and `removed` (a
change), `error` and `warning` (a problem); a later mark over the same
line wins, and one past the code marks nothing. `wrap`, on by default,
wraps long lines; off, they are cut in cells and scroll sideways on a
host. It has no height: a HottyScrollView scrolls a long listing. It takes
no focus; selecting and copying it are KIT-12's.

The renderer lexes it (package `highlight`, chroma's lexers) into tokens
of a few kinds, which each rendition colours with roles (§3.4), so every
theme and the terminal's own palette colour code. The kit carries the
lexers of the languages coding agents write most (`highlight/lexers`,
copied from chroma by `make lexers`); a program that wants every language
chroma knows imports `highlight/all`, at the size those take. A Text's
fenced code blocks are highlighted the same way.

### 6.13 HottyDiff

A change to code, as a review shows it: a `patch`, a unified diff of one
or more files as git diff or diff -u writes it, or the `old` and `new`
texts of one file, named `file`, compared with `context` unchanged lines
around each change (3 by default). The patch's hunks are shown as they
are: a hunk cut short ends where it stops. Each line is highlighted for
`language`, or for its file's name. `view` is `unified` (the default),
one column, or `split`, the old side beside the new where there is room
(§3.4); `lineNumbers` and `wrap` are on by default, as a HottyCode's.
Unchanged runs left out between hunks are folded, a row that says how
many lines they are. It has no height: a HottyScrollView scrolls a long
diff.

A changed line's changed words are marked, as delta and GitHub mark
them: in a run of removed lines followed by added ones, each removed
line is set against the added line in its place, both split into words,
runs of space and single other characters, and compared (package `diff`).
A change that only inserts or only deletes slides as far right as the
same words let it, and changes that then touch join, so that one run is
marked where two would do. Lines with too little in common (what changed
is over 6 tenths of their characters besides space) are marked as whole
lines only.

It is a list of its hunks, as a HottyTable is of its rows: it takes
focus when it has one, `selected` holds the selected hunk's id, its
file's name, `:` and the first line of its new side (`api/handler.go:18`;
the line alone without a name), and `onActivate` runs on Enter or a
click on the selected hunk (§3.7). Bind `selected` to a path that the
action's context reads to stage, revert or comment on that hunk. Its
keys are §3.7's. A big patch belongs in props rather than the data
model, where every change to it would be sent again.

### 6.14 HottyTree

Nodes in a hierarchy, one selected at a time, as a file tree or an
outline shows them: lipgloss's tree in cells (§3.4), a host's tree view
on a host (§2). `items` are objects with a `label`, an optional `icon`
(a name as a HottyIcon takes, §6.15, which a host draws), a `value` and
`children`, more such
objects; literal, or best bound to the data model, so that the agent
changes them with `updateDataModel`. A node is identified by its `value`,
as text, or by its place without one (`0.2.1`, its index at each level
from the roots down). `selected` (a DynamicString) is the selected
node's, as a HottyList's is (§6.9).

`expanded` is the ids of the open branches. Bound, the renderer writes it
as the user opens and closes them, so that the agent sees what is open
and opens a branch by writing it; a literal is only where the tree
starts, and the renderer keeps the folding in its own state from then
on. The branches the selected node is in show open whatever `expanded`
says, so that the agent reveals a node by selecting it; moving the
selection then writes them into `expanded`, so that they stay open.
Closing the branch the selection is in selects the branch.

`filter` (a DynamicString) narrows the tree; bind it to a TextField's
value. A node shows when its label holds the filter's characters in
order, ignoring case (sahilm/fuzzy, as a HottyList's filter), when one
inside it does, so that the branches leading to a match show, or when it
is inside one that does; every branch among them shows open, the nodes
keep their order, and the characters matched are marked. Unlike a
HottyList's, the filter is the data model's, as the field holding it is.

Enter, or a click on the selected leaf, runs `onActivate`, whose context
reads the selection from where `selected` is bound; on a branch, Enter
opens or closes it. Its keys are §3.7's. `height` is the rows it shows,
moved as the selection moves (§3.4); without one, it is as tall as the
nodes that show. `emptyText` shows when none does (`Nothing here.` by
default).

### 6.15 HottyIcon

An icon the basic Icon's 59 names lack, by its Material Symbols name,
Sharp and filled as the 59 are (gov R-4), or a path of the agent's own,
filled or stroked. `name` is Material's name in snake_case, as Google
Fonts writes it (`account_tree`, `terminal`), which models know; the 59
are there too by Material's names (`account_circle`, `play_arrow`,
`credit_card`), but for `favoriteOff` and `starOff`, which Material's
filled `favorite` and `star` aren't. Or `name` is `{svgPath, strokeWidth}`:
an `svgPath` as the basic Icon takes one (§2, Icon), stroked `strokeWidth`
wide instead of filled when it has one, as outline packs such as Lucide
draw (2).

Material's whole style is 4,033 icons, 347 KB gzipped, so a program
carries only the names it draws: `make icons` writes them into a package
of its from a list (`cmd/iconsgen -set dir`, `dir/icons.txt`), which
registers them (`icons.Register`); the storybook's are
`storybook/icons.txt`. A name the program hasn't registered draws as one
it has no shape for: `◇`. A package of the whole style, which a program
imports to take any name, is gov R-4's `icons/materialsymbols/`, still to
come.

It is as tall as the text around it and takes its colour, and takes no
focus. In cells and text it is its glyph (§3.4), as an Icon is.

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
