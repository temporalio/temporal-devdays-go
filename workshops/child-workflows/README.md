# Child Workflows (Go)

**Level:** Foundations. Requires Temporal Essentials.
**Duration:** 15 minutes of lab, after the Child Workflows section of the Deck.
**Track:** `child-workflows-go` (not yet published)

Run a child, chain children five deep, fan out, fan in. Challenges 1 to 3 are
run and observe: the learner runs one starter command and reads the result in
the Temporal UI. Challenge 4 is the one TODO: write the fan in.

| # | Challenge | Learner does | Check passes when |
|---|---|---|---|
| 1 | Run a child | `go run ./starter child` | a `ShipmentWorkflow` child has Completed |
| 2 | A deeply nested chain | `go run ./starter chain --fail` | six `ChainWorkflow`s have Failed (one five-level chain) |
| 3 | Fan out | `go run ./starter pack`, TODO 1 still undone | five `PackItemWorkflow`s were Terminated |
| 4 | Fan in | TODO 1, restart the Worker, then `go run ./starter pack` | the `TestChallenge1_`..`TestChallenge4_` tests pass, and five `PackItemWorkflow`s have Completed |

Challenge 3 is meant to look broken: with TODO 1 undone, `PackWorkflow` returns
without waiting, and `PARENT_CLOSE_POLICY_TERMINATE` (the default, written out
in the child options) kills all five children. Challenge 4 fixes that same
function by calling `Get` on every child's Future.

## The Go design

- `workflow.ExecuteChildWorkflow` starts a child and returns a
  `ChildWorkflowFuture` at once. `Get` waits for the result; that is the whole
  of Challenge 1's `OrderWorkflow`.
- `workflow.ChildWorkflowOptions`, set on the Context with
  `workflow.WithChildOptions`, gives each child its `WorkflowID` and, in
  `PackWorkflow`, its `ParentClosePolicy`.
- Fan out keeps one Future per item and waits on
  `GetChildWorkflowExecution()` only, so every child has started before the
  parent goes on. Fan in calls `Get` on each Future in item order; the children
  already run in parallel, so the loop takes as long as the slowest one.
- A child's failure reaches its parent as a `*temporal.ChildWorkflowExecutionError`
  wrapping the child's own error. Five levels deep, the starter unwraps five
  of them to reach the `*temporal.ApplicationError` of type `BottomOfChain`.

## Layout

```
exercise/        the learner's terminals start here; TODO 1 undone
  TODOS.md       every TODO, in order
  workflows/     every Workflow, and workflows_test.go: TestChallenge<N>_... per Challenge
  worker/        go run ./worker
  starter/       go run ./starter <child|chain|pack>
solution/        the proven code, identical except inside TODO 1
instruqt/        the Track: track.yml, config.yml, track_scripts/, 01..04/
sandbox/
  challenges.tsv the coding Challenge and the TODO its Skip splices in
```

Both trees are the module
`github.com/temporalio/temporal-devdays-go/workshops/child-workflows`. The tests
start their own dev server from the `temporal` CLI on the `PATH`, with
`testsuite.StartDevServer`, and run one Worker for every test. The Sandbox
Image is built from [`../_sandbox/`](../_sandbox/). Task Queue:
`child-workflows`.

## Run it locally

Go 1.26 and the [Temporal CLI](https://docs.temporal.io/cli).

```bash
temporal server start-dev                       # terminal 1
cd workshops/child-workflows/exercise           # or solution
go run ./worker                                 # terminal 2; Ctrl+C and rerun after a change
go run ./starter child                          # terminal 3
go run ./starter chain --fail
go run ./starter chain --depth 3
go run ./starter pack
go test ./...                                   # starts its own dev server
```

## Prove it

From the repo root:

```bash
(cd workshops/child-workflows/solution && go vet ./... && go test -count=1 ./...)   # Solution is proven
(cd workshops/child-workflows/exercise && go vet ./...)                            # Exercise compiles as shipped
bash scripts/test-checks.sh workshops/child-workflows                              # Challenge 4's Check gates TODO 1
node scripts/check-snippets.mjs workshops/child-workflows                          # assignment snippets match the code

cd workshops && docker buildx build --platform linux/arm64 -f _sandbox/Dockerfile \
  -t temporal-devdays-go-sandbox:local --load . && cd ..
bash scripts/test-track-local.sh workshops/child-workflows temporal-devdays-go-sandbox:local
```

## Instruqt Track

Every Challenge has the same four tabs, so tab buttons never shift:

| Tab | Type | Target |
|---|---|---|
| `tab-0` Exercise | service | code-server on 8080, `/?folder=/root/workshop/child-workflows` |
| `tab-1` Worker | terminal | `exercise/` |
| `tab-2` Client | terminal | `exercise/` |
| `tab-3` Temporal UI | service | 8233 |

`config.yml` names `ghcr.io/temporalio/temporal-devdays-go-sandbox:latest` until
publish, when it is pinned to the image's digest. `track.yml` has no `id:` or
`checksum:` yet: the first `instruqt track push` writes them, and those are
committed.

## Known issues

- The observe-only Checks count Workflows in the dev server, so they pass for
  anyone who ran the starter, even from `solution/`. That is the point of an
  observe-only Check.
- Challenge 3 relies on TODO 1 being undone. A learner who writes the fan in
  early gets five Completed children instead of five Terminated, and
  Challenge 3's Check (and its Skip) cannot pass; undo TODO 1 or restore
  `workflows/pack.go` from its `.bak` to run it.
- A running Worker keeps the code it was built from. After TODO 1 the learner
  must stop the Worker with `Ctrl+C` and run `go run ./worker` again, or the
  old code terminates the children again. Challenge 4's Check says so when
  the tests pass but no child has completed.
- Challenge 4's Skip stops any running Worker (by its binary name, `worker`),
  runs the starter on a fresh one, and stops that too.
- Each Challenge opens fresh terminals, so the Worker does not survive into
  the next one. Every assignment starts it again.
