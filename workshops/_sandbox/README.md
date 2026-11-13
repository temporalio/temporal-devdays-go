# Sandbox Image and Check helpers

One image for every Go Track (ADR 0004), and the gated Check and Skip helpers
(ADR 0003).

- `Dockerfile` builds `ghcr.io/temporalio/temporal-devdays-go-sandbox` from the
  `workshops/` folder as build context: Go, the Temporal CLI, gopls, code-server
  with the Go extension, the helpers below, and every Workshop's `exercise/` and
  `solution/` at `/root/workshop/<slug>/`, with modules and build cache warm.
  The build fails if any Solution's tests fail or any Exercise does not compile.
- `bin/workshop-check <exercise-dir> <N>` is the Check of a coding Challenge:
  `go vet ./...`, then the tests named `TestChallenge1_...` to
  `TestChallengeN_...`, then the first failing test's own message.
- `workshop-solve/` builds `workshop-solve <exercise-dir> <solution-dir> <file>
  "TODO N" ...`, which is Skip: it splices only those TODOs in from the Solution
  and keeps the learner's file as `.bak`.
- A Workshop's own `sandbox/bin/`, when it has one, is installed into
  `/usr/local/bin` too, e.g. an Event History reader behind an observe-only Check.

## Rules for a Workshop's code

- `exercise/` and `solution/` are each a Go module with the **same module
  path**, so a TODO spliced from one compiles in the other and the two trees
  can be diffed.
- The Exercise compiles as shipped. A TODO leaves a working stub (returning a
  zero value, or the bug the Challenge fixes), never a syntax error.
- Tests are named `TestChallenge<N>_<What>`. Their failure messages are for
  learners: they name the file, the function and the TODO, **on one line**.
  The Check shows only the first line of a failing test's message.
- Tests run against a real dev server, started from the `temporal` binary on
  `PATH` with `testsuite.StartDevServer`.

```bash
cd workshops
docker buildx build --platform linux/arm64 -f _sandbox/Dockerfile \
  -t temporal-devdays-go-sandbox:local --load .                     # test locally
docker buildx build --platform linux/amd64 -f _sandbox/Dockerfile --build-arg PROVE=0 \
  -t ghcr.io/temporalio/temporal-devdays-go-sandbox:latest --push .  # publish
```
