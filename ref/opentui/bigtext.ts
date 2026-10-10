// story: hotty/bigtext
//
// The reference for HottyBigText (vault KIT-19): OpenTUI's ASCIIFont, with
// the content of the kit's story hotty/bigtext, so that a shot of one can be
// set beside a shot of the other (scripts/ref-shot.sh): HOTTY in its
// "block" font, centred over a caption, the score in "tiny", and the
// sizes' line in "tiny" (2 rows) and "block" (6 rows, with a shadow); the
// wrapped line has no equivalent, since ASCIIFont does not wrap: a line
// wider than the terminal is cut.
//
//	bun ref/opentui/bigtext.ts      Control+C quits
//
// Bubble Tea has no big text, so this reference is OpenTUI's. Its fonts are
// cfonts' (GPL-3.0), which the kit neither copies nor ships: it draws its
// own (rendition/cells/bigfont.go).
import { ASCIIFontRenderable, BoxRenderable, TextRenderable, createCliRenderer } from "@opentui/core"

const renderer = await createCliRenderer({ exitOnCtrlC: true })
const page = new BoxRenderable(renderer, { flexDirection: "column", width: "100%", height: "100%" })
const text = (content: string) => new TextRenderable(renderer, { content })
const big = (text: string, font: "tiny" | "block") => new ASCIIFontRenderable(renderer, { text, font })

page.add(text("Big text"))
const splash = new BoxRenderable(renderer, { flexDirection: "column", alignItems: "center", width: "100%" })
splash.add(big("HOTTY", "block"))
splash.add(text("Web pages in a terminal"))
page.add(splash)
page.add(text("Score, bound to the data model"))
page.add(big("3 : 1", "tiny"))
page.add(text("Sizes: tiny and block"))
page.add(big("Ready? Go!", "tiny"))
page.add(big("Ready? Go!", "block"))
renderer.root.add(page)
