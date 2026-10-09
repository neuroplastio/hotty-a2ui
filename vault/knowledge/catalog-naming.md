# Catalog naming (was Q-0001)

Status: answered 2026-10-09: option A, the `Hotty` prefix ("Hotty prefix
sounds good"). Applied in KIT-00.

## The rule

- Every component of the hotty catalog is `Hotty` and the plain noun A2UI
  would use: `HottyTable`, `HottyProgress`.
- Every function is `hotty` and the verb: `hottyFocus`, `hottyScrollTo`.
- Extension keys stay under `io_neuroplast_hotty`.
- The catalog id stays `https://neuroplast.io/hotty/a2ui/v1/catalog.json`.
- When A2UI adds a component that does what one of ours does, the renderer
  maps theirs onto the same view kind, and ours gets `deprecated` and
  `x-deprecated-reason` (or both stay, if they differ).

profile.md §6 states the rule, as does the catalog's `instructions`.

The question as it was asked follows.

## What I need decided

A2UI may later add a component under the same name as one of ours. Which
naming rule keeps the kit's new components and functions clear of that? My
lean is option A below.

## Why I cannot decide it

Names are what agents write, and what the stories, vectors and profile spell
out. A rename today touches four names. Once twenty components exist, and
agents' prompts and saved surfaces carry them, the same rename becomes a
migration.

## What I found

- **A name is scoped by its catalog.** A2UI v1.0 first looks for a
  component's `catalogId`, then the surface's default, and doesn't fall back
  to anything else. The renderer here maps components by catalog and type
  (`view/build.go`, `mappers[catalog][type]`). A `Table` in basic and a
  `Table` in hotty are both legal, and the renderer tells them apart.
- **A shared name still hurts in three places.**
  1. A prompt that holds both catalogs: a model fills one `Table`'s props into
     the other, or leaves out `catalogId`. The component then resolves to the
     surface default's `Table`, which gives a schema error at best and a
     valid but different component at worst.
  2. A composite catalog (KIT-CAT, a2ui-limits L1) is one map of names, so it
     can't hold both.
  3. Models learn A2UI's own components. Once basic has a `Table`, that is
     what "Table" means to them.
- **Basic is mostly closed to new components.** The maintainers declined a
  ProgressBar (upstream #1992): "In general, we're not looking to add more
  components to the basic catalog (unless very strictly needed); instead, we
  believe people should be creating their own custom catalogs."
- **New official components come in catalogs of their own.** Examples: `mcp`
  (functions), and the iframe catalog's `WebAppFrameUrl` and
  `WebAppFrameSrcdoc` (upstream #2798). A file upload component (#534) is
  proposed. Each catalog is versioned on its own (`catalogs/basic/v1/`).
- **Precedent.** The Gemini Enterprise composite catalog, an upstream sample,
  keeps basic's names as they are and prefixes its own: `MaterialTable`,
  `MaterialProgressBar`, `MaterialProgressSpinner`, `GcbpTable`, `VegaChart`.
- **What A2UI reserves.** The `Surface` component name, the `a2ui_` extension
  prefix and `@` keys. It reserves no component prefix, so a prefix of ours
  can't collide.

## Options

**A. A prefix on every kit component and function.**
- Examples: `HottyTable`, `HottyProgress`, `HottySpinner`, `HottyCode`,
  `HottyDiff`, `HottyKeyHints`, and functions `hottyScrollTo`, `hottyCopy`.
- The existing four become `HottyForm`, `HottyShortcut`, `hottyFocus` and
  `hottyBlur`.
- Pros:
  - It can never collide.
  - It tells a model which `catalogId` to write.
  - It keeps a composite catalog safe.
  - When A2UI adds an equivalent, the renderer maps theirs onto the same view
    kind and ours gets `deprecated` and `x-deprecated-reason`, or both stay if
    they differ.
- Costs:
  - Longer names.
  - A one-time rename across profile §6, the catalog's instructions, the
    stories and the vectors. hotty-demo spells none of these names; it only
    needs a version bump.
- `Hotty` rather than `Kit` or `Tty`, because it names the profile that says
  how they draw.

**B. Plain names, renamed on a collision.**
- Examples: `Table`, `Progress`, `Spinner`.
- Pros: natural, and models' habits from A2UI carry over.
- Costs:
  - When an official catalog takes a name, we rename then, which breaks
    agents' prompts and saved surfaces.
  - A composite catalog has to leave out whichever one loses.

**C. Descriptive names without a prefix.**
- Examples: `DataTable`, `ActivityIndicator`, `CodeView`.
- They lower the odds but guarantee nothing. They also read unevenly, since
  some nouns have no natural second word.

## My lean

**A**, applied to components and functions alike.
- The catalog id stays `https://neuroplast.io/hotty/a2ui/v1/catalog.json`.
  Nothing has been released under it, so the rename doesn't need a v2.
- Extension keys stay under `io_neuroplast_hotty`, as A2UI asks.
- The rule for later: a new component is `Hotty` followed by the plain noun
  A2UI would use, and a new function is `hotty` followed by the verb.
