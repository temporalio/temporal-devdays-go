# Deck snippets

The code slides from `../segments/`, as Go you can run. Each package holds one
topic's slides between `// slide:` and `// end slide` markers, plus whatever
scaffolding it needs to run (types, stub Activities, a Worker):

```go
// slide: 09-signals.md "Go Signals"
...the code exactly as the slide shows it...
// end slide
```

The quoted anchor is any text in the segment before the code block, usually the
walkthrough's `heading:`. The region, with its common indentation removed, must
equal the first `go` block after the anchor. A region may cover a whole file
(start the marker above the `package` clause) or only the declarations or
statements the slide shows.

A slide that shows several files in one block (an Activity, then the Workflow
that calls it) is split into parts, each in the file that compiles it:

```go
// part: 04-design-patterns.md "Fan out" 1
...
// end part
```

The block must equal the parts in order, joined by one blank line.

```bash
go vet ./...
go test -count=1 ./...      # needs the temporal CLI on PATH
```

What the tests check:

- **`go vet ./...`** compiles and vets every package against
  `go.temporal.io/sdk` v1.51.0.
- **`TestSlidesMatch`** (`deck_test.go`) fails when a slide and its package
  drift apart: the code between the markers must equal the slide's block
  exactly. Edit one, edit the other.
- **`TestPartsMatch`** does the same for slides assembled from parts.
- **`TestSlidesParse`** parses every Go block in the Deck with `go/parser`,
  including fragments no package runs. A fragment is tried as a whole file,
  as declarations under a `package` clause, as a function body, as composite
  literal or struct fields, and as `case` clauses.
- **`TestGofmt`** keeps every file gofmt-clean, so the slides are too.
- **`TestWalkthroughs`** checks every `code-walkthrough` slide's bubbles
  against its highlights: one note per step, no step starting or ending on a
  blank line, and any `<code>` a note names must be on a lit line. When a note
  mentions code that is deliberately elsewhere, write it
  `<code class="ref">run</code>`: it looks the same and is not checked.
- **The runtime tests** (`<package>/*_test.go`) run the code against a real
  dev server that each package starts once from the `temporal` CLI
  (`testsuite.StartDevServer` with `ExistingPath`, see `internal/testenv`), or
  against the SDK's in-memory test environment where a slide waits for minutes
  of durable time, and assert what the speaker notes claim.

The tests prove the right lines are lit. To prove the rendered bubble sits on
them, run the browser check against a dev server from `decks/go-master`:

```bash
npm run dev                    # terminal 1
npm run check:walkthroughs     # terminal 2; SLIDEV_URL=... if not on :3030
```

Both run in CI on any change to the Deck
(`.github/workflows/go-master-deck.yml`).
