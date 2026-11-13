# Authoring notes for Go Workshops

The full body of Instruqt authoring lessons lives in
[`temporalio/edu-nexus-code/INSTRUQT.md`](https://github.com/temporalio/edu-nexus-code/blob/main/INSTRUQT.md).
Read that first. This file holds only what is specific to this repo.

## A Workshop teaches one SDK. Do not name the others.

A Workshop is written in Go for people working in Go. Naming another SDK puts a
comparison in front of a learner who did not ask for one, and frames Go by what
it lacks or has. Write the Go design as the design. Sweep before publishing:

```bash
grep -rn "Java\|Kotlin\|Python\|TypeScript\|\.NET" workshops/<slug>/ decks/ --include="*.go" --include="*.md"
```

## Go-specific rules

- **Same module path in both trees.** `exercise/go.mod` and `solution/go.mod`
  declare the same module, so Skip's spliced TODO compiles and the trees diff
  clean.
- **The Exercise compiles as shipped.** A TODO is a working stub or the bug the
  Challenge fixes. `go vet` failing on an untouched Exercise is a Track bug.
- **Commands the learner types** are `go run ./worker` and `go run ./starter`
  from the Exercise folder.
- **A Crash is a signal to the Worker binary, not to `go run`.** `go run` forks
  a child; kill the child by name, or the Worker survives its own crash.
- **Track slugs end in `-go`**, and every Track uses the one Sandbox Image
  (ADR 0004), pinned by digest, never `:latest`.
- **New Tracks have no `id:` or `checksum:`** in `track.yml`. The first push
  writes them; commit what it writes.
