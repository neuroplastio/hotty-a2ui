# What A2UI v1.0 can carry for the kit's new components

As of 2026-10-09: A2UI at `third_party/a2ui/REV` (ae466ff). The upstream
repository (a2ui-project/a2ui) was read on the same day, and nothing was
posted there.

**Verdict.** No gap in [gap-analysis.md](gap-analysis.md) needs A2UI to
change. Every component fits as a catalog component with its own props,
`Action` props, renderer functions and `io_neuroplast_hotty` extensions, which
are the extension points AGENTS.md allows. Four limits shape how the
components are designed (L2, L4, L5 and L10). The rest are notes.

## The limits

**L1. A catalog can't import another.** A catalog's `$ref`s reach only its own
components and functions and `common_types.json` (protocol, *Catalog Schema
Rules* 3). Mixing is per component: a surface has one default catalog, and
every component of another catalog carries its `catalogId`.
- Effect: an agent writes `catalogId` on every hotty component, as the stories
  do now.
- Fix: a composite catalog, generated from basic's definitions verbatim plus
  ours, for an agent to use as the surface default. Upstream's own samples do
  this (`basic_with_mcp_catalog.json`, Gemini Enterprise's composite). It's
  ticket KIT-CAT on the board, for when the agent integration starts.

**L2. An action carries no event payload from the renderer.**
- An `Action`'s `event.context` holds literals, data-model paths and function
  calls, evaluated in the component's scope. The `action` message adds only
  `sourceComponentId` and a timestamp.
- So there is no `event.detail`: nothing tells the agent which row, which
  hunk, which suggestion or which key.
- There are two ways round it:
  - **Composed rows.** A Button in a template row resolves relative paths in
    that row's scope, so `{"@path": "id"}` is the row's id. This is good for
    small lists with rich rows.
  - **Select, then act.** A data-driven component writes the item it acts on
    into a bound path (`selected`, `value`) before it runs its action, and the
    action's context reads that path. ChoicePicker already works this way.
- Effect: KIT-01, KIT-04, KIT-06, KIT-09, KIT-11 and KIT-17 each declare a
  bound "current item" prop and say in the catalog instructions that their
  actions read it.

**L3. Templates make one component instance per item.** A `ChildList`
template creates a node for each element of the array, and A2UI has no
notion of virtualization. The core also caps list writes at index 10,000
(`a2ui.MaxListIndex`).
- Design rule: a big collection is data on a data-driven component, which
  the renderer windows (Table, Rich list, Tree, Scroll view). Templates are
  for small collections with composed rows.
- Recursive templates work: a scope that differs per level isn't a cycle, up
  to `a2ui.MaxDepth` (50). That suits a small Tree.

**L4. The data model only replaces or deletes the value at a path.**
`updateDataModel` has no append, insert, splice or text-append, and doesn't
define RFC 6901's `-` (the end of an array).
- Effect on streaming content (a log, a terminal, an agent's output, chart
  points):
  - It arrives as an upsert at the next index, so the agent tracks the
    length. The core pads with nulls past the end, up to 10,000.
  - Or the agent sends the whole value again.
  - Each upsert of a long string sends all of it again.
- Effect: KIT-07, KIT-10 and KIT-18 take a list of chunks or lines, not one
  growing string, and they rotate at a bound.

**L5. `sendDataModel` sends the whole model with every action.** With
`sendDataModel: true`, every click carries the surface's entire data model
back to the agent.
- Effect: big read-only content shouldn't sit in the data model. That covers
  a diff, a file's code, a table's thousand rows, and a log.
- It goes in literal props sent with `updateComponents`, or the surface keeps
  `sendDataModel` off and each action names what it needs.
- The catalog instructions for KIT-01, KIT-05, KIT-06 and KIT-07 say so.

**L6. The protocol has no time.** Nothing in A2UI ticks, and actions are
interaction handlers.
- A spinner's frames, a timer's seconds and a toast's timeout are the
  renderer's presentation, so that isn't a limit.
- A timer's `onTimeout` is an `Action` prop whose trigger the catalog
  defines, which A2UI allows (any component may have several `Action` props;
  see upstream #1389).
- A2UI asks the renderer to show a *pending* state while an async function
  value resolves (protocol, *Functions in A2UI Content Execution* §2). KIT-03
  gives that state its look.

**L7. Styling is limited to weight, justify and align.** v1.0 removed `theme`
"to separate layout from branding", and basic's components have closed props.
- The kit's own components define what they need (column widths, alignment,
  tone).
- A hint on a basic component goes in `metadata.extensions.io_neuroplast_hotty`.
  The protocol names "component-instance styling overrides" as a use of
  component extensions.
- The look itself is the renderer's theme. That is where "looks like Bubble
  Tea" lives anyway.

**L8. The agent steers the view only through renderer functions.**
- Scrolling to an item, focusing, copying to the clipboard and dismissing a
  toast are renderer functions called with `callRendererFunction`, guarded by
  `allowedCallers` (and `requiresUserActivation` for copy), as `focus` and
  `blur` are now.
- This is supported, so it isn't a limit.

**L9. The agent doesn't know the rendition.** Capabilities list catalogs only.
Nothing tells an agent "cells, 80×24" or "a HOTTY host".
- This is by design, and it matches the profile: one tree, three renditions.
- Every component has to make sense in cells, so phase 1 is the floor.

**L10. A2UI has nothing to say about the user's machine.** A file picker and a
terminal run where the renderer runs, which is the user's machine, while the
agent may be remote. Function guards (`allowedCallers`,
`requiresUserActivation`) exist, but no resource scopes do. The profile has to
set the boundary:
- **File picker.** It lists only roots the program grants. A picked path
  reaches the agent only through the data model, with the user's next action.
- **Terminal.** It runs only commands the program has registered, which the
  agent names by id. It never runs a command line from the agent.

**L11. Composition rules exist and are worth using.** `allowedParents` and
`allowedChildren` (with `Surface` as the root) let a catalog say what a
composed Table or Tree may contain. A renderer reports `UNALLOWED_PARENT` and
`UNALLOWED_CHILD`.

**L12. An Icon is one of 59 names, or one filled path.** Basic's `name` is
an enum or `{svgPath}` (added 2026-10-10, gov R-4).
- **Names:** A2UI's renderers map the 59 to the Material Symbols font by
  snake_case and four overrides. In Iconify's data, which the kit draws
  from, two of them are the font's other names for a glyph: `payment` is
  `credit_card` and `phone` is `call` (cmd/iconsgen).
- **svgPath:** one path in a 24 box, filled, nonzero. A stroke icon must be
  outlined first. SwiftUI drops arcs, and Flutter (genui) has no `svgPath`.
- **Beyond the 59:** any other name needs a component of the hotty catalog
  (`HottyIcon`, profile §6.15, journal 2026-10-10.3), since basic's `name`
  is an enum and a basic component isn't forked. Its `svgPath` takes a
  `strokeWidth`, so a stroke icon needn't be outlined.

## Per component

| Ticket | What it needs from A2UI | Fit |
| --- | --- | --- |
| KIT-01 Table | `columns` (header, key, width, align), `rows` as data (or composed rows), a bound `selected`, `onActivate`, sorting as renderer state | Fits. L2, L3, L5 |
| KIT-02 Progress | `value` (DynamicNumber), optional range, label; indeterminate when there's no value | Fits |
| KIT-03 Spinner | `label`, `active` (DynamicBoolean), a frame set | Fits. L6 |
| KIT-04 Rich list | `items` (label, description, value), `filterable`, bound `selected`, `onActivate` | Fits. L2, L3 |
| KIT-05 Code | `text`, `language`, `lineNumbers`, marked lines, wrap | Fits. L5 |
| KIT-06 Diff | `patch`, or `old` and `new`, plus `language`, `view` (unified, split), bound `hunk`, `onHunk` | Fits. L2, L5 |
| KIT-07 Scroll view | a child, or `lines` as data; `follow`; a `scrollTo` renderer function | Fits. L4, L5, L8 |
| KIT-08 Key hints | nothing beyond the surface's Shortcuts and keymaps | Fits |
| KIT-09 Tree | `items` with `children`, or a recursive template; bound `selected` and `expanded` | Fits. L2, L3 |
| KIT-10 Chart | `series` as data, `kind`, axes | Fits. L4 |
| KIT-11 Suggestions | `suggestions`, a DynamicStringList or an agent function. A2UI evaluates the function again as the field changes, and shows it pending meanwhile | Fits. L2, L6 |
| KIT-12 Selection and copy | nothing: renderer behaviour; a `copy` renderer function if the agent should copy | Fits. L8 |
| KIT-13 Toast and tooltip | toast: `message`, `tone`, `duration`, bound `open`; tooltip: `accessibility.description` | Fits. L6 |
| KIT-14 Timer and stopwatch | `start` or `duration`, bound `running`, `onTimeout` | Fits. L6 |
| KIT-15 Paginator | bound `page`, `pages` | Fits |
| KIT-16 Confirm | bound `value`, labels | Fits |
| KIT-17 File picker | `root` (a grant's name), bound `path`, `filter` | Fits. **L10** |
| KIT-18 Terminal | `command` (a registered id), `lines` or a stream | Fits. L4, **L10** |
| KIT-19 Big text | `text`, `font` | Fits |
| KIT-20 QR code | `value` | Fits |

The prop names above are sketches. Each phase-1 leg settles its component's
props, under `Hotty` names ([catalog-naming](catalog-naming.md)).
