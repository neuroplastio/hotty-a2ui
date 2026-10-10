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
  document, changed by deltas (§2). Its markup also goes into an ordinary
  web page, with no program behind it (§2.1).
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

**The document.** It holds the kit's stylesheet and three elements: the
surface (`~s`); the layer an open Modal shows in (`~o`), empty while
none is open; and the toasts' region (`~t`, §6.23), over both, empty
while there are none. The document asks the network for images over HTTPS only
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
| Row, Column, List | `div`, a flex row or column; List scrolls. A Row wraps, as a page's inline content does, when it holds a field or only inline things (texts, icons, buttons, small pictures); a Row of columns or cards shrinks them instead. A weighted child of a Row takes its share of it, whatever it holds (`flex: weight`, as A2UI's Lit renderer has it), so Rows weighted alike line up as a grid's; in a Column, a weight is a share of the rows to spare. A List whose items move (`reorder`, §6.21) gives the program Alt+ArrowUp and Alt+ArrowDown (`data-keys`), so that they move the item the keyboard is in, as in cells; its drags are cells' alone until KIT-23h |
| Card | `div`, a raised fill and a rounded border |
| Text | `div` with the Markdown as HTML, its emoji in spans; links open in the terminal (`target=_blank`, SPEC.md §9). A word breaks only when it fills a line alone. A Text that is one number (`$850,000,000.00`, a grid's cell) may also break after its group separators, a zero-width space after each, so a narrow column breaks it between groups, not between any two digits. An ordered list item's number is written out, a `span.k-n` the sheet sets in the indent with a gap, ending at the same place for `9.` and `10.`, because Blitz draws a list's own numbers flush against the text and has no counters to draw them otherwise |
| Image | `img`, sized by its variant in `rem`. A weighted one in a Row takes its share of the row whatever its variant, and is as tall as its picture at that width, no taller than the surface: a picture to look at, beside what is said about it |
| Icon | `span role=img`, holding an inline `svg` 1em square whose one `path` is filled with `currentColor`, so it takes the text's colour, the theme's or a Button's. One of the 59 names is its Material Symbols shape (Sharp, filled; `favoriteOff` and `starOff` unfilled), generated into `icons/` by `make icons` (gov R-4). An `svgPath` is drawn as A2UI's reference renderers draw it, in a 24 box, and only when it is path data and nothing else, at most 8 KB (`icons.Path`). The kit writes the `svg` itself, so a path is only ever an attribute. A name with no shape, or a path it doesn't accept, is the glyph cells draw. A HottyIcon is drawn the same way (§6.15); its path, with a `strokeWidth`, is stroked with `currentColor` that wide, with round caps and joins, and not filled |
| Video, AudioPlayer | `a` with `href`: a link the program opens (§7) |
| Divider | `hr`, or a vertical rule |
| Button | `button type=button`, `disabled` while its checks fail. A Row as its label stays a row, so an icon is beside its text. While it has the host's focus it is filled with `--k-focus`, its label in `--k-on-accent`, as cells reverses it in the accent (§3.4) and huh fills the answer a Confirm would give (§6.24): a fill, where the host's own 1px focus ring may drop out on a rounded edge at a fractional scale. A vertical List's Button is its row (`k-item`), as a menu's: as wide as the List, its label at the start, its variant only a fill, so that focusing or picking a row moves nothing; it keeps the ring |
| TextField | `input` (`text`, `password`, `number`) or `textarea`, with `data-on=input`. With suggestions (§6.22), the `input` names (`list`) a `datalist` (`~x`) of the agent's options, which follows them by deltas: a host that draws a datalist, as a browser's DOM does, shows its own list, whose pick comes back as `input`. Blitz draws none (checked with `hotty render`), so on hottyterm the field is plain. The kit's list under the field, its ghost text and its keys are cells' until a list of the kit's own, a surface at a higher z (vault KIT-11h): a host keeps Tab (SPEC.md §10.2 ignores a binding of it) and leaves the value of the field it has focused alone (SPEC.md §6.2) |
| CheckBox | `input type=checkbox` |
| ChoicePicker | one of several shown as checkboxes is a select: a `button aria-haspopup=listbox` with the picked option's label and a chevron in `--k-muted` (Material's `keyboard_arrow_down`, `keyboard_arrow_up` while the list is open, 1.25em), whose list opens in a surface of its own (below); else its options, as checkboxes (several) or chips (`button aria-pressed`). Hosts draw `select` unevenly (Blitz not at all), and a select's list is the program's to place (SPEC.md §9). An option's icon (§6.16) goes before its label: in the list, on a chip or a checkbox, and on the select while the option is picked |
| Slider | `button role=slider` drawing the track (its rail, fill and knob), between `−` and `+` buttons out of the Tab order, then an `output` as wide as the widest value (§3.3's), in `ch`. The track is cut into notches, one a value (at most 41; twenty steps without a `step`), each centred on where the knob stands at its value and half one at either end, so that a tap, a click, sets the value tapped. On a host with `steps` (SPEC.md §4, §9.1) the track is the drag target, `data-on="drag"` and `data-steps` its notches' count less one, and the notches take `click` alone: a drag sets the value of the step under the pointer wherever the pointer goes, off the track and out of the surface too, until it is let go, and the click it ends with changes nothing more. Without `steps` each notch has `data-on="drag click"`: a drag sets the value of the notch under the pointer, and stops where there is none. The track's `touch-action: pan-y` lets a finger drag it: one that moves along the track drags, and one that moves up or down scrolls. Its fill runs from the rail's start to the knob, or, with `fill` `"end"` (§6.16), from the knob to the rail's end. Hosts draw `input type=range` unevenly (Blitz not at all) |
| DateTimeInput | `input type=text` with the ISO 8601 value, its form as the placeholder, as in cells (§3.5). Hosts draw date and time inputs unevenly (Blitz not at all), and none takes an offset such as `Z` |
| Tabs | a `tablist` of `button role=tab`, each its icon (§6.16) and its title, then the tab shown |
| Modal | its trigger; while open, its content in the layer, over a backdrop, the surface `inert` |
| HottyForm | `form` with a hidden submit button out of the Tab order, so Enter submits |
| HottyProgress | `div role=progressbar` with `aria-valuemin`, `aria-valuemax` and, when it has a value, `aria-valuenow`: the label, then a rounded track with a fill as wide as the fraction (in `--k-info`, `--k-success` once full) and an `output` with the percentage. Without a value, a quarter of the track sweeps across it with the clock (§3.4): the element's `--k-at` is where it starts, so that a tick is one attribute's delta |
| HottyTable | `div tabindex=0 role=grid` holding a `table`: the header in `thead`, then in `tbody` the rows the view shows (the same window as cells, §3.4), each a `tr data-on=click` whose id is the table's and `~y` and the row's index, `aria-selected` on the selected one, which is filled (`--k-tonal`, tinted with `--k-focus` while the table has the keyboard). Its `data-keys` give the program the arrows, Page Up, Page Down, Home and End (SPEC.md §10.2, keys for the program), on a host that would scroll with them; Enter reaches the program anyway, a focused box using no keys. While it scrolls, a note under it says which rows show. A column's `width` is for cells: on a host the table lays its columns out. A host's own scrolling, a sticky header and the row under the pointer are vault KIT-01h. A `reorderable` one's `data-keys` also give the program Alt+ArrowUp and Alt+ArrowDown, which move the selected row (§6.21); a drag is vault KIT-23h |
| HottyList | `div tabindex=0 role=listbox`, as a HottyTable's box: its title (or, while its filter is typed, `Filter:`, the text and a caret that shows while it has the keyboard), its status line (`~u`), then the items of the page the view shows (the same page as cells, §3.4), each a `div role=option data-on=click` whose id is the list's and `~i` and the item's index, its label over its description in `--k-muted`, `aria-selected` on the selected one, which is filled with `--k-tonal` and has a 3px bar at its start, in `--k-border`, and in `--k-accent` with its text while the list has the keyboard; a label's characters that matched the filter in `span.k-match`, underlined; then a dot for each page, the page shown's in `--k-fg`, in a row (`~d`, `role=status`) that says which page shows (a HottyPaginator's dots, §6.26). Its `data-keys` give the program what a HottyTable's do, the arrows left and right and Space; the characters its filter types, Backspace, Enter and Escape reach the program anyway, a focused box using no keys, so the filter is typed as in cells. Its empty text shows when it has no items. Two-line rows in proportional type and the item under the pointer are vault KIT-04h. A `reorderable` one's `data-keys` give Alt+ArrowUp and Alt+ArrowDown too (§6.21) |
| HottyKeyHints | `div`: in the short view, a line of hints, each a `kbd` (the key, in `--k-muted`) and what it does (fainter), ` • ` between them, cut where it does not fit; in the full view, its groups (§6.10) side by side, `4ch` apart, each a grid of keys and what they do. The host moves focus among the elements it works itself (fields, boxes, Buttons) without telling the renderer (the keyboard, below), so their keys are left out, where a guess would go wrong at the next Tab; a HottyTable's, a HottyList's, a HottyDiff's, a HottyTree's, a HottyPaginator's, a Slider's and a select's show, as the program's keys go there. `?`, which a focused field types, reaches the program from anywhere else and switches the views. Keycaps are vault KIT-08h. On a surface with a component that has an `accessibility.description`, a row (`~z`, `k-hints-tip`) comes first: the tooltip (§6.23), the description of the element the host says the pointer is over, else of the element with the keyboard, in the text's colour and italics, cut with an ellipsis; empty, a line high, when neither has one |
| a toast (`hottyToast`, §6.23) | in the toasts' region (`~t`, `div role=region aria-label=Notifications`), fixed at the surface's top right corner, a column `min(22rem, 100% − 1rem)` wide, the newest at the top: a `div data-on=click` whose id is `hottyToast:` and the toast's id, `role=alert` for a warning or an error, which a screen reader reads at once, `role=status` for news and a success, read when it is idle. It holds the kind's mark, the basic catalog's icon (`info`, `check`, `warning`, `error`) in the kind's colour (`--k-info`, `--k-success`, `--k-warning`, `--k-error`), `aria-hidden`; the message; and its action, a borderless `button` whose id is the toast's and `/action/0`. It is filled with a tenth of its kind's colour over `--k-surface`, ringed with it, rounded and shadowed. The region is cut at the surface's edges: a toast as a surface of its own at a higher z (SPEC.md §5.2) is vault KIT-13h |
| HottyScrollView | `div tabindex=0`, a box `--k-rows` terminal rows tall (its `height`, by SPEC.md §8's `--hotty-cell-h`), filled with `--k-tonal`, a ring while focused, whose content overflows it: the host scrolls it, with its own scrollbar, the wheel, a touch drag and the keys a browser scrolls with (SPEC.md §5.3; the storybook places its surfaces with `scroll`). Its child goes inside as itself; its lines are a `div` (`~j`) of a row each, in the mono face and `white-space: pre`, so that the box shows `height` of them, cut at the box unless it wraps them. A log's (`follow`) box is `role=log`, which a screen reader reads as lines arrive, and a line written to the next index arrives as an `append` delta. A program cannot set where a host has scrolled, so the box starts at its top, `follow` or not, and `hottyScrollTo` does nothing there. Its `data-keys` binds the keys bubbles' viewport scrolls with to the host's scroll actions (SPEC.md §10.2, *Scrolling keys*), so that they scroll it as in cells (§3.7): j and k, f, b, Space and Shift+Space, u and d (Control+u, Control+d), g and G, and h and l while its lines are cut. A Button in it keeps Space and a field in it types the letters, since a key an element uses stays its own and a field leaves scroll actions out. Following the tail on a host is vault KIT-07h |
| HottyCode | `div` in the mono face, filled with `--k-tonal` as a Text's code block is, a flex row (`k-cl`) for each line: its number, right-aligned as wide as the widest (`--k-ln`), in `--k-muted` and `aria-hidden`; its mark's sign, when the code has marks; then its code (`k-src`), a span for each token that is not plain, of class `k-t-` and its kind, which the sheet colours with the roles cells uses (§3.4). A marked row has `k-m-` and its kind: a highlighted one is filled with `--k-selection`, the others with a sixth of their role's colour (`color-mix`). Lines wrap, `pre-wrap`; with `wrap` false they do not, and the box scrolls sideways, which the host does itself (SPEC.md §5.3), every row as wide as the widest so that a tint reaches the end. A Text's fenced code block is goldmark's `pre` and `code`, its tokens in the same spans |
| HottyDiff | `div` in the mono face, filled with `--k-tonal` as a HottyCode's; with hunks, `tabindex=0 role=listbox`, whose `data-keys` gives the program ArrowUp, ArrowDown, Home, End, k, j, g and G (a HottyScrollView around it binds the letters to scroll actions, SPEC.md §10.2). For each file with a name (or of several), a row (`k-diff-file`): its name, bold, then what the change adds and removes, `+N` in `--k-success` and `-M` in `--k-error`, or new, deleted or binary. A run of unchanged lines left out is a muted row, `⋯ N unchanged lines` (`k-fold`). Each hunk is a `div tabindex=-1 role=option data-on=click` (`k-hunk`), whose id is the diff's, `~b` and the hunk's index: its header in info and its section muted (`k-hh`), then its lines. Unified, a line is a flex row (`k-dl`): its old number and its new (`--k-lo` and `--k-ln` wide, `aria-hidden`), its sign, then its code, its tokens in a HottyCode's spans. Split, the hunk is a grid of two equal columns, the old side and the new (`k-ds`), paired as cells pairs them, a rule between them and the header across both. A removed line (`k-d-del`) is filled with a sixth of `--k-error`, an added one (`k-d-add`) with a sixth of `--k-success`, and its changed words (`k-w`) more. The selected hunk (`k-sel`, `aria-selected`) has a rail on its left and its header reversed, muted, and in `--k-focus` while the diff has the keyboard (`k-on`). It then has the host's focus itself, so that the host scrolls it into view (SPEC.md §5.3) as cells keeps it in sight. A split diff stays split at any width, its sides wrapping: the markup does not know the width, and a host has no container queries |
| HottyTree | `div tabindex=0 role=tree`, as a HottyList's box, holding the nodes the view shows (the same rows as cells, §3.4), each a `div role=treeitem data-on=click` whose id is the tree's and `~q` and the node's index, with `aria-level`, `aria-expanded` on a branch and `aria-selected`. A row is a flex row indented `--k-5` a level (`--k-level`): its fold (`k-node-fold`, a column wide), `▸` or `▾` in `--k-muted` before a branch and blank before a leaf, so that a level's labels line up, and none in a tree with no branches (a docs nav's pages); its icon, as an Icon draws it, or a blank one as wide where another node has an icon; then its label, the filter's matches in `span.k-match`, and a closed branch's count (`k-node-count`, as cells', §3.4) in `--k-muted`. The selected node is marked as a HottyList's selected item: filled with `--k-tonal`, a 3px bar at its start in `--k-border`, and in `--k-focus` with its text while the tree has the keyboard (`k-on`). Its `data-keys` are a HottyList's; its letters and Enter reach the program anyway, a focused box using no keys. Each row is `--k-node-h` (2rem) tall. With a `height`, the box is that many rows tall and holds every node that shows, and the host scrolls it (SPEC.md §5.3), with the wheel and a thin scrollbar; the selected node is `tabindex=-1` and has the host's focus while the tree has the keyboard, so that the host scrolls it into view as the selection moves, as a HottyDiff's selected hunk does. Guides are cells' (§3.4): a host has the room to indent instead. Its empty text shows when no node does. A `reorderable` one's `data-keys` give the program Alt with each of the four arrows too (§6.21) |
| HottySwitch | `button type=button role=switch` with `aria-checked`, `disabled` while it is, holding its pill (`k-switch-track`, `aria-hidden`) and its label (`~l`), so that a click on either flips it, as do Space and Enter, which a host clicks a button with (SPEC.md §10.2); then its error. On or off is that one attribute, so a flip is one delta. Off, the pill is filled with `--k-tonal` and ringed in `--k-muted`, its knob `--k-muted` at its start; on, it is `--k-accent`, its knob `--k-on-accent` at its end, so that the two differ by more than a shade where the accent is the foreground (no theme), as Material's switch does. The ring, and a focus ring of `--k-focus` around the pill while it has the host's focus, are shadows, which draw where a thin rounded border may not. The knob does not slide: Blitz runs no transitions, and a slide by deltas is vault KIT-22h |
| HottyRangeSlider | a Slider's field (§6.20): its label (`~l`), then a row (`~r`) of the track and an `output` as wide as the widest range, in `ch`, naming both knobs (`for`); no `−` or `+`. The track is `div role=group`, named by the label (`aria-labelledby`), holding the rail, the fill from the start's knob to the end's, a Slider's notches, and the two knobs, each `button type=button role=slider` (`k-thumb`, the knob's view id, `price/knob/0` and `price/knob/1`) named `Price, start` and `Price, end` (`aria-label`; `Start` and `End` without a label), with `aria-valuenow` and its own range up to the other knob in `aria-valuemin` and `aria-valuemax`, as a multi-thumb slider has it. Each knob is a Tab stop, a button, whose arrows, Home and End the host leaves to the program (SPEC.md §10.2). On a host with `steps` the track is the drag target, as a Slider's, the notches take `click` alone, and the knobs nothing: a press anywhere on the track, a knob too, moves the nearer knob to the step under the pointer and gives it the keyboard, and the drag moves it on. The press's blur, which the host sends as the track takes no focus (SPEC.md §10.1), leaves the keyboard on the knob, and the next update gives the host it back (`a=focus`). Without `steps` each knob and each notch has `data-on="drag"` (the notches `drag click`): a drag from a knob moves it to the notch under the pointer. A tap on a notch moves the nearer knob there. A knob is a disc, `--k-accent`, ringed by a 2px outline of `--k-focus` while it has the host's focus (an outline, which Blitz draws round, where it draws a spread shadow's corners square). A disabled one's knobs are `disabled`, its notches take nothing, and the field is faint (`k-off`), its label and value too |
| HottyChart | `div role=img`, named by its `accessibility.label`, else `Line chart` or `Bar chart` and its series' labels; a grid of its ticks' labels and its plot. The labels (`k-chart-y`, `aria-hidden`, in `--k-muted`) are each placed at its value's height, all of them, in a column as wide as the widest, in `ch`. The plot (`~a`) is `--k-rows` terminal rows tall (its `height`, by SPEC.md §8's `--hotty-cell-h`), its left and bottom borders the axes, 2px in `--k-border`, since a 1px straight edge drops out at a fractional scale. The axis is the view's, cells' (§6.18), so a value sits at the same height in both. A line is an inline `svg` a series (`~g0`, `~g1`, …), stretched over the plot (`preserveAspectRatio=none`), holding one `path` (`~c0`, …) that hotty-go's `chart.Line` makes: a break where a value is missing, a lone value a dot, round joins, stroked 1.5 wide in `currentColor` with presentation attributes only, which every host's SVG takes. A new point is so an attribute's delta a line. Bars are boxes, a group a point and a bar a series in it, each from the axis's 0 to its value, placed by percentages, so that one below 0 hangs from it. The series' colours are cells' (§3.4), set by the series' class: `--k-info`, `--k-warning`, `--k-success`, `--k-error`, and round again. Under the plot: a line's points' labels, three evenly spread (the first and the last among them), or a bar chart's each under its bars; then the legend, when it has several series or a labelled one, a key a series (a stroke for a line, a square for bars) and its label in `--k-muted`. With no values, the plot says `No data`. The `svg`'s box is 480 by 20 a row, a guess at the plot's width that the markup cannot know: on a plot much wider or narrower the stroke thickens where the line is steep (vault KIT-10h) |
| HottySparkline | `span role=img`, named by its `accessibility.label`, else `Sparkline`: a box `--k-n` columns wide (its `window`, else its values' count, at most the room), `--k-rows` terminal rows tall, in `--k-info`, holding a bar a value, a terminal column wide (`1ch` of the mono face), from its bottom, as tall as its value is between its `min` and `max`, at least an eighth of a row; a missing value has no height. The bars come newest first in a row that runs right to left (`row-reverse`), so that the newest is at the box's right edge and a box too narrow loses the oldest, as in cells |
| HottyPaginator | `div tabindex=0 role=group` (`k-pager`), named by its `accessibility.label`, else `Pages`: its child, which holds the page the view shows (the same page as cells, §3.4), then its pages: a dot for each (`span.k-dot`), each a `data-on=click` whose id is the paginator's and `~d` and the page's index from 0, the page shown's in `--k-fg` and the others in `--k-border`, in a row (`~d`, `role=status`) that says `Page 3 of 10`; with `displayStyle` `numbers`, that row holds `3/10` in tabular figures. Its content starts past a 3px bar at its start, where a HottyList's items do, in `--k-focus` while it has the host's focus, the page shown too. Its `data-keys` give the program the arrows left and right, Page Up, Page Down, Home and End, on a host that would scroll with them; h and l reach the program anyway, a focused box using no keys. A page is as tall as its items, a host laying it out, where cells keeps every page as tall as the tallest. Without a child it is the row alone (`k-bare`). The host's own look is vault KIT-15h |
| HottySpinner | `span role=status`: the frame of its set the clock is at (§3.4) while it spins (blank while it does not), `aria-hidden`, in the mono face and as wide as the set's widest frame in cells (in `ch`), then the label, which so stays put as in cells. The frame's id is the Spinner's and `~f`, so that a tick is one text's delta |
| HottyTimer, HottyStopwatch | `span role=timer`, which a screen reader does not read out at each tick: its label, then its time as cells writes it (§6.25), in a span whose id is the timer's and `~f`, so that a tick is one text's delta, in tabular figures (`font-variant-numeric`), so that it does not jitter as it ticks, and in `--k-muted` while it stands still (`k-still`, one attribute's delta). The program's draws move it on the clock (§3.4), as a Spinner's frames: a host runs nothing of its own here. The time in display type is vault KIT-14h |
| HottyBigText | `div` (`k-big`) holding its text as it is, its lines apart at a `br`, in display type: weight 800, tabular figures, as many terminal rows high (`--hotty-cell-h`) as cells' letters, 3, 4 or 5 by its size (`k-big-small`, `k-big-medium`, `k-big-large`), a line one row more, which leaves room for descenders; `k-big-center` or `k-big-end` aligns its lines. A line too long wraps as text does. A screen reader reads the text. The host's own look is vault KIT-19h |

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
| `input`, `change` | writes the control's value: to its bound path, else as the renderer's; a text field's new value then runs its suggestions' `onInput` (§6.22) |
| `click` | a Button runs its action; a HottySwitch flips (a disabled one does not); a Table's row or a HottyList's item is selected, or acted on once it is; a tab is shown; a chip toggles; a link opens; a Modal's trigger opens it; the backdrop closes it; a Slider's `−` or `+` steps it, and a tap on its track sets it; a tap on a HottyRangeSlider's track moves its nearer knob there, and a click on a knob gives it the keyboard; a click on a HottyPaginator's dot shows its page, and one anywhere on it gives it the keyboard; a select opens or closes its list, an option in the list is picked, and a click beside the list closes it; a click on a toast dismisses it, and one on its action sends its event and dismisses it (§6.23) |
| `submit` | the HottyForm submits (§6.2) |
| `dragstart`, `drag` | a Slider takes the value of the notch the pointer is on (SPEC.md §9.1); `dragend` and a drag off the track leave it. A HottyRangeSlider's dragstart picks the knob that moves (§6.20), and each drag moves it, stopped where it meets the other. The click a drag ends with, which comes right after its dragend and only when it was let go on the slider, changes nothing more |
| `focus`, `blur` | the surface has the keyboard, or not |
| `hover` | on a surface placed with `v=1` (SPEC.md §9.4; the storybook asks where the host lists `hover`): the element under the pointer, whose description shows as the tooltip, and a toast under it waits (§6.23); out, neither. A key, a click or a focus the user moves hides the tooltip until the pointer goes to another element, as in cells |

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
(§6.3) names, `a=blur` for `hottyBlur`. When the user moves focus within
the surface, by Tab or a click, the host sends `focus` naming the element
(SPEC.md §10.1, `t` the nearest id from it outward), and the renderer
takes that element as focused. Typing (`input`) and a control's `change`,
which comes at once, say so too. A text field's `change` is its commit as
the host's focus leaves it, often for where the renderer just moved the
keyboard, so it says nothing about where the keyboard is.

### 2.1 Pages

The same markup also goes into an ordinary web page, not a terminal: page
mode, for a docs site's plain page (gov NEIO-14), which readers without
JavaScript and crawlers read. No program is behind it, so nothing on it
may need one to be read: what only a host's events and keys make work is
left out, and what a program would show a piece at a time shows whole.
The reference is `rendition/html`'s `PageCSS` and `Rendition.Page`, on a
view built for a page (`view.BuildPage`); `storybook -page` writes
stories, or a stream, as one such page.

**The stylesheet goes in once** a page, however many surfaces it shows.
`PageCSS()` is the kit's sheet, then page mode's rules, for a `<style>` in
the page's head or a file it links. `Page(o)` is a surface's markup alone,
a fragment for the page's body: no head, no sheet. `o` (`PageOptions`)
is what the page knows of itself that the kit can't: `Heading`, the level
of the heading a Tabs' titles are in its outline (3 under a `##`
section; 0, titles that are no heading), and `Origin`, its own site
(`https://kubecom.neuroplast.io`), whose absolute links stay in the tab
as relative ones do. The rendition's name
is its top element's id (`k-surface k-page`) and prefixes every id in it
and every attribute that names one (`for`, `aria-labelledby`,
`aria-describedby`): `<name>~~<id>`, `~~` being no element's own (§2), so
that two surfaces of the same components share no id.

**The colours are the page's.** It sets the host's palette (SPEC.md §8's
names) on `<html>`, a theme's class, or any element around its surfaces;
the kit's variables follow it on the surface itself, so a box around one
surface may set its own. What the page leaves out is the kit's default,
light or dark as the page's `color-scheme` says (`light dark` follows
the reader):

| the page sets | the kit's | what it colours | default |
| --- | --- | --- | --- |
| `--hotty-bg` | `--k-bg`, `--k-on-accent` | what the fills mix into; text on a primary Button | `Canvas`, the page's own |
| `--hotty-fg` | `--k-fg`, `--k-accent` | text; a primary Button, a tab's underline (NEIO-4) | `CanvasText` |
| `--hotty-accent` | `--k-link`, `--k-focus` | links, focus | `#0969da` light, `#4493f8` dark |
| `--hotty-ansi-8` | `--k-border` | rules, a selection's bar | `GrayText` |
| `--hotty-ansi-9`, `-3`, `-2` | `--k-error`, `--k-warning`, `--k-success` | errors, signals, a diff | `#cf222e`, `#9a6700`, `#1a7f37` light; `#f85149`, `#d29922`, `#3fb950` dark |
| `--hotty-ansi-6` | info | a Progress, code's keywords | the link colour |

`--k-muted`, the fills and the rules are made from these, as on a host. A
theme set on the rendition (`SetTheme`) wins, and its surface is a box of
its colours, padded and rounded (`k-themed`). The type is the page's: the
surface inherits its font, and spaces in `rem`. A page's own rules for
bare elements (`a`, `ul`, `h1`) reach into its surfaces; it scopes them.

**What a page shows**, component by component; anything not listed is as
on a host:

| component | on a page |
| --- | --- |
| Tabs | every tab, in order, each a `section` named by its title (`aria-labelledby`): the title drawn as a tab bar of one tab, the tab shown's look, then its content, as GitHub shows the same Markdown. The title is a heading of the level the page gives (`PageOptions.Heading`), still in the tab's look, else a `div`. A host's tabs, one at a time, are a program's |
| HottyTree | every node a row, the host's row (fold, icon, label, a closed branch's count), in `ul`s (`role=list`) that nest as the nodes do. A branch is a `details`, open or closed as given (`expanded`, and the branches the selection is in), whose `summary` is its row: the browser opens and closes it with no program, and the sheet draws its fold, `▸` or `▾`, and its count while it is closed. The selected node is marked as on a host, its label `aria-current`. A node with an `href` has an `<a>` for its label, `aria-current=page` when selected (vault KIT-26 adds the `href`). A filter is a program's: a page draws every node |
| Text, Video, AudioPlayer | links are `<a href>`: one that leaves the site (`http`, `https`, `//host`, to a host not the page's `Origin`) opens in a new tab (`target=_blank rel=noopener`), any other (a path, a fragment, a query, the page's own site, `mailto:`, `tel:`) in the same one, and one to any other scheme loses its `href`. The host's `target=_blank` (§2) goes. One function decides this for every link on a page |
| TextField, DateTimeInput | the value, `readonly`: it can be selected and copied, not changed. Suggestions' `datalist` goes |
| CheckBox, HottySwitch, Slider, HottyRangeSlider, ChoicePicker, Button | as on a host, their state shown (a select's picked option, closed; a chip pressed or not), taking no click (`pointer-events: none`, a checkbox's label too) and out of the Tab order (`tabindex=-1`): no program would hear them |
| HottyForm | a `div role=form`: a submit would leave the page |
| Modal | closed: its trigger alone |
| HottyTable | every row, not a window of them; the selected one marked, `aria-current` |
| HottyList | its title, its status line and every item; no filter, no dots |
| HottyPaginator | its child with every page's items, one after another, and no dots; a bare one, whose pages are the agent's, says which page shows, as on a host |
| HottyKeyHints, toasts | left out: no key reaches a program, and no notice comes |
| HottyProgress | an indeterminate one sweeps by a CSS animation, 5 s across as on the clock (§3.4); still under `prefers-reduced-motion` |
| HottySpinner, HottyTimer, HottyStopwatch | the frame and the time the rendition's clock was at, still |

Everywhere, the host's `data-on`, `data-keys` and `data-steps` go, and so
do what only a program's keyboard makes true: a box's `tabindex`, the
`grid`, `listbox` and `option` roles of boxes a program selects in (their
selected row `aria-current`), an open list's `aria-controls`,
`aria-activedescendant`, `aria-haspopup` and `aria-expanded`.

A page host, the addon's surface code (DOM, events, deltas) without the
terminal, would make the same HTML live from a WebAssembly program
running the kit: progressive enhancement, which nothing on the page needs
to be read. It is not part of this.

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
| Button | the label + 2, a column of padding each side, a List's too; borderless, the label |
| TextField, DateTime | 2 for the gutter, the label and a space when it has one (padded to its run's widest label, §3.5), the inset (1) and its input: a text field's 32, as wide as an email address takes, whatever is typed; a DateTime's, its value's form (`YYYY-MM-DDTHH:MM`, `YYYY-MM-DD` or `HH:MM`) or its value, whichever is wider, and a column for the caret. A longText, the label's or 20 (an HTML textarea's size) |
| CheckBox | 2 for the gutter, 3 for the box, plus 1 and the label when it has one |
| HottySwitch | 2 for the gutter, 3 for the switch, plus 1 and the label when it has one |
| a select | a one-line text field's, its input the widest option's label, a space and `▾` |
| a Choice's options | 2 for the gutter, plus the label's width or the widest option row (`> [ ] label`, or `> ( ) label` for chips that pick one), whichever is wider |
| Slider | 2 for the gutter, the label + 1 if it has one, a track of 10, 1, and the value's width: the widest of min, max, the value, and a value on a step (its ends' whole part, a point and the step's decimals), so that the track keeps its length as the value moves |
| HottyRangeSlider | a Slider's, its value's width that of the widest range: twice a Slider's value's width, and 1 for the `–` |
| HottyProgress | its label's width or 25 (a bar of 20 and ` 100%`), whichever is wider |
| HottySpinner | its set's widest frame, plus 1 and the label when it has one |
| HottyTable | its columns' widths, each a column's `width`, else its header's or its widest cell's, whichever is wider, plus 2 for each column (a column of padding each side), and 2 for the scrollbar while its body scrolls |
| HottyList | 2 for the indent, plus its widest line: its title + 2, its status line, its empty text, `Filter: ` + 1 when it filters, each item's label and description |
| HottyKeyHints | its short line, or in the full view its columns, uncut; not its tooltip's row (§6.23), which is cut to the rest, so that nothing moves as it changes |
| HottyScrollView | its widest line unwrapped (whether it wraps them or not), or its child's natural width, plus 1 for the padding and 2 for the scrollbar |
| HottyCode | its gutter (§3.4) and its widest line |
| HottyDiff | 1 for the rail, then its gutter (§3.4) and its widest line; split, twice a side's gutter and that line, and 3; or a file's row, a hunk's header or a fold, if one is wider |
| HottyTree | 2 for the bar, plus its widest node, whether it shows or not, so that it keeps its width as branches open and close: its guides (4 a level below the roots), its fold (2), its label and, on a branch, a space and its count; or its empty text |
| HottyChart | its widest tick label, 2 for the axis and a space, and its plot or its legend, whichever is wider: a line's plot is 40; bars', a group a point, as wide as its widest label or 2 a series, one column between groups (at least 8, and `No data`) |
| HottySparkline | its `window`, else its values' count |
| HottyPaginator | 2 for the gutter, plus its widest page's natural width (its child as it stands for each page) or its dots, a column a page, whichever is wider; with `numbers`, its last page's of its pages (`10/10`), so that it stays put as the pages turn |
| HottyBigText | its widest line unwrapped, in its letters (§3.4) |
| Image, Icon, HottyIcon, Media, Placeholder | what they paint (§3.4) |
| Divider | 1 |
| Card | its content + 4 |
| Row | its children's, plus one column between each two |
| Column, HottyForm, Modal | its widest child's |
| Tabs | its titles two columns apart, or its content, whichever is wider |

Minimums:

- A Text's minimum is its longest word. A list item's indent and marker
  count with its first word.
- A one-line text field's and a DateTime's is the gutter, its label and a
  space when it has one, the inset and 3 columns of input, which scrolls;
  a longText's, the gutter and its label's longest word, and at least 3.
- A Slider's is the gutter, 4 and its value's width; a HottyRangeSlider's
  too, its value's width a range's.
- A Choice's options' is the gutter and its widest option, or its label's
  longest word if that is wider.
- A HottyProgress's is its label's longest word, and at least 8: a bar of 3
  and the percentage.
- A HottyTable's is its columns', each at most 3 wide, with their padding,
  and its scrollbar while it scrolls.
- A HottyList's is its indent and 4.
- A HottyKeyHints's is 1: it cuts what does not fit (§3.4).
- A HottyScrollView's is its child's, or 1 for its lines (it cuts them or
  wraps them), plus 1 for the padding and 2 for the scrollbar.
- A HottyCode's is its gutter and 8 columns of code, or its widest line
  when that is narrower.
- A HottyDiff's is its rail, its unified gutter and 8 columns of code.
- A HottyTree's is its bar, a level of guides and 4.
- A HottyChart's is its tick labels, the axis and a space, and 8.
- A HottySparkline's is 1: it shows the newest values that fit (§3.4).
- A HottyPaginator's is the gutter, and its pages' widest minimum or its
  numbers (`10/10`), which its dots become where they do not fit.
- A HottyBigText's is its widest word; narrower, a word breaks between
  its letters (§3.4).
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
    HottyScrollView; a HottyPaginator is a field) when one is a field and
    the other is not, as bubbles
    sets a form's button apart, or when neither is a field and either has
    a title row (§3.4; a HottyProgress's label is one, a HottyTable's
    header, and a HottyList's status line). Fields follow one another with
    no blank row, as a GUI form's do: an input's underline (§3.5) parts
    one from the next. A stack of Buttons or of HottySpinners stays tight
    too. A control after a Text whose last block is a heading has a blank
    row before it, as huh sets a form's title apart from its fields; a
    control after other Text has none, so that a card's fields stay with
    its text. A HottyKeyHints has a blank row before it,
    whatever comes before, as bubbles' help sits a row under its list.
    A HottyScrollView has one after it, whatever comes after: its box
    draws no edge but the scrollbar, so what follows would read as its
    content. So do a HottyCode and a HottyDiff, which draw no edge at all,
    and a HottyChart, whose legend or labels would read as a title of what
    follows, and a HottyPaginator, whose dots end its pages. Before it, the rule for controls applies, so a title above
    it stays on the row before its box.
    A HottyBigText that has text has one after it too, as its own lines
    are a row apart, so that its letters do not meet what follows.
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
| Button | ` label ` on one row, a column of padding each side, filled as huh's buttons are: a grey a step or two off the background, a tint 48/255 of the way from `bg` to `fg` (§3.6). The label is the plain text of the Button's Texts and the glyphs of its Icons, a space apart. Primary is bold, and otherwise as the others, since `accent` is focus's. While it has the keyboard it is reversed in `accent`, unfilled. Disabled, its label is `fg` and faint on half the fill (24/255): `muted` would vanish at the floor, where the fill is `muted`'s bright black too. Borderless, it is its label alone, underlined and unfilled. Under NO_COLOR, where nothing fills, it is underlined, padding too, so that it still reads as a control, and a Row's column between two keeps them apart; brackets would make it wider there than in colour, and every output mode draws one layout. A vertical List's Button is its row: ` label `, unfilled, its style (focus's reverse, say) across the List; bold unless borderless, and never underlined. |
| TextField, DateTime | in the field's box past the gutter: a one-line field's label and its input on one row (§3.5), a longText's title row and its rows; then the error. A text field's suggestions go over what is under it (§3.5) |
| CheckBox | past the gutter, `[•] label` or `[ ] label`, its box reversed in `accent` while it has the keyboard; then the error |
| HottySwitch | past the gutter, its track with the knob at one end, then a space and its label, on one row: `▬▬■` in `accent` while it is on, a bar into a filled knob, and `□⎯⎯` in `muted` while it is off, a hollow knob on a thin line (`⎯`, not `─`: a terminal draws `─` itself at the cell's middle, where the font's square need not sit, and a font draws `⎯`, often a symbol font by fallback, since few coding fonts have it; a Slider's thin track is the same), so that without colour the knob's end, its fill and the track's weight still say which; then the error. Its label is in `accent` while it has the keyboard. A disabled one is `muted` and faint throughout, its knob still at its end. The switch and its label take a click |
| a select (one value, `checkbox` display) | past the gutter, its label and its input on one row, as a one-line text field's (§3.5): the picked option's label, or `…` in `muted` when none is picked, and `▾` in `muted` at the input's end, a column after the widest option's label, so that the chevron stays put as the pick changes. While its list is open (§3.7), the options follow one a row under the input, their labels under its value's: `● label` for the picked one and `○ label` for the others, each after `  `, or after `> ` in `accent` on the highlighted row, whose label is `accent` and bold. Then the error. |
| a Choice's options (several values, or `chips`) | past the gutter, a title row, then the options one a row, each after `  `, or after `> ` in `accent` for the option with the keyboard (huh's multiselect): `[•] label` or `[ ] label` where several may be picked, and `(•) label` or `( ) label` for chips that pick one, as a GUI's radio buttons; pills wrapped across the field read as neither a form nor a list. Then the error. |
| Slider | past the gutter, `label ━━━━■⎯⎯⎯⎯ 50` on one row: the label as a title and a space, then the track, a space and the value. The track fills the columns left. Up to the knob it is `━`, the knob is `■` at round((value − min) / (max − min) × (track − 1)), and after it the track is `⎯` in `border`. With `fill` `"end"` (§6.16) the two sides swap, `⎯⎯⎯⎯■━━━━ 50`: `⎯` in `border` up to the knob and `━` after it, the glyphs the same. When there is no label, or the track would be shorter than 3, the label is dropped. Then the error. |
| HottyRangeSlider | a Slider's row with two knobs, `label ⎯⎯■━━━━■⎯⎯ 20–70`: the start's and the end's, each `■` where a Slider's would stand at its value, `━` between them and `⎯` in `border` either side, then the range, `start–end`. Where both would take one column, the end's goes a column right (the start's a column left, at the track's end), so that both show, `■■`. A disabled one is `muted` and faint throughout. Then the error. |
| HottyProgress | its label on a row, when it has one, then the bar and the percentage on the next: ` 42%`, five columns (`%3.0f%%` after a space, half to even). The bar takes the columns before them, filled in eighths of a cell: `█` for each full cell, then one of `▏▎▍▌▋▊▉` for a cell part filled, blending across the bar (§3.6): where the theme leaves `accent` to the terminal and the terminal has said its accent (OSC 4;12) and its bright magenta (4;13, else 4;5), from the one into the other, blue into pink in most palettes, as bubbles' default gradient goes (magenta is no role, and nothing else uses it); otherwise from `info` into `accent`. The blend spans the whole bar, wherever the fill ends, so a full bar shows all of it (the percentage says it is done). The rest is `░` in `border`. Without a value, a segment of a quarter of the bar (at least one cell) is `█` in `info`, moving with the clock (below), and the percentage's columns are blank. When the bar would be shorter than 3, the percentage is dropped. |
| HottyTable | a header row, bold; a rule of `─` in `border`; then a row for each row the body shows, a cell for each column: its text, padded a column each side, cut with `…` when wider than its column, at the start, the centre (rounded down) or the end as `align` says. The columns take their natural widths; while the table is too wide, a column at a time comes off the widest that is wider than 3 (the first of equals), and once none is, off the widest. The body shows `height` rows, or all of them (one, `No rows` in `muted`, when it has none): from the first it showed, moved as little as brings the selected row into view, never past the last (the view's `Top`, the same in both renditions). While it scrolls, the rule ends with ` 4–10 of 12 ` in `muted` and one more `─`. The selected row is reversed across the table, in `accent` while the table has the keyboard and in `muted` otherwise. While it scrolls, a HottyScrollView's scrollbar runs down the body after a blank column, past the last column's padding, the rule running over both, so that the counter says where it is and the bar how much more there is either way. |
| HottyList | as bubbles' list with its default delegate, every line 2 columns in: its title, ` Title ` in `accent` reversed, and a blank row (neither without a title); its status line in `muted` and a blank row; then the items of the page that shows, each its label and, under it, its description in `muted` (one row an item, with no blank rows between, when none has a description; else two, a blank row between). The selected item's first two columns are `│ `: in `accent` with its label and description while the list has the keyboard, and in `muted` otherwise, its label bold, as a HottyTree's (the maintainer, round 6). A label too wide is cut with `…`; its characters that matched the filter are underlined. With a `height`, a page shows that many items, from the page that holds the selected one, and the body keeps their rows on every page; after it, while the items take more than a page, a blank row and a dot for each page (a HottyPaginator's dots: `•`, the page shown's in `fg`, the others in `border` and faint; `3/10` in `muted` when the dots do not fit), whose two rows stay while a filter leaves one page. The status line is `12 items` (`1 item`, `No items`); while a filter applies, `“query” 7 items`, or `Nothing matched`, then ` • 5 filtered`. While the filter is typed, `Filter: ` in `accent` and the text take the title's row, with the cursor after them. With no items, the empty text shows in `muted` in the body. |
| HottyKeyHints | as bubbles' help draws it: the short view one line, each key in `muted`, a space, what it does in `muted` faint, and ` • ` in `border` faint between them (bubbles' three steps; the Terminal theme's `muted` and `border` are one colour); the hints that do not fit go, and ` …` ends the line where it fits. The full view (after `?`, §3.7) is its groups (§6.10) as columns four apart, each a row a key, the keys padded to the widest, then a space and what they do; a group that does not fit goes, and ` …` follows the first row where it fits. On a surface with a component that has an `accessibility.description`, a row comes before the keys: the tooltip (§6.23), the description of the element under the pointer, else of the element with the keyboard, in the text's colour and italic, cut with `…` to the keys' width; blank when neither has one, so that nothing moves. |
| a toast (`hottyToast`, §6.23) | a box with rounded corners, its border in its kind's colour (`info`, `success`, `warning`, `error`), over the frame's top right corner a column from its edge (the column blank), the toasts stacked down from the top, the newest first, all as wide as the widest needs (its mark, its message, and its action, 2 columns apart, and 4 for the borders and padding), at most 40 and the frame: a column of padding each side, then on its first row its kind's mark, the glyph of its icon (`ⓘ`, `✓`, `!`, `✗`, the Icon glyphs of `info`, `check`, `warning` and `error`, §3.6), bold in its kind's colour, and a space; then its message, wrapped, continuation rows indented 2 under the message; then its action's label at the end of the last row, bold and underlined (reversed in `accent` while it has the keyboard), or on a row of its own when it does not fit there. A toast is drawn over what is under it, as an open Modal's panel is: it hides it and takes its clicks (§3.7), and a field's cursor under it is not shown. The frame is at least as tall as the stack |
| HottyScrollView | a box `height` rows tall, as bubbles' viewport: a column of padding, its content, 3 columns narrower than the box, then a blank column and the scrollbar down the last. The box is filled with `surface` (a background of `surface` cells, §3.6), so that it stands apart from what is around it; a cell its content tints keeps its own tint. Where the theme does not colour `surface`, there is no fill, and the padding alone sets it off. Its lines are a row each, cut at the box and scrolled sideways six columns at a time, or with `wrap` broken between characters into the rows they take; a child is laid out at the content's width, and shows through the box at the rows scrolled to: what it hides takes no click, and the element with the keyboard inside it is scrolled into sight. It starts at its top, or with `follow` at its end, where it stays as lines arrive until the user scrolls up, and follows again once back at the end. The scrollbar, while the content is taller than the box, is a track of `│` in `border` faint and on it a thumb of `┃` as long as the share that shows (at least a row), where it shows, in `muted`, and in `accent` while the box has the keyboard. bubbles' viewport draws no scrollbar: without one, a box with no edge gives no sign that it scrolls |
| HottyCode | a row for each line, or with `wrap` (the default) as many as the line takes, broken by cluster under the code: first its number, right-aligned as wide as the widest, in `muted`, and a space, when `lineNumbers` is on; then, when the code has marks, the line's sign and a space (`▎` in `info` for highlight, `+` in `success` for added, `-` and `✗` in `error` for removed and error, `!` in `warning`); then the code. Continuation rows leave the gutter blank. Without `wrap` a long line is cut with `…`, and the wheel scrolls the code sideways, six columns a notch (§3.7), the gutter staying put: a `…` then stands in the code's first column too, while columns before it are hidden, and the widest line's end is as far as it goes. Each token is coloured by its kind (package `highlight`): keywords `info` and bold, types and builtins `info`, functions' names bold, strings `success`, numbers and constants `warning`, as are preprocessor lines, decorators and attributes, comments `muted` and italic, a diff's added and removed lines `success` and `error`, its hunk headers `info` and bold, and the rest `fg`. Not `accent`, which marks only focus (§3.6). A marked line's rows are tinted across the width (§3.6): toward `selection` itself for highlight, a sixth of the way toward `success`, `error` or `warning` for the others. As OpenTUI's Code and LineNumbers; glamour, the reference shot's, draws code blocks with neither numbers nor marks |
| HottyDiff | a column for the rail, then rows. For each file: a blank row before all but the first; its name when it has one (or the diff has several files), bold, then `+N` in `success` and `-M` in `error`, or `new`, `deleted` or `binary`. For each hunk: `⋯ N unchanged lines` in `muted` when it leaves lines out before it (and after the last, where the diff knows, from two texts); its header in `info` and its section in `muted`; then its lines. Unified, a line is its old number and its new (blank on the side that does not have it), right-aligned as wide as the widest, in `muted`, each and a space, when `lineNumbers` is on (the default); its sign (`-` in `error`, `+` in `success`) and a space; then its code, its tokens coloured as a HottyCode's, wrapped under the code or, without `wrap`, cut with `…`. Split, where each side has room for 16 columns of code (unified where not): the old side, ` │ ` in `border`, then the new, each a number, a sign and code; a context line on both, and in a run of removed lines followed by added ones, each removed line beside the added line that replaces it; a row is as tall as its taller side, and a side with no line is blank. A removed line's rows are tinted a sixth of the way toward `error`, an added one's toward `success`, its changed words about a quarter (68/255), under its tokens' own colours (§3.6). The selected hunk's rows have `▎` in the rail and its header is reversed, both in `muted`, in `accent` while the diff has the keyboard; a scroll view around it keeps that hunk in sight. With no hunks and no name, `No changes` in `muted`. As OpenTUI's Diff, which has no names, folds, word marks or selection |
| HottyTree | as lipgloss's tree draws one, a row a node, every row 2 columns in: for each level below the roots, its guides in `border`, `│   ` down past an ancestor a sibling follows (else four blanks), then `├── ` before a node a sibling follows and `└── ` before the last; a branch's fold, `▶ ` closed or `▼ ` open, in `muted`, as bubbles' tree has it, or a root leaf's two blank columns, so that the roots line up, none in a tree with no branches (a docs nav's pages, flush with a caption over them but for the selection's two columns); its label, the filter's matches underlined, cut with `…` when it does not fit; and a closed branch's count, ` 3`, in `muted`; a branch with nothing in it yet (`children: []`, §6.21) is drawn as a closed one, ` 0`. A node's siblings are those that show, so that while a filter applies a line ends at its last match. The selected node's first two columns are `│ `: in `accent` with its label, bold, while the tree has the keyboard, and in `muted` otherwise, its label bold, so that the node it is on (a docs nav's page) still shows beside the guides (docs, 2026-10-10; the maintainer, round 6). With a `height`, that many rows show: moved as little as brings the selected node into view when the selection moves, and three rows a notch by the wheel, which leaves the selection where it is (§3.7). A node's icon is a host's: few have a glyph a column wide, and the guides and folds say what a tree is. With no node shown, its empty text in `muted`. |
| HottyChart | as ntcharts draws one: its ticks' labels in `muted`, right-aligned as wide as the widest, each on the row its value is drawn in (`view.TickRow`), `┤` there on the axis and `│` on the other rows, and `└` then `─` across under the plot, in `border`. The axis (§6.18) has its ticks a free row apart where the plot is tall enough; where it is not, 0's label, the top tick's and the bottom one's come first, then each other one with a free row either side, as plothot labels its axis. A line is braille dots (hotty-go's `braille`), two across and four down a cell, in its series' colour, joining its points: a point's dot row runs from the middle of the bottom row, the axis's low end, to the middle of the top row, so that a tick's label is level with its value; its dot column is its slot's (§6.18). A missing value breaks the line, and a lone value is a dot. A cell takes the colour of the last series to put a dot in it. Under the plot, three of its points' labels (five where the plot is 60 columns or wider), evenly spread over its slots: the first starting at its slot, the last ending at its own, the others centred, and one that would touch the one before it left out. Bars are a group a point, the plot's columns shared out between them a column apart, each series' bar an equal part of its group, in its colour, from the axis's 0, which may fall inside a row, to its value: a cell its bar covers from the cell's bottom is that many eighths (`▁` to `█`, at least `▁` in the cell 0 is in); one it covers from the top, below 0 or above a 0 inside the row, is `▀` from a quarter, `█` from three quarters, and `▔` for less, since a cell has no other blocks that hang. Where the groups do not fit at a column a bar, the last points that fit show. Each label is centred under its group, cut with `…`. Then the legend, when it has several series or a labelled one: a key in the series' colour, `━` for a line and `■` for bars, and its label in `muted`, three columns between series. The series' colours are `info`, `warning`, `success` and `error`, then round again; not `accent`, which marks only focus (§3.6). With no values, `No data` in `muted` in the plot's middle row. |
| HottySparkline | as ntcharts' sparkline: a column a value, in `info`, rising from its bottom in eighths of a row (hotty-go's `blocks`), `height` rows tall, from its `min` at the bottom to its `max` at the top: the newest value in its last column and those before it to the left, as many as fit, so that a sparkline with a `window` keeps its width as values arrive. A value at `min` or below is `▁`, so that only a missing value is blank. |
| HottyPaginator | as bubbles' paginator under what it pages: its child past the gutter, holding the page the view shows, the rows of the tallest page, so that the dots stay put as the pages turn (bubbles' rise on a shorter last page); then a blank row and its pages: a dot for each (`•`, the page shown's in `fg`, the others in `border` and faint, as bubbles' list's very subdued dots, so that without colour the page shown is the one dot not faint), each a click target; with `displayStyle` `numbers`, or where the dots do not fit, the page shown of how many, `3/10`, in `fg`. Without a child it is that row alone. While it has the keyboard, the gutter's bar runs down its rows, its child's and its dots', and the page shown, its dot or its numbers, is in `accent`. A click anywhere on it gives it the keyboard. |
| HottySpinner | the frame of its set the clock is at, in `info`, then the label after the widest frame's columns and a space, so that it stays put. One that is not active leaves the frame's columns blank. |
| HottyTimer, HottyStopwatch | on one row, as bubbles' timer and stopwatch examples draw theirs (`Exiting in 4s`, `Elapsed: 1.5s`): its label and a space, then its time as its `format` writes it (§6.25), `4m59s` or `4:59`, in `fg` while it counts and in `muted` while it stands still, stopped or run out, so that a time that does not move does not look stuck. Without a label, the time alone. Its width is its time's, which changes as it ticks (`9s` after `10s`), as bubbles' does: put it at the end of its row. |
| HottyBigText | as OpenTUI's ASCIIFont, block letters in `fg`, of the kit's own pixel fonts (§6.28): its size's glyphs, a column apart, a word a space's glyph and a column each side; at small (3 rows) and medium (4) a pixel is a column and half a row, two pixels a cell (`▀ ▄ █`), at large (5) two columns and a row (`██`). Each line at its `align` across its box; a line too long wraps at its last space that fits, and a word too long for a line of its own between letters, so that nothing is cut; its lines a blank row apart. Lowercase is drawn in capitals, a Latin letter's mark is left off (`É` is `E`), and a character the fonts lack is `?` |
| Placeholder | `…` in `muted` while pending; `! Type` in `warning` when the type is unknown or the component contains itself |
| an error | `✗ message` in `error` under its control, wrapped, its continuation lines indented 2. A one-line field's starts in its input's column, under the value it is about (the maintainer, round 3), unless that leaves it fewer than 16 columns |

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
Update on a host. A toast counts its time down on the same clock, by the
draws that show it, and asks for one every quarter second while it does
(`view.ToastStep`, §6.23). A HottyTimer or a HottyStopwatch counts by the
draws too, whether they show it or not (§6.25), and asks for one every
tenth of a second while it counts, or every `interval` where that is
shorter, down to a thirtieth (`view.TimerStep`): its time shows within a
step of when it changes, and a timer runs out within a step of its end. The program draws or updates again then: the
storybook's `Book.Tick` and `storybook -bare` do, on Bubble Tea's
`tea.Every`. So two renditions of one surface made at once show the same
frame, and a test sets the clock (`Rendition.Clock`, in both).

**Fields** are the text fields, DateTime, CheckBox, HottySwitch, Choice,
Slider and HottyRangeSlider: the controls huh calls fields, drawn as huh
draws them; and a HottyPaginator, which has a field's gutter, and no
label.
- **The gutter.** A field's first two columns are its gutter. While the
  field has the keyboard (for a Choice's options, one of them), the
  gutter is `┃ ` in `accent` down the field's rows; otherwise it is blank.
- **The label** is in the text's colour faded 30% toward the background
  (`fg` alone where the terminal has not said its colours, §3.6), a step
  under the value but not `muted`'s grey, which read as too grey (the
  maintainer, round 4); in `accent` while the field has the keyboard. A one-line text field's, a DateTime's and a select's is on its
  input's row, before it (§3.5). A longText's and a Choice's options'
  label is a title row above them, as a GUI form's label over a textarea
  or a group. A CheckBox's, a HottySwitch's, a Slider's and a
  HottyRangeSlider's labels are on their one row. A HottyRangeSlider has
  the keyboard while either knob does.

**Focus** (the element with the keyboard). Each focused element is in
`accent`, with a glyph or an attribute so that the focus still shows
without colour:
- A field shows its gutter's bar, and its label is in `accent`.
- A Button, a Tabs' title, a Media link and a Modal trigger that is not
  a control are reversed whole; a Button's fill gives way to the reverse,
  its padding reversed too.
- A CheckBox's box is reversed in `accent`; an option's row starts with
  `> ` in `accent`, its box or circle in `accent`.
- A HottySwitch's label is in `accent`.
- A HottyPaginator's page shown, its dot or its numbers, is in `accent`,
  its gutter's bar down its rows.
- A Slider's track is in `accent` up to the knob, and so is the knob
  (from the knob, with `fill` `"end"`).
- A HottyRangeSlider's range is in `accent`, and so is the knob with the
  keyboard, whose number in the value is in `accent` and bold, so that
  without colour the value still says which knob the keys move.
- What Shift, Control+a or a drag selected (§5) is on the `selection` colour,
  or reversed where the theme can't tint, as a terminal shows its own.

**A drag** (§6.21) is painted over the frame, so that no row moves:
- What it lifted is faint: a list's item, a tree's node and the nodes
  under it that show, a List's item, a drag source.
- Where the item would land is a line in `accent`, an underline (SGR 4)
  along the row above the place, across the item's columns, a tree's from
  the item's level: the item above, the blank row between two items, a
  HottyList's blank row under its status line, a HottyTable's rule, a
  List's heading. After an item that another follows is before that one,
  so the line takes the blank row between them where there is one. The
  underline's colour is the accent's (SGR 58: the theme's, or the
  256-colour index of the ANSI one), so the row's text keeps its own; a
  terminal without SGR 58 draws it in the text's colour. At the frame's
  top, with no row above, it is an overline (SGR 53) along the row under
  the place, that row's text in `accent`, since an overline takes the
  text's colour.
- Into a tree's node, the node's row from its level is reversed in
  `accent`; so is a drop target, whole.

### 3.5 Text fields

- **The input.** A one-line field's (a text field's, a DateTime's, a
  select's) value is in its input, after its label and a blank column,
  the inset, or after the gutter and the inset when it has no label. The
  input is underlined (SGR 4), as a GUI form's input is outlined, and
  nothing else on its row is: not the label, not the inset, which parts
  the line from the label (the maintainer, round 2, vault feedback
  2026-10-10). The underline starts at the value's first column and ends
  where the input does. The input is a fixed width, as a GUI form's is,
  and never grows as the value is typed: a text field's is 32 columns,
  room for an email address; a DateTime's, its value's form or its
  value, whichever is wider, and a column for the caret; a select's, its
  widest option, a space and `▾` (§3.4). It takes less where the field's
  box is narrower, and scrolls. The underline's colour is `border` at
  half its contrast with the background, and `accent` at half while the
  field has the keyboard, so that it is there without drawing the eye
  from the text (the maintainer, rounds 3 and 4). It is SGR 58, as a
  drag's line in §3.4, so the value keeps its own colour; a terminal
  without SGR 58 draws the line in the text's. Where the colour or the
  background's is unknown (the terminal has not said them, §3.6), the
  line is in `border` or `accent` itself. There is no prompt: the
  underline says where to type. A longText is not underlined: its rows carry `┃` in the
  gutter, in `border`, as bubbles' textarea does. The placeholder shows
  while the value is empty, in `muted` and faint. A DateTime's
  placeholder is the form its value takes: `YYYY-MM-DD`, `HH:MM` or
  `YYYY-MM-DDTHH:MM`.
- **The label** of a one-line field is on its input's row, a space after
  it; in a run of such fields next to each other in a Column, the labels
  are padded to the widest, so that their inputs start in one column. A
  label too wide for the field is cut to leave a space and 3 columns. A
  click on the label puts the cursor at the value's start. A longText's
  label is a title row above it. At 40 columns, Name with the keyboard,
  Ada typed, and an empty Email after it:

  ```
  ┃ Name   Ada
    Email
  ```

  where `Ada` and the columns after it, to the field's edge here, are
  Name's input, underlined in the toned `accent`, and Email's from the
  same column, underlined in the toned `border`.
- **obscured** shows `•` for each cluster.
- **longText** shows its hard lines, at least 3 rows and at most 8, with no
  soft wrap.
- **Without the keyboard** a field shows its value from the start, each
  line cut with `…`.
- **With the keyboard** a field shows the cursor:
  - The cursor is an index into the value's clusters; it starts at the
    end.
  - `Frame.Cursor` reports it, and the terminal's cursor goes there. Its
    shape says which keys the field has. Under the default keymap, a
    GUI field's, it is a line (`Frame.BlockCursor` false; a program asks
    the terminal for a bar, DECSCUSR), since a GUI's muscle memory comes
    with a line. Under a keymap set over it, a terminal's
    (`TerminalKeys`), it is a block, and the cell under it is reversed
    too, except beside a selection, where a block would read as one more
    selected cell. A HottyList's filter has the same cursor.
  - It blinks as a GUI field's caret does: it shows when the field takes
    the keyboard, after a key and after a press, then hides and shows by
    turns, 530 ms each, and the frame says when to draw it again
    (`Animating`). The terminal's cursor is asked to hold steady, so the
    two don't blink at odds.
  - It is in `accent` where its colour is known, the theme's or what the
    terminal said (§3.6, `Known`): a program sets the terminal's cursor
    colour (OSC 12), as Bubble Tea's `Cursor.Color` does, and the
    terminal's own comes back when the program exits. Elsewhere, and
    without colours (NO_COLOR), it is the terminal's.
  - The pointer is an I-beam over a field's input (`Rendition.Pointer`
    is `"text"`; a program sets it with OSC 22, and hears the pointer move
    with no button down), as a browser shows over a field. On a host the
    field's `cursor: text` asks for the same (SPEC.md §9).
  - The field scrolls as little as keeps the cursor in it: across, by
    columns (it shows from the start whenever the cursor's line fits),
    and for a longText, down by lines.
- **Suggestions** (§6.22), while a one-line field has the keyboard and
  its value leaves some:
  - The rest of its suggestion (the highlighted one, else the first) shows
    after the value, in `muted` and faint, as bubbles' text input draws
    it, cut at the field's edge. The cursor stays where it is: at the
    value's end it is on the ghost's first cell.
  - Unless its `list` is false (§6.22), a list drops down over what is
    under the field, as a GUI's menu does: it moves no row, hides what it
    covers and takes its clicks, as an open Modal's panel and a toast do
    (§3.7), and the frame grows to hold it. It is a box with rounded
    corners in `border`, from the row under the value, a column of
    padding inside it, so that its text stands under the value's; as wide
    as its widest row needs, at most the frame, moved left to fit. In it,
    a row a suggestion: what the value typed of it plain, its rest in
    `muted`. The highlighted one is reversed in `accent` across the box,
    as a menu's item. At most 5 show: past that, the 5 from the first,
    moved as little as keeps the highlight in view, then a row
    `1–5 of 12` in `muted`, inside the box. A row too wide is cut with
    `…`. The gutter's bar stops at the input's row. At 30 columns, `git c`
    typed, a Note field after it:

    ```
    ┃ Command  git commit
      Note   ╭─────────────────╮
             │ git commit      │
             │ git checkout    │
             │ git cherry-pick │
             │ git clone       │
             │ git config      │
             ╰─────────────────╯
    ```

    where the first `ommit` is the ghost and the others' rests are `muted`.
    With nothing typed, nothing the value leaves, or the list shut, the
    field is as without them: its placeholder, or its value alone.
  - With `list` false, the ghost alone shows, as bubbles' text input has
    it; ArrowDown and ArrowUp still move through the suggestions, and the
    ghost follows the highlighted one.

### 3.6 Colour and attributes

Cells name NEIO-4's roles (`fg`, `bg`, `muted`, `accent`, `selection`,
`surface`, `border`, `success`, `warning`, `error`, `info`), never colours,
but for one: a HottyProgress's fill may blend into the terminal's own
magenta, where the terminal said it (§3.4). `accent` goes only on focus, and on a HottySwitch that is on, as a
switch's colour says so. Every state that has a colour also has a glyph
or an attribute: `✗`, `!` and `…`, reverse for focus, faint for disabled
and for placeholders. A toast's kind (§6.23) is its border's and its
mark's role, `info`, `success`, `warning` or `error`, and its mark, `ⓘ`,
`✓`, `!` or `✗`, says it without colour.

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
k / 255 in integers, truncated. It takes effect where both roles' colours
are known (below); elsewhere, and at the ANSI-16 floor, the cell is the
first role's.

A cell's background may be tinted (a HottyCode's marked line): k/255 of
the way from the background to a role's colour, as above. It takes effect
where both are known. A HottyDiff's changed lines and words are tinted so
too, under their tokens' own colours, and a surface's fill is a tint all
the way to `surface`: a HottyScrollView's box (§3.4), which is not filled
where it cannot be. Where nothing can be tinted, a changed line's sign
says it alone, and its changed words are underlined; under NO_COLOR they
are reversed, as git's diff-highlight shows them. A Button's fill (§3.4)
is a tint toward `fg`; where nothing can be tinted it is the terminal's
bright black as the background (SGR 100), which is what huh's grey is
there, under the terminal's text; under NO_COLOR, an underline.

**Known colours.** A role's colour is known where the theme colours it,
or, where the theme leaves it to the terminal, once the terminal has said
it. A program asks once, at the start (`cells.TerminalQuery`): OSC 10 and
11 for the text and the background, and OSC 4 for each colour of the
floor above by its number (1, 2, 3, 6, 8 and 12) and for its magentas (5
and 13), which a HottyProgress's fill alone takes (§3.4). It takes the answers
into the theme it paints with (`theme.Theme.Term`; `theme.Terminal.Hear`,
or `storybook.TerminalColours` under Bubble Tea, which decodes OSC 10 and
11 itself), and draws again. Nothing waits on them: until they come, and
on a terminal that never answers, the cells are as at the floor. Known so:
- `fg`, `bg` and each role of the floor are the terminal's own;
- `selection` is a fifth of the way (51/255) from `bg` to `accent`;
- `surface` is 24/255 of the way from `bg` to `fg`, a step off the
  background.

These are only what a blend or a tint is worked out from. A role the
theme leaves to the terminal is still painted in the terminal's palette,
by its number, so that the terminal's own colours change it, and the
terminal's background is never painted under every cell: a cell has a
background only where a tint changes it, so a terminal drawn through (a
transparent one) stays so elsewhere.

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
- **A text field's suggestions** (§6.22) take keys before its keymap
  (`view.Controller.SuggestKey`), as bubbles' text input binds them:
  - Tab takes the suggestion; with none to take it moves focus as ever.
    ArrowRight at the value's end, with nothing selected, takes it too.
  - ArrowDown and Control+n highlight the next suggestion, ArrowUp and
    Control+p the one before, round from either end; from none, ArrowDown
    highlights the first and ArrowUp the last. ArrowDown and Control+n
    open a list Escape shut. The highlighted one is Tab's.
  - Enter picks the highlighted one; with none highlighted, it submits as
    ever. Escape shuts the list until the value changes; with no list it
    goes on as ever.
  - A taken suggestion is the value as written (`ger` takes `Germany`),
    the cursor at its end, and the list stays shut until the next edit.
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
- **A HottyRangeSlider's knob** steps as a Slider does, by its range's
  step, and stops where it meets the other knob: the start's goes no
  higher than the end, and the end's no lower than the start, Home and
  End included. Tab goes from the start's knob to the end's.
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
- **A HottyPaginator** turns its pages as bubbles' paginator does
  (`view.Controller.PageKey`, in both renditions): ArrowLeft and
  ArrowRight (h and l, PageUp and PageDown) a page, not past the first or
  the last, and Home and End to them, which bubbles' leaves out. It takes
  each even at an end. Each turn writes `page`, then runs `onChange`
  (§6.26). Its child's controls take keys as they would anywhere; the box
  is a stop of its own in the Tab order, before them.
- **Alt with the arrows** moves an item (§6.21): Alt+ArrowUp and
  Alt+ArrowDown the selected item of a `reorderable` HottyList,
  HottyTable or HottyTree, or a List's item (`reorder`) the keyboard is
  in, a place; in a HottyTree also Alt+ArrowLeft out of its branch and
  Alt+ArrowRight into the node above (`view.Controller.MoveKey`, in both
  renditions). A text field takes them first where its keymap binds them.
- **Escape** during a drag drops it where it started.
- **Space and Enter** activate a Button, a Tabs' title, a chip, a Media
  link, or a Modal trigger that is not a control, and flip a HottySwitch:
  it is a button on a host, which Enter clicks, so in a HottyForm too
  Enter flips it and does not submit.
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
  Then it dismisses the newest toast (§6.23), as on a host; a focused text
  field keeps the keyboard. On a toast's action, a button, Escape
  dismisses that toast before anything else, and Space and Enter send its
  event and dismiss it; the keyboard goes with it.

**A click** lands on the topmost thing drawn at its cell:
- **Something that takes focus** gets the keyboard, and then:
  - a field puts its cursor before the cluster clicked (a click on its
    label line only focuses it), and a drag from there selects to the
    cluster under the pointer, to the value's start or end past them, as
    in a GUI's field. With Shift (`ShiftClick`), a press in the field that
    has the keyboard extends the selection from its anchor, or from the
    cursor, instead; a terminal sends Shift with a click only when the
    program asks (XTSHIFTESCAPE), which the storybook does;
  - a select opens or closes its list, and a click on a row of the open
    list picks that option and closes the list;
  - a click on a row of a text field's suggestions takes that one, the
    cursor after it; a click on the list's border lands on the list, not
    on what it covers, and leaves the field as it is;
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
  - a click on a HottyRangeSlider's track moves the nearer knob to the
    value at that column (the start's, midway between them) and gives it
    the keyboard; a click on a knob gives it the keyboard where it is.
    Either way the knob follows the pointer while the button stays down,
    stopped where it meets the other. Where the knobs meet, the click
    moves neither, and the first column the pointer moves to picks the
    knob it goes towards (§6.20);
  - a click on a HottyPaginator's dot shows its page (§6.26); a click
    elsewhere on it, its page's text included, only focuses it, and what
    in its page takes a click takes its own;
  - anything else is activated.
- **A press on what can be dragged** (§6.21: a `reorderable` list's
  item, a List's item with `reorder`, a drag source) may start a drag. It
  selects a list's item at once; what else the click does (the second
  click's action, a branch opening, a Button in a card running) waits for
  the release, and a drag does not do it. Once the pointer leaves the item
  with the button down, the item is lifted and the line follows the
  pointer (§3.4); the release drops it where the line shows. A release
  where no line shows, over what does not take the item or outside the
  surface, leaves it where it was, and so does Escape. The innermost one
  under the pointer is dragged: a drag source in a List's item drags
  itself. A press in a text field, or on a Slider or a HottyRangeSlider,
  is theirs, in a card too.
- **Where in a row.** A row is a cell high. Where the terminal reports
  the pointer in pixels (SGR-Pixels, mode 1016), an item's top half is
  before it and its bottom half after it, and a tree's node that takes
  children has thirds, its middle one into it. The storybook asks the
  terminal whether it has the mode (DECRQM) and how big its cells are
  (XTWINOPS 16), sets the mode once both answer, turns reports back into
  cells, and resets the mode on its way out
  (`cells.Rendition.DragAt`'s `sub`). Elsewhere an item of one row has no
  halves: it takes the place of the item it is over (after it, coming
  from above; before it, from below), or goes into a node that takes
  children; coming from another list, before it. An item of several rows
  (a HottyList's with descriptions, a card) has halves by its rows.
- **A disabled Button** does not take the keyboard, but the click still
  activates it, as in rendition/html; the controller decides what that
  does. Nor does a disabled HottySwitch, which the click leaves as it is.
- **A click on nothing that takes focus** gives the keyboard back.
- **A click outside an open Modal's panel** closes the Modal.
- **A click on a toast** (§6.23) is on it, not on what it covers: on its
  action, the action sends its event and the toast goes, and the keyboard
  with it; anywhere else on it, the toast goes and the keyboard comes
  back, as a click on a box does on a host. A press on a toast starts no
  drag.

**The pointer** with no button down, where the terminal reports every move
(mode 1003, which the storybook asks for), is the rendition's
(`Rendition.Hover`, which reports whether the frame changes): over a
toast, the toast's time waits; over an element with an
`accessibility.description`, or in one, the innermost such element's
description is the tooltip (§6.23), until the pointer leaves it. A key or
a click hides it, as a GUI's tooltip goes, until the pointer leaves the
element. Under an open Modal's panel only its content is described, and
under a toast nothing.

**The wheel** over a HottyTree with a `height` moves its rows, three a
notch, while they can move that way. Else it scrolls the innermost
HottyScrollView, or HottyCode whose lines are cut, under the pointer that
can still move that way: three rows a notch, or sideways (the wheel's
left and right, or Shift with it) six columns, as bubbles' viewport takes
it. A HottyCode scrolls only sideways, and needs no focus for it. A
notch past an end goes to the box around it, if any. The storybook
passes the wheel to the story under the pointer; a program without mouse
reports has no wheel.

### 3.8 What cells keep

The cursor, the scroll offsets, which select's list is open, a press
that may start a drag, and what the pointer is over (the tooltip it
shows, the toast it holds, the element a key hid the tooltip of) belong
to the rendition, by element id; a drag under way is the controller's
(`view.Drag`), and so are a text field's highlighted suggestion and
whether its list is shut (`view.Suggest`), and the toasts and their
time (`view.ToastState`).
Everything else is the surface's state (§1).

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
  character, and an empty one its placeholder in parentheses. Its
  suggestions (§6.22) say nothing: they show while it has the keyboard,
  which a pipe never gives it.
- A CheckBox is `[x] Label` or `[ ] Label`; a ChoicePicker `Label: ` and
  the labels picked; a Slider `Label: value (min–max)`; a
  HottyRangeSlider `Label: start–end`, followed by `(disabled)` while it
  is.
- A HottySwitch is `Label: on` or `Label: off` (`on` or `off` alone
  without a label), followed by `(disabled)` while it is.
- A Button is `[ label ]`, followed by `(disabled)` while its checks fail.
- A HottyProgress is `Label: 42%`, or `Label: …` without a value; a
  HottySpinner is `Label: …` while it spins, else its label alone.
- A HottyTimer or a HottyStopwatch is `Label: 4m59s`, the time it showed
  when the view was made (the last draw counted it), in its format; the
  time alone without a label. A pipe has no time to tick in.
- A HottyBigText is its text as it is, a line each of its lines, its case
  and its accents kept: what it says, not its letters' blocks.
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
- A HottyChart is a table of its values: a row a point, its label first
  (else its number, from 1), and a column a series under its label
  (`value` for one without), the numbers as the data model writes them,
  aligned to the end, a missing one blank, two spaces apart; `No data`
  when it has no values. A pipe reads the numbers, not a picture.
- A HottySparkline is its values on a line, a space apart, `–` where one
  is missing; nothing when it has none.
- A HottyScrollView is all of its content, whatever its `height`: its
  lines, a line each, or its child.
- A HottyPaginator is its child with every page's items, whatever the
  page, as a HottyScrollView is all of its content, then the page shown,
  `Page 2 of 6`; without a child, that line alone.
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
- The toasts (§6.23) follow that, after another `───`, a line each, the
  newest first: the kind's mark (`ⓘ`, `✓`, `!`, `✗`), a space, the
  message, and the action as a Button, two spaces after it
  (`✗ Could not reach the server  [ Retry ]`). A pipe has no time: a toast
  is listed while the view has it.
- A component with `accessibility.hidden` says nothing, and an
  `accessibility.description` is not written: a pipe has no pointer and no
  focus to show it for.

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
  sets the surface's on its top elements as `data-keys`: none by default,
  so that fields have SPEC.md's default keymap, a GUI field's, and a line
  cursor in cells (§3.5). A program may set another over it (`SetKeys`),
  such as hotty-go's `TerminalKeys`, the keys of Bubble Tea's text input
  (Control+e, Alt+b and Alt+f, Control+w, Control+k, Control+u…), which
  cells shows with a block cursor. Either way a field keeps a GUI's
  muscle memory:
  Shift with any move selects from where the caret was, Control+a selects
  all (in `TerminalKeys` too, where Home goes to the line's start), and
  typing or deleting replaces the selection. macOS's Command keys come as
  Meta and do what they do in a Mac field: Meta+ArrowLeft and
  Meta+ArrowRight go to the line's ends, Meta+ArrowUp and Meta+ArrowDown to
  the input's, Meta+Backspace deletes to the line's start and Meta+a
  selects all. The full key help lists them after the others. A component's `keys` (§6.5) overrides it key by key for
  the fields inside. The cells rendition resolves the same keymap
  (`hotty.Resolve`), so a key does the same in both.
- **HottyShortcuts** (§6.1) take the keys that reach the program, for the
  surface that is active: the one that has the keyboard, or when none has
  it, the one the program says is current.

Shortcuts apply to a whole surface, in both renditions, though a host now
names the element the user focuses (SPEC.md §10.1, §2's keyboard). Whether
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
renditions: a story played by keys, focus, blur, clicks, and taps and
drags on a slider's track, Alt with the arrows moving items (§6.21), with who has the keyboard, the actions the
agent got, the data model and the toasts shown (§6.23) expected after each step. The host side
runs on hottytest's host, whose `Key` takes keys as SPEC.md §10.2 has a
host take them, and which reports a drag's steps (SPEC.md §9.1).

## 6. The hotty catalog

`https://hotty.neuroplast.io/a2ui/v1/catalog`, in
[catalog/hotty/catalog.json](../catalog/hotty/catalog.json). The id is a
name, as A2UI's catalog ids are ("not a resolvable URI"): nothing is
served there. An agent gets the catalog document from its program
(`hotty.Doc`), in its prompt or as an inline catalog. `v1` is this
catalog's own major version; the A2UI version it targets is its
`protocolVersion`.

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
`data-keys`; in cells the rendition resolves it the same way. A key bound
to `program` there reaches the program from any element inside with the
keyboard, a Button too, before the host scrolls with it (SPEC.md §10.2,
*Keys for the program*): a Confirm's Row gives its Buttons' arrows to
its HottyShortcuts so (§6.24). In cells a Button leaves those keys to the
HottyShortcuts anyway (§3.7). Other renderers ignore it.

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

With `reorderable` and `rows` bound, the user moves its rows (§6.21).

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

With `reorderable` and `items` bound, the user moves its items (§6.21).

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
   a HottyForm, a CheckBox's `space toggle`, a HottySwitch's
   `space/enter toggle`, a Button's `enter press`, a
   Slider's or a HottyRangeSlider's knob's `←/→ adjust`, a HottyPaginator's `←/→ page`, as
   bubbles' paginator example's, a select's `enter open`, a HottyScrollView's
   `↑/k up`, `↓/j down`, `f/pgdn page down`, `b/pgup page up`, a
   toast's action's `enter` and its label, lower-cased, and `esc
   dismiss` (§6.23); and `esc close` while a Modal is open.
2. The surface's HottyShortcuts that have a `label` (§6.1), each its key
   as bubbles writes keys (`ctrl+s`, `alt+←`, `pgdn`) and the label.
   Those with the same label are one hint, their keys joined by `/`, as a
   bubbles binding of several keys shows them (`←/→ move`, §6.24). While
   a text field has the keyboard, a HottyShortcut whose key it takes, a
   character it types or a key its keymap binds to an edit, is left out,
   since the key never reaches it (§6.1).
3. In the short view, `? more` while `toggle` is on.

The full view, which `?` switches to and back from (`toggle`, on by
default; turn it off where `?` is the surface's own key), shows those as
groups in columns, as bubbles' full help: the component's keys, more of
them, in a column or two as bubbles splits them (a list's moves, its
pages, `g/home` and `G/end`, then its filter and Enter; a table's rows,
then its pages; a diff's hunks and ends; a paginator's `→/l/pgdn next page`, `←/h/pgup prev
page`, `home first page` and `end last page`; a scroll view's rows and ends, then its pages and half
pages, then `←/h move left` and `→/l move right` for lines that do not
wrap, as bubbles' viewport's; a text field's moves, then its edits, from
its keymap, §5, two keys an action at most); the HottyShortcuts; then `tab next`,
`shift+tab back` and `? close help`.
Which view shows is the renderer's state, shared by the surface's
renditions. It takes no focus and sends nothing.

On a surface where a component has an `accessibility.description`, the
HottyKeyHints is also its status line for tooltips: a row over its keys
shows the description of the element under the pointer, else of the
element with the keyboard (§6.23).

On a host, the component's keys are those of a HottyTable, a HottyList,
a HottyDiff, a HottyTree, a HottyPaginator, a Slider, a HottyRangeSlider's knob or a select, whose keys the program works; the host moves focus
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
wraps long lines; off, they are cut, and scroll sideways under the wheel
in cells (§3.7) and on a host. It has no height: a HottyScrollView
scrolls a long listing. It takes no focus; selecting and copying it are
KIT-12's.

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
moved as the selection moves and by the wheel (§3.4); without one, it is
as tall as the nodes that show. `emptyText` shows when none does
(`Nothing here.` by default).

`hottyExpandAll({id})` and `hottyCollapseAll({id})`, renderer functions
the agent or a Button may call (`allowedCallers: rendererOrAgent`), open
every branch or close them all, the id resolved as `hottyFocus`'s is
(§6.3). They write `expanded` where it is bound, every branch's id or an
empty list, and the renderer's folding where it is not. Closing them all
selects the root the selection is in, as closing the branch it is in
would.

With `reorderable` and `items` bound, the user moves its nodes, among
their siblings, into a branch and out of one (§6.21). A node whose
`children` is a list, an empty one too, is a branch: it takes children.

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

### 6.16 icons and fill

`metadata.extensions.io_neuroplast_hotty.icons` gives the titles of a
Tabs, or the options of a ChoicePicker, an icon each, which the basic
catalog has no place for: on a Tabs, a list of names, a tab's by its
index; on a ChoicePicker, an object from an option's `value` to its
name. A name is a HottyIcon's (§6.15). A tab or an option it names
nothing for has no icon, and an entry past the last tab is ignored. On
a host the icon goes before the title or the label (§2), hidden from a
screen reader, which the label tells; cells and text leave it out, as
few icons have a glyph a column wide, as a HottyTree's (§6.14). Other
renderers ignore it, as they do autofocus (§6.4) and keys (§6.5).

`metadata.extensions.io_neuroplast_hotty.fill` says which side of a
Slider's knob is filled: `"start"`, the default, from `min` to the knob,
as every Slider is; or `"end"`, from the knob to `max`, for a value whose
chosen part lies above it, such as a minimum rating, where what the
filter keeps is what is filled. In cells the track's two sides swap
(`⎯⎯⎯⎯■━━━━`, §3.4), its glyphs and the knob unchanged, and on a host the
fill goes on the other side of the knob (§2). The value, its keys and its
clicks are a Slider's either way. Any other value is the default, and
other renderers ignore it, filling from the start.

### 6.17 HottySwitch

An on/off switch for one setting, as a phone's settings have them: a
boolean as a CheckBox's is, with its `label`, its `value` (a
DynamicBoolean, best bound, so that the user's flips are written to the
data model) and its `checks`, whose error shows as a field's does (§6.2).
Use it for a setting that takes effect at once, and a CheckBox for a
choice that a form sends later, as Apple's and Material's guidance has
it.

Space, Enter or a click flips it (§3.7). On a host it is a button (§2),
which the host clicks on Space and Enter, so Enter flips it in a
HottyForm too, as it presses a Button there, rather than submitting the
form as Enter on a CheckBox does; Enter in a text field still submits.
`disabled` (a DynamicBoolean) keeps it from flipping and from the
keyboard, Tab passing over it; bound, it disables the switch while
another setting rules it out. Without a `label`, give
`accessibility.label`, which names the switch on a host.

A component of the hotty catalog takes only that catalog's functions
(vault a2ui-limits L1), so its `disabled` and its checks' conditions are
literals or paths, not basic's `not` or `or`.

### 6.18 HottyChart

Numbers on an axis: `kind` `line` (the default), a trend over time, or
`bar`, amounts by category. `values` are one series' numbers, oldest
first; `series` are several, each a `label` for the legend and its
`values`. A null, or anything that is not a finite number, is a missing
value. `labels` are a label a point, in the same order: times for a line,
categories for bars. Each of them is literal or bound, and best bound to
a list in the data model, where the agent adds a point by writing the
next index (`/cpu/42`): the data model has no append (vault a2ui-limits
L4). A long series belongs out of a surface with `sendDataModel`, which
would send it back with every action (L5). It takes no focus and no
keys.

`height` is the plot's rows (8), without the axis, the labels and the
legend. `window` is the points its x axis holds: the last `window` of
them, the newest at the right edge, so that a series that grows slides
along, and the slots before its first point blank while it fills.
Without one, its points spread across the plot. A point's slot is its
place among `max(window, points)`, so that cells and a host put it at
the same place across.

The axis is the view's (`view.ChartAxis`), shared by every rendition:
fitted to the values by hotty-go's `series.Scale`, from 0 when they are
all of one sign, with some headroom, its ends on round ticks, about
`max(3, min(6, height / 2))` steps of them. When those ticks would not
each fall in a row of their own with a free row between, so that cells
labels them all, it is the axis Scale makes with the most steps that
does, if that spans at most half as much again; in a plot too short for
any, the tightest of them, with the fewest ticks. `min` and `max` fix an
end instead; a `max` not above the `min` is dropped. A value outside the
axis is drawn at its edge. A tick's label has the decimals its step
needs (`0.25`, `0.50`), and is in thousands (`k`), millions (`M`) or
billions (`G`) once the step is one (`2.5k`, `125k`), so that the labels
stay narrow.

In cells a line is braille dots and bars eighths of a row (§3.4); on a
host, an `svg` line and boxes (§2); in text, a table of the values (§4).
Horizontal bars, a scatter of points, an area under a line and a second
axis are not drawn.

### 6.19 HottySparkline

A trend in a few cells, with no axis and no labels: `values`, as a
HottyChart's (§6.18), a column each, rising from the bottom. Put it in a
Row between a Text that names it and one with the latest value, bound to
the last index. Its bottom is `min`, else 0, or its lowest value when
that is below 0; its top is `max`, else its highest value. `window` is
the columns it keeps: its last values, the newest at the right edge, so
that it keeps its width as values arrive and shows its slots blank while
it fills; without one, a column a value. `height` is its rows (1). It
takes no focus and no keys.

### 6.20 HottyRangeSlider

A range of numbers picked on one track, such as a price filter: a
Slider with two knobs, the start's and the end's, and the range between
them filled. Give it a `label`, `max` (and `min`, 0 by default), and bind
`start` and `end` (DynamicNumbers) each to a path of its own, where the
user's moves are written; a literal end is the renderer's state once
moved, as an unbound Slider's value is. `steps`, a whole number, cuts the
range into that many steps, which both knobs snap to, as a Slider's
`steps` does; without it a knob moves by a twentieth of the range and
lands on a hundredth. An end with no value is the range's own end, and
the start is never past the end: each is clamped to the range, and the
end to the start.

Each knob is a Tab stop, the start's first. The arrows, Home and End
move the knob with the keyboard as they move a Slider, and it stops
where it meets the other (§3.7), so that the two can meet but never
cross. A click or a tap on the track moves the nearer knob there and
gives it the keyboard (the start's, midway between them); a click on a
knob only gives it the keyboard. A drag moves the knob its press picked
with the pointer, stopped where it meets the other. Where the knobs
meet, a press moves neither and the drag's first move picks the knob it
goes towards, so that a range shut to one value can open either way.
Meanwhile the keyboard is on the knob pressed, in cells, where each has
a column; on a host, where one hides the other, on the end's, which can
move away (the start's at `max`). A move counts as touching it, for its
`checks`.

Cells draws `label ⎯⎯■━━━━■⎯⎯ 20–70` (§3.4); a host draws a track with
two knobs, each a `role=slider` named `Label, start` and `Label, end` for
a screen reader (§2); text is `Label: 20–70` (§4). Without a `label`,
give `accessibility.label`, which names the knobs. `disabled` (a
DynamicBoolean) keeps both knobs from moving and from the keyboard. Its
`disabled` and its checks' conditions are literals or paths, as a
HottySwitch's are (§6.17).

### 6.21 Drag and drop

The user moves a list's items, with the pointer or Alt and the arrows,
and drops things onto components (vault drag-and-drop, the maintainer's
model C with a line, P2). A2UI's actions carry nothing from the renderer
(vault a2ui-limits L2), so the renderer writes, then acts: the items
where they are bound, the move to a path, then the action, whose context
reads it.

**Lists whose items move.** A HottyList, a HottyTable or a HottyTree with
`reorderable` true, its `items` (a HottyTable's `rows`) bound:
- A drag moves an item to where the line shows (§3.4, §3.7). Alt+ArrowUp
  and Alt+ArrowDown move the selected item a place, and at an end it
  stays. In a HottyTree they move a node among its siblings; Alt+ArrowLeft
  moves it out of its branch, after the branch, and Alt+ArrowRight into
  the node above it among its siblings, as its last child, where that
  takes children. A node never goes into itself or a node under it.
- A HottyTree's node takes children when its `children` is a list, an
  empty one too: a folder with nothing in it yet.
- The items are written where they are bound, the moved one at its new
  place, still selected. Then `moved`, bound, is written:
  `{"item": …, "from": {"index": 2, "path": "/todos"}, "to": {"index": 0,
  "path": "/todos"}}`, the item as `selected` names it (its `value`, a
  row's `rowKey` field), an index among its siblings (`to`'s once it
  left), and the path its list is bound to, which tells two lists apart;
  a HottyTree's places also have `parent`, the value of the node it is
  in, `null` at the top. Then `onMove` runs; give its context
  `{"@path": "<moved's path>"}`.
- Literal items move in the renderer's state, as a literal value does
  when the user edits it; `moved` and `onMove` still go.
- `dragType` names what the items are. Two of a kind on one surface with
  the same `dragType` take each other's items: the item goes to the one
  it is dropped in, whose `onMove` runs (else the one it left's). A drop
  target that accepts the type takes one, as data (below). With a
  `dragType` and not `reorderable`, its items drag onto drop targets
  only, and stay where they are.
- Nothing moves while a filter applies.

**A List's items.** A basic List whose children are a template moves its
items with the extension's `reorder`:
`"metadata": {"extensions": {"io_neuroplast_hotty": {"reorder": {"type":
"card", "moved": {"@path": "/moved"}, "onMove": …}}}}`. The template's
list is written, then `moved` (`item` the item's data), then `onMove`,
as above; Lists with the same `type` take each other's items, a kanban's
columns. Alt+ArrowUp and Alt+ArrowDown move the item the keyboard is in,
a card's Button, and the keyboard goes with it.

**Drag sources and drop targets.** Any component:
- `io_neuroplast_hotty.drag`, `{"type": "ticket", "value": {"@path":
  "id"}}`, makes it a drag source carrying `value`, resolved in its scope.
- `io_neuroplast_hotty.drop`, `{"accepts": "ticket", "value": {"@path":
  "/dropped"}, "action": …}` (`accepts` a type or a list of them), makes
  it a drop target. What is dropped on it is written to `value`'s path (a
  drag source's value, or a list's item's data: a HottyTable's row, a
  HottyList's item, a HottyTree's node, a List's item), then `action`
  runs in its scope. The item stays where it was: what a drop means is
  the agent's.
- The innermost wins, under the press and under the drop.

**Where.** Cells drags (§3.7) and takes the keys. A host takes the keys,
which the elements give the program (§2); its drags wait for HOTTY to say
where in a target the pointer is and to name a target that is not
draggable (vault KIT-23h, G1 and G2). Until then a host ignores the
extensions and draws the lists as without them, and a press there selects
as before. Text draws nothing of it.

**Grips.** None in cells: the whole row is the handle, as it is for a
mouse in a terminal. On a host KIT-23h puts a grip, `⠿`, on every row
that moves, a touch's handle that leaves the rest of the row to scroll
(`touch-action: pan-y`); always, since a document can't tell a finger
from a mouse.

### 6.22 suggestions

Values a one-line TextField offers as the user types, as a shell
completes a command or a form a place: the rest of one faint after the
cursor, for Tab to take, and a list of them dropped down over what is
under the field, as bubbles' text input and fish have them (vault
KIT-11c). It is an extension on the
basic TextField, not a component of its own, so the field keeps its
`label`, `value`, `placeholder`, `checks`, its keymap (§6.5) and its
place in a HottyForm, and other renderers ignore it and show the field:

```json
"metadata": {"extensions": {"io_neuroplast_hotty": {"suggestions": {
  "options": {"@path": "/cities"},
  "onInput": {"event": {"name": "suggest", "context": {"query": {"@path": "/city"}}}}
}}}}
```

- `options` (a DynamicStringList, required) are the suggestions, in the
  order they show. Bind it to a list the agent rewrites as it hears what
  was typed, or give a literal list for a closed set. An empty string and
  a second of the same are left out.
- `onInput` (an Action, optional) runs after every change the user makes
  to the value: an edit, a taken suggestion, a host's `input`. An action
  carries no payload (vault a2ui-limits L2), so its context reads the
  field's bound `value`. On a surface with `sendDataModel`, each one
  carries the whole data model (L5), so keep the options and the model
  short there.
- `list` (a DynamicBoolean, optional, default true) shows the list of
  suggestions over what is under the field (§3.5). False shows the ghost
  alone, as bubbles' text input and fish have it: for a field whose
  suggestions one types through rather than picks from, or where a list
  would cover what the user needs to see. The keys are the same; the
  arrows move through the suggestions the ghost shows. A host's
  `datalist` ignores it until the kit draws a list of its own there
  (vault KIT-11h).
- The field shows the options that start with its value, case aside, but
  for the value itself, in their order: the renderer narrows them at each
  key, so that the list is right while the agent's rewrite is on its
  way. With nothing typed none show. An agent that matches otherwise (a
  word inside, a fuzzy match) still sees its options narrowed to those
  that start with the value.
- Only a one-line field that shows what it holds takes it: not a
  `longText`, where the arrows move between lines, nor an `obscured` one.
  It is ignored on a DateTimeInput.

Tab, ArrowRight at the value's end, the arrows, Control+n and Control+p,
Enter and Escape work them (§3.7): nothing is highlighted while the user
types, so that Enter submits what was typed; the arrows highlight one,
which Tab and the ghost then follow and Enter picks. A taken suggestion
is the value as written in `options` (`ger` takes `Germany`), and the list
shuts until the next edit. The highlight is kept by its text, so it
stays while the agent rewrites the list around it.

**Where.** Cells draws the ghost and the list (§3.5) and takes the keys. A
host draws the field with a `datalist` of the options (§2), and edits it
as any field; the ghost, the list as a surface of its own and the keys
wait for vault KIT-11h, since a host keeps Tab (SPEC §10.2) and doesn't
set a focused field's value from a program's delta (§6.2). Text shows the
field alone (§4).

### 6.23 Toasts and tooltips

**A toast** is a notice the renderer shows over a surface, in its top
right corner, for a while: a message, a kind, a time and an action at
most, as Textual's `notify` and opencode's toasts on OpenTUI. It is not a
component: it belongs to no place in the tree, outlives the components
that showed it, and an agent's prompt would have to place it somewhere and
take it out again. It is a renderer function, `hottyToast`
(`rendererOrAgent`), as `hottyFocus` is (§6.3), called by a Button's
action or by the agent (`callRendererFunction`):

| arg | |
| --- | --- |
| `message` | what it says, a line or a few; a literal or a path |
| `kind` | `info` (the default), `success`, `warning` or `error`: its mark and its colour |
| `timeout` | how long it shows, in milliseconds: 5000 by default, as Textual's; 0 keeps it until it is dismissed |
| `id` | names it: a toast shown with the id of one there takes its place, where it stands (progress, `Uploading…` then `Uploaded`), and `hottyDismissToast({id})` takes it away. Without one the renderer names it (`toast-1`, …) |
| `actionLabel`, `action` | a button on it, labelled `actionLabel`: `action` is `{"event": {"name", "context"}}`, which it sends as a Button's event, as the toast's `sourceComponentId` (`hottyToast:` and its id). The context is read as the toast shows, so `Undo` sends what was deleted even after the data model moved on. Each needs the other |

`message` and the context take literals or paths, not basic functions
(vault a2ui-limits L1: a hotty function's args call hotty functions only).
Called by a Button, the toast shows on the Button's surface, its paths in
the Button's scope; by the agent, on the surface with the keyboard, else
the first, its paths at the root, unless `surfaceId` names one.
`hottyDismissToast({id})` dismisses the toast of that id: on the caller's
surface, or on every surface for the agent's call; one that is not there
is no error.

The toasts stack, the newest nearest the corner. A surface shows five at
most: a sixth takes the oldest's place. Each counts its time down on the
rendition's clock (§3.4), by the draws that show it: the time between two
draws counts only while nothing held it, so a toast the pointer or the
keyboard is on waits (an error to read, an action to reach). Dismissing
one, by a click on it, Escape (the newest, or the one whose action has
the keyboard), or its action, sends nothing; its action sends its event.
Its action is a Tab stop, after the surface's own (and an open Modal's),
the newest toast's first; its keys are a Button's, and the keyboard goes
with the toast. The renderer keeps the toasts (`view.ToastState`), as it
keeps which tab shows, not the data model.

| rendition | a toast |
| --- | --- |
| cells | a box, rounded, its border in its kind's role, over the top right corner a column in (§3.4): `╭──────────────────────╮`, `│ ✓ Draft saved        │`, `╰──────────────────────╯`; its action at the end of its last row, `│ ✗ Could not reach the server  Retry │`. Its mark (`ⓘ ✓ ! ✗`) says its kind without colour |
| host | a box in the toasts' region over the surface and the layer (§2): its kind's icon in its colour, its message, its action a borderless button; `role=alert` for a warning or an error, `role=status` otherwise. The region is the surface's, cut at its edges: a toast in a surface of its own at a higher z (SPEC.md §5.2), which can reach past the surface, is vault KIT-13h |
| text | after the surface, after `───`: `✓ Draft saved`, `✗ Could not reach the server  [ Retry ]` (§4) |

**A tooltip** is a component's `accessibility.description`, A2UI's own
property for the words that say more than its label: no new component,
and an agent that writes descriptions for screen readers gets tooltips
for nothing. The HottyKeyHints shows it (§6.10), in a row over its keys,
as a status line: the description of the element under the pointer, else
of the element with the keyboard, an element's own or the nearest one
around it. A key, a click or a focus the user moves hides the pointer's,
until the pointer goes to another element. Without a HottyKeyHints on the
surface, there is no tooltip; a surface whose components have no
description has no row.
- **In cells** the pointer is the terminal's motion reports (mode 1003),
  which the storybook already asks for to set the pointer's shape:
  hovering costs a frame when the pointer goes onto or off an element with
  a description or a toast, and nothing else. A floating box at the
  pointer would cover what the user is reading, needs room on every side
  in a grid, and could not show for the keyboard; a fixed row does both.
- **On a host** the host reports the element under the pointer (SPEC.md
  §9.4, `hover`, to a surface placed with `v=1`), and the description of
  the element with the keyboard shows otherwise. Each element also has
  it as `aria-description`, for a screen reader. A floating tooltip, a
  surface of its own by the element (SPEC.md §5.2), is vault KIT-13h.
- **In text** it is not written: a pipe has no pointer and no focus.

### 6.24 Confirm, a pattern

A yes or no question, as huh's Confirm asks one ("Delete 3 files?"), is
no component of the hotty catalog. It is two basic Buttons and four
HottyShortcuts, as A2UI's own notification example asks with two Buttons
(catalog instruction 24; the stories hotty/confirm and
hotty/confirm-form; vault KIT-16c):

- a Text with the question, then a Row of two Buttons, Yes and No, each
  with an action of its own: the answer is an event, sent at once, as a
  Button's is;
- HottyShortcuts `y` and `n`, whose `press` is Yes and No; and
  ArrowLeft and ArrowRight, labelled alike, whose actions call
  `hottyFocus` (§6.3) with Yes and with No, so that the help line reads
  `←/→ move • y yes • n no` (§6.10);
- on the Row, `keys` (§6.5) `ArrowLeft=program ArrowRight=program`, so
  that a host, which scrolls with the arrows a focused button leaves,
  gives them to the program;
- `autofocus` (§6.4) on the answer given by default: No before what
  cannot be undone, as huh's Confirm starts on false.

| rendition | a Confirm |
| --- | --- |
| cells | the Buttons (§3.4), ` Yes   No `: the one with the keyboard reversed in the accent, the other on its grey fill, as huh fills the answer its Confirm would give. Enter or Space presses the one with the keyboard |
| host | the same Buttons, the one with the host's focus filled with `--k-focus` (§2) |
| text | the question, then `[ Yes ]  [ No ]` (§4) |

Where it is not huh's Confirm:
- **The keys are the surface's** (§6.1), not the question's. A surface
  asks one question so. A text field with the keyboard keeps y, n and the
  arrows (§3.7); from an element that leaves them (a Button elsewhere, a
  CheckBox, a HottySwitch) they answer and move. Whether a HottyShortcut
  can belong to a component is NEIO-11's open question 4.
- **The answer is an action, not a value.** Once the keyboard leaves,
  both Buttons are grey, where huh keeps its answer filled. A yes or no
  the user sets and changes is a HottySwitch (§6.17) or a CheckBox, as
  the stories hotty/switch and hotty/form hold huh's Confirm.
- **In a form** Enter in a field submits the HottyForm (§6.2), where
  huh's goes on to the next field, so Yes at the form's end sends the
  form's event.
- huh's h and l, and its Y and N, would be four more HottyShortcuts; the
  stories leave them out.

### 6.25 HottyTimer and HottyStopwatch

A time that counts, as bubbles' `timer` and `stopwatch`: a HottyTimer
counts down from its `duration` and does something when it gets to 0; a
HottyStopwatch counts up from 0. Two components, as bubbles has two
packages and a phone's clock two tabs, so that an agent finds the one it
means by its name; one view kind draws both (`view.Timer`).

| prop | |
| --- | --- |
| `label` | what it times, before the time: `Tea`, `Retrying in`, `Elapsed` |
| `duration` | a HottyTimer's, required: how long it counts down, in milliseconds (a DynamicNumber, as `hottyToast`'s `timeout` is milliseconds). A new one changes the time left, not the time counted. Bound to a path that holds nothing, the timer does not count, nor run out |
| `running` | a DynamicBoolean, true by default: whether it counts. Bound, it is the agent's handle on it, and the renderer writes it (below); unbound, the renderer keeps what the functions make it |
| `interval` | the step its time shows in, in milliseconds: 1000 by default (whole seconds, as bubbles' default), 100 for tenths. A stopwatch's time is rounded down to it, a timer's time left up, so that a timer shows 0 just as it runs out, as bubbles' does (its first second shows its whole duration) |
| `format` | `duration` (the default) writes the time as Go and so bubbles do, `4m30s`, `9.5s`, `300ms`, `0s`; `clock` as a clock, `4:30`, `0:09.5`, `1:02:03`, minutes and seconds, the hours from an hour, and as many digits after the point as the interval has, to milliseconds (`view.FormatTimer`) |
| `onTimeout` | a HottyTimer's Action, which runs once when it gets to 0 (an event, or a function such as `hottyToast`) |

**The time is the renderer's.** A2UI has no time (vault a2ui-limits L6),
so the time a timer has counted is the surface's state, as a toast's is,
not the data model's (`view.TimerState`, by its node's key), and it is
counted on the rendition's clock (§3.4) as the rendition draws
(`view.Controller.TickTimers`, before the toasts' time): since the draw
before, one that counted then has counted that much longer, and one that
runs from now counts from now. A timer so starts counting at the first
draw that has it, and a stop or a start takes effect at the draw after it,
which a program makes at once, after each key, click or message. Every
timer of the surface counts, whether the view shows it or not: one in a
tab not shown runs out on time. A timer that goes from the surface loses
its time; one that comes back starts again.

**When a HottyTimer runs out**, at the first draw at or after its end, it
shows 0 and stops: its `running` is set false, written to the data model
where it is bound, so that the agent and a Button's checks that read the
path see it stopped. Then its `onTimeout` runs, once, as the timer's
action (`sourceComponentId` the timer), without the user's activation,
so that it cannot call a function that needs it, such as `openUrl`. A
timer set running again once it has run out, by the functions or by the
agent writing `running`, starts over from its duration.

**The functions**, renderer functions a Button, a HottyShortcut or the
agent call (`rendererOrAgent`), each with `{id}`, the timer's component
id (a template's, the instance in the caller's scope; found whether the
view shows it or not), as bubbles' models have Start, Stop, Toggle and
Reset:
- `hottyStartTimer`: running true; it counts on from where it stopped.
- `hottyStopTimer`: running false; it stays where it is.
- `hottyToggleTimer`: the one of the two it is not doing, for one key or
  one Button that does both (bubbles' stopwatch example's `s`).
- `hottyResetTimer`: back to 0 counted, a HottyTimer to its whole
  duration; one that runs counts on from there, one that is stopped stays
  stopped.

Basic's Button can't write the data model, so a Start, Stop or Reset
Button calls these; its checks may read the path `running` is bound to
(basic's `not` is a Button's to call), so that Start is off while the
timer runs. `duration` and `running` themselves take a literal or a
path, not a function call (vault a2ui-limits L1). A timer takes no focus
and no keys.

| rendition | a timer |
| --- | --- |
| cells | `Tea 4m59s`, the time in `fg` while it counts and `muted` while it stands still (§3.4) |
| host | `span role=timer`, the time a text delta a tick, muted while still (§2); display type is vault KIT-14h |
| text | `Tea: 4m59s` (§4) |

### 6.26 HottyPaginator

Pages, as bubbles' paginator turns them under a list: a dot for each page,
the one shown marked, or the page shown of how many, `3/10`
(`displayStyle` `dots`, the default, or `numbers`, bubbles' `Dots` and
`Arabic`). It has two shapes, as bubbles' paginator serves two: it pages
items the program has, or tells the program which page to fetch.

- **With a `child`**, a List, Column or Row (a List of a template, for a
  list in the data model), it shows `perPage` (10) of the child's items
  at a time, the page's, with its pages under them. The view cuts the
  child's items into pages, as bubbles' program slices its items
  (`GetSliceBounds`), and the child holds the page shown alone, so that
  every rendition shows the same items and Tab goes through those alone.
  An agent sends the whole list once and the renderer pages it, with no
  round trip a turn. Any other child is one page.
- **Without one**, it shows `pages` pages (a DynamicNumber), of data the
  agent holds, such as a search's results: the agent hears of a turn by
  `onChange` and writes that page's items with `updateDataModel`. Give
  one of the two (the catalog's `oneOf`).

`page` (a DynamicNumber) is the page shown, from 1, as `3/10` reads and
as a user says it; bound to a path, the user's turns are written there,
and writing it shows that page. A literal or a missing one is the
renderer's state once turned, as an unbound Slider's value is. It is held
to the pages. A turn writes `page`, then runs `onChange`, whose context
reads that path (an action carries no payload, vault a2ui-limits L2); a
turn to the page shown, at an end, sends nothing.

It takes the keyboard: the box is a Tab stop, before the controls of its
page, as a HottyScrollView's is. ArrowLeft and ArrowRight (h and l, Page
Up and Page Down) turn a page, as bubbles' paginator's `PrevPage` and
`NextPage` keys do, and Home and End go to the first page and the last
(§3.7); its hints are bubbles' paginator example's, `←/→ page` (§6.10).
A click on a dot shows its page; a click elsewhere on it gives it the
keyboard (§3.7).

| rendition | a HottyPaginator |
| --- | --- |
| cells | its child past a field's gutter, as tall as its tallest page so that the dots stay put, a blank row, then `••••••`, the page shown's in `fg` and the others `border` and faint; `3/10` with `numbers` or where the dots do not fit; the gutter's bar down it and the page shown in `accent` while it has the keyboard (§3.4). A HottyList draws its pages with these dots |
| host | a focusable `div role=group` holding its child, then a row (`role=status`, `Page 3 of 10`) of dots that take clicks, or `3/10`; its `data-keys` give the program the arrows, Page Up, Page Down, Home and End (§2). A page is as tall as its items there. The host's own look is vault KIT-15h |
| text | its child with every page's items, then `Page 3 of 10`; that line alone without a child (§4) |

`perPage` is a literal whole number, as a HottyList's `height` is: how a
list is paged is the layout's, not the data's.

### 6.28 HottyBigText

Text in large letters, as OpenTUI's ASCIIFont draws it: a banner, a title
screen, a score or a time read from across the room. Bubble Tea has
none.

| prop | |
| --- | --- |
| `text` | required, a DynamicString: the text as it is. `\n` starts a line; a line too long for its room wraps between words. Bound, it follows the data model (a score, a countdown's time written by the agent) |
| `size` | `small`, `medium` (the default) or `large`: in cells, letters 3, 4 and 5 rows tall; on a host, display type as many rows high |
| `align` | `start` (the default), `center` or `end`: where each line goes across the room it has |

**The fonts are the kit's own**, drawn for it pixel by pixel
(`rendition/cells/bigfont.go`), not taken from figlet's or cfonts'
(ASCIIFont's are cfonts', under the GPL): capitals, the digits, as wide
as each other so that a number that changes stays put, and `!"#$%&'()*+,-./:;<=>?@[\]_`.
Medium is 5 by 7 pixels, a dot-matrix display's; small and large share
one of 3 by 5 (the round letters, N, M and W wider). Block elements
draw them, which terminals draw themselves, so that the pixels meet:
two pixels a cell in half blocks, about square, or at large a pixel two
columns of `█`. They are not figlet's ASCII art, which a terminal's
font draws with gaps and which reads as letters made of letters.

A size, not a font: the three are what an agent picks between (how
large), and on a host they are one face at three sizes. A style of
letters (outlined, shaded) would be a prop of its own.

It takes no focus and no keys.

| rendition | a HottyBigText |
| --- | --- |
| cells | block letters in `fg`, its lines a blank row apart, then a blank row (§3.3, §3.4); lowercase in capitals, a Latin letter's mark left off, `?` for a character the fonts lack |
| host | its text as it is in display type, as many rows high as cells' letters (§2); the host's own look is vault KIT-19h |
| text | its text as it is, a line each of its lines (§4) |

## 7. Fallbacks

A component never fails the surface it is in.

- **Video and AudioPlayer** are a labelled link, `▶ Video` or `▶` and the
  description: on a host a link the program opens (SPEC.md §9), in cells
  an OSC 8 hyperlink, in text the label and the URL. It takes focus, and
  Enter or Space opens it, through the renderer's `openUrl`.
- **Image** is a picture on a host, fetched only as the network policy
  allows (§2), or sent by the program in band and named by a `cid:` URL
  (SPEC.md §7.1). In cells and text it is `[image: description]`.
- **A child still to come.** A component may name one the agent has not
  sent yet, as a stream does. It shows `…` in `muted` until it comes.
- **A component that contains itself.** Where the loop closes, it shows
  `! Type` in `warning`; the renderer reports the cycle to the agent once,
  and the rest of the surface shows.
- **A component this renderer cannot draw**, of a catalog the processor
  knows but the renderer has no rendition for, shows `! Type` too. A
  component of a catalog the surface does not support is the agent's
  error: A2UI refuses the message, and the renderer reports it.
