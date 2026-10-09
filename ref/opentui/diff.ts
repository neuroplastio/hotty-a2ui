// story: hotty/diff
//
// The reference for HottyDiff (vault KIT-06): OpenTUI's DiffRenderable,
// with the content of the kit's story hotty/diff, so that a shot of one
// can be set beside a shot of the other (scripts/ref-shot.sh). It takes a
// patch of one file, and has no file names, word marks, folds or
// selection: the story's first file, unified, then its Python change,
// split.
//
//	bun ref/opentui/diff.ts      Control+C quits
//
// Bubble Tea has no diff, so this reference is OpenTUI's, a Bun package of
// its own (package.json), out of the kit's go.mod and ref/'s.
import { BoxRenderable, DiffRenderable, SyntaxStyle, TextRenderable, createCliRenderer } from "@opentui/core"

const go = "diff --git a/api/handler.go b/api/handler.go\nindex 3b18e51..a9c2f4d 100644\n--- a/api/handler.go\n+++ b/api/handler.go\n@@ -1,7 +1,7 @@\n package api\n \n import (\n-\t\"errors\"\n+\t\"fmt\"\n \t\"net/http\"\n )\n \n@@ -18,9 +18,10 @@ type Server struct {\n // Handle serves a request, or says why it cannot.\n func Handle(w http.ResponseWriter, r *http.Request) error {\n-\tif r.Method != http.MethodGet {\n-\t\treturn errors.New(\"only GET\")\n+\tif r.Method != http.MethodGet && r.Method != http.MethodHead {\n+\t\treturn fmt.Errorf(\"only GET or HEAD, not %s\", r.Method)\n \t}\n+\tw.Header().Set(\"Content-Type\", \"text/plain\")\n \tw.WriteHeader(http.StatusOK)\n \t_, err := w.Write([]byte(\"ok\\n\"))\n \treturn err\n }\n"

const py = "--- a/fetch.py\n+++ b/fetch.py\n@@ -1,10 +1,12 @@\n+import time\n+\n import requests\n \n \n-def fetch(url, tries=3):\n+def fetch(url, tries=3, backoff=0.5):\n     for attempt in range(tries):\n         try:\n             return requests.get(url, timeout=5)\n         except requests.ConnectionError:\n-            pass\n-    raise RuntimeError(f\"gave up on {url}\")\n+            time.sleep(backoff * 2**attempt)\n+    raise RuntimeError(f\"gave up on {url} after {tries} tries\")\n"

// The kit's roles for tokens (profile §3.4), in Catppuccin Mocha, the
// shot's theme: keywords and types info, strings success, numbers warning,
// comments muted.
const syntaxStyle = SyntaxStyle.fromStyles({
  keyword: { fg: "#89dceb", bold: true },
  type: { fg: "#89dceb" },
  function: { bold: true },
  string: { fg: "#a6e3a1" },
  number: { fg: "#f9e2af" },
  constant: { fg: "#f9e2af" },
  comment: { fg: "#7f849c", italic: true },
})

const renderer = await createCliRenderer({ exitOnCtrlC: true })
const page = new BoxRenderable(renderer, { flexDirection: "column", width: "100%", height: "100%" })
const lines = (patch: string) => patch.split("\n").length
page.add(new DiffRenderable(renderer, { diff: go, filetype: "go", syntaxStyle, view: "unified", height: lines(go) - 3 }))
page.add(new TextRenderable(renderer, { content: "Retry with backoff, split", marginTop: 1 }))
page.add(new DiffRenderable(renderer, { diff: py, filetype: "python", syntaxStyle, view: "split", height: lines(py) - 1 }))
renderer.root.add(page)
