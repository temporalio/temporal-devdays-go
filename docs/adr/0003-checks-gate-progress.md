# Checks gate progress

Every Challenge's Check fails until that Challenge's work is done, and says what
is missing and where. This carries over ADR 0008 of `temporal-devdays-ts` (and
ADR 0002 of the Java and Python repos) to every Go Workshop, the Saga Workshop
included. TypeScript's Saga Track still passes instantly; Go's does not.

A customer Dev Day's wrap-up (2026-09-15/16) ranked instant-pass Checks as its
top finding: learners clicked Check without doing the work, advanced, and hit a
failure three Challenges later with no way to tell which step they had missed.

## How a Check decides

`workshops/_sandbox/bin/workshop-check <exercise-dir> <N>` stops at the first
failure of:

1. **The code compiles.** `go vet ./...` over the Exercise, tests included. The
   learner sees the first error as `file.go line N: message`.
2. **This Challenge's tests pass.** `go test -run '^TestChallenge(1|..|N)_'`,
   against a real dev server each test package starts from the `temporal` CLI
   with `testsuite.StartDevServer`. Earlier Challenges' tests run too.
3. **The failure is specific.** The learner sees the first failing test's own
   `t.Fatal` message, which names the file, the function and the TODO.

Observe-only Challenges check the outcome instead: the Workflows the learner was
asked to run exist, in the expected state, in the lab's dev server.

Skip (`workshop-solve`) splices only that Challenge's TODOs in from the
Solution, between `// TODO N start` and `// TODO N end`, and keeps the
learner's file as `.bak`.

## Considered Options

Go's in-memory `TestWorkflowEnvironment` would make Checks faster. Rejected: it
never writes an Event History the learner can open, so the Check and the
Temporal UI would disagree about what happened.

Instant-pass Checks and grepping the source were rejected for the reasons in
the TypeScript ADR.

## Consequences

Both trees of a Workshop share one module path, so a spliced TODO compiles, and
the Exercise must compile as shipped: a TODO leaves a working stub, never a
syntax error. `scripts/test-checks.sh` proves every Check fails before its
TODOs and passes after, and `scripts/check-snippets.mjs` keeps the code shown in
assignments identical to the files.
