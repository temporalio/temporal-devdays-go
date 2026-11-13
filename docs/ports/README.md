# Port cost records

One file per Port, `docs/ports/<workshop>-go.md`, written at the end of it.

The slug is the **working directory**, so every Port is worked from this repo.
Transcripts are local and per-machine, so this only measures Ports done on this
laptop, and only while the transcripts survive: **run the script before the
session's transcripts age out**.

Several Ports in one session are told apart by how they were run:

- **Delegated to a subagent** (the Go Ports were, in parallel): `--agent <id>`
  costs only that subagent's transcripts. Record the agent id in
  [`LOG.md`](LOG.md) when the agent is launched.
- **Done in the main session, one after another**: `--since`/`--until` with the
  timestamps logged in `LOG.md`.

```sh
node scripts/port-cost.mjs -Users-nikk-source-temporal-go-temporal-devdays-go \
  --agent <agent-id> --label "Child Workflows: TypeScript -> Go" \
  > docs/ports/child-workflows-go.md
```

The main session's own work (scaffold, image builds, Track pushes, merging) is
shared across all Ports and is recorded once, in `go-devdays-session.md`, with
`--scope main`.
