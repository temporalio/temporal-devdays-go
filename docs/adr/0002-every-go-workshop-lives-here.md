# Every Go Workshop lives here, including Nexus

The TypeScript and Java repos send a Workshop to its topic repo when one
exists and already holds other languages; that is why the TypeScript Nexus code
and Track live in `temporalio/edu-nexus-code/typescript/`. That rule existed to
avoid duplicating work that was already in place, not as a principle.

For Go there is nothing in place to reuse, so the Nexus Workshop's Exercise,
Solution, Track and Deck all live here in `workshops/nexus-decouple-monolith/`
and `decks/nexus-decouple-monolith/`, beside every other Go Workshop.

## Considered Options

Adding `go/` to `edu-nexus-code` was the rule the sibling repos follow.
Rejected: it would split one Go program across two repos, two CIs and two
Sandbox Images, to sit next to languages a Go learner never sees.

## Consequences

This repo is self-contained: every Go Workshop's code, Track, Sandbox Image and
Deck is here. `edu-nexus-code` is a topic repo without a Go sibling, and that is
deliberate. A reader comparing the TypeScript README will find Nexus in a
different place, and that is correct.
