# Working in temporal-devdays-go

The Go half of the DevDays program. Vocabulary is in `CONTEXT.md`; use its terms
exactly (Workshop, Port, Track, Challenge, Check, Skip, Exercise, Solution,
Deck, Sandbox Image, Social Invite). Decisions are in `docs/adr/`.

- Go only (ADR 0001). Every Go Workshop, Nexus included, lives here (ADR 0002).
- Checks gate progress (ADR 0003). Read `workshops/_sandbox/README.md` before
  writing a Workshop: module, TODO-marker and test-naming rules live there.
- One Sandbox Image for every Track (ADR 0004).
- Instruqt rules: `INSTRUQT.md`.
- Every Port gets a cost record: `docs/ports/README.md`.

Prove a Workshop before calling it done:

```bash
(cd workshops/<slug>/solution && go test -count=1 ./...)
(cd workshops/<slug>/exercise && go vet ./...)
scripts/test-checks.sh workshops/<slug>
node scripts/check-snippets.mjs workshops/<slug>
```
