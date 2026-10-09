# Vault

What an agent needs to move the kit forward, and what the maintainer needs to
steer it. Start here, then read [`/AGENTS.md`](../AGENTS.md) and
[`docs/profile.md`](../docs/profile.md).

## Map

| Path | What it is |
| --- | --- |
| [`roadmap.md`](roadmap.md) | The phases in order, with exit criteria |
| [`tasks/board.md`](tasks/board.md) | The live board: tickets in the order they are built |
| [`feedback/`](feedback/) | **Human → agent.** Steers the work. Preempts the board. |
| [`questions/`](questions/) | **Agent → human.** Decisions the agent must not make. Blocks work. |
| [`journal/`](journal/) | One entry per leg: what changed, how to see it. |
| [`knowledge/gap-analysis.md`](knowledge/gap-analysis.md) | The kit against Bubble Tea and OpenTUI: every gap, with its ticket |
| [`knowledge/a2ui-limits.md`](knowledge/a2ui-limits.md) | What A2UI v1.0 can and cannot carry for those components |

## The loop

An agent works one leg at a time: read `feedback/`, then the board; take the
first open ticket that nothing blocks; build it; check it (`make check`, the
storybook relaunched and looked at); write a journal entry; tick the board;
push. A leg that changes how something looks stops there, so the maintainer is
never more than one visual change behind.

The maintainer steers with a file in `feedback/`, unblocks by answering a file
in `questions/` (`Status: answered`), and judges a leg by running the storybook
(`go run ./cmd/storybook`, or <https://hotty.neuroplast.io/storybook>).
