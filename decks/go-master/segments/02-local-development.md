---
layout: section
---

<ZiggyHero scene="demo">

# Demo

</ZiggyHero>

<!--
Source: PPT slide 20. Carried over from the original Essentials Deck.

Run a Workflow live if you have one ready: start a Worker, start a few
Workflows, and open the Web UI at localhost:8233. Kill the Worker mid-run and
start it again. The Workflow carries on from where it stopped. That is the
whole pitch in thirty seconds, and it lands harder than any slide.

If you have nothing prepared, the hello-world project on the next slide is
enough: add a `workflow.Sleep(ctx, 30*time.Second)` to its Workflow, kill the Worker during the
sleep, restart it, and watch the Workflow finish.
-->

---
layout: default
---

# Local development

```bash
brew install temporal
temporal server start-dev

git clone https://github.com/temporalio/samples-go && cd samples-go
go run ./helloworld/worker    # the Worker
go run ./helloworld/starter   # in a second terminal: start a Workflow
```

- **MIT-licensed open source.** Same server here, on-prem, or in Temporal Cloud
- **One binary**, with the Web UI at `localhost:8233`
- **Self-hosted state** in PostgreSQL, MySQL or Cassandra
- **Workers connect out** to the server, from wherever they run

<!--
Source: PPT slide 21. Carried over from the original Essentials Deck.

The original listed MySQL as the default store and "running locally in Docker
containers." `temporal server start-dev` is simpler than either: one binary,
SQLite in memory unless you pass `--db-filename`, UI included on the gRPC port
plus 1000 (7233 and 8233 by default). Mention Docker Compose only if someone
asks how to run the full cluster locally. On Linux or Windows, install the CLI
from the Temporal docs instead of Homebrew.

`temporalio/samples-go` is one Go module with a folder per sample, each with
a `worker` and a `starter` package. The hello-world sample is the two you need:
`go run ./helloworld/worker` builds and runs the Worker, and
`go run ./helloworld/starter` runs the Client that starts a Workflow and prints
its result. After an edit, stop the Worker with Ctrl+C and `go run` it again;
the build is a second or two. The first `go run` downloads the SDK into the
module cache, so do it before the session, not in front of the room.

On the stores: PostgreSQL and MySQL are the usual choices for self-hosting.
Cassandra is supported but it is not what you want on a laptop.

The original slide also said Workers are polyglot: a Workflow can call an
Activity that another team's Worker runs, on its own Task Queue, whatever that
Worker is written in. Say it only if the room has a mixed stack; this Workshop
is Go end to end.
-->
