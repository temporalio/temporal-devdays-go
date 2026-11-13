---
layout: section
---

<ZiggyHero scene="bridge">

# Go Clients

How your application talks to Temporal.

</ZiggyHero>

<!--
The Workers section showed the process that runs the code; this one shows the
code that asks for it to run.
-->

---
layout: default
---

# What is a Client?

Your application's way in to Temporal.

- **Starts Workflows**, and Signals, Queries, Updates or cancels running ones
- **Talks to the Temporal Service**, never to a Worker directly
- **Runs anywhere**: an API handler, a CLI, a cron job, another service
- **Runs none of your Workflow or Activity code.** Workers do

<!--
The picture to leave in the room: the Client and the Worker never meet. The
Client asks the Service to start a Workflow; the Service puts a Task on the
Task Queue; a Worker polling that queue picks it up. Either side can be down
while the other carries on, which is why a Client can start a Workflow even if
no Worker is running yet. It simply waits on the queue.

The Client lives in `go.temporal.io/sdk/client`, the third of the four SDK
packages. It is plain Go and safe for concurrent use, so it goes wherever your
application already runs, and one Client serves every goroutine in it.
-->

---
layout: code-walkthrough
heading: Go Client
notes:
  - "<code>client.Dial</code> connects to the Temporal Service and returns the Client. The same call the Worker makes"
  - "<code>Close</code> on the way out. Create one Client per process and share it"
  - "<code>ExecuteWorkflow</code> starts the Workflow and returns a run as soon as it is recorded"
  - "The business ID is the Workflow <code>ID</code>. While it runs, a second start with it starts nothing new"
  - Must match the Task Queue the Worker polls
  - "<code>transfers.TransferWorkflow</code> is passed for its <strong>name</strong>. The Client never runs it"
  - The input, after the Workflow. Checked against the Workflow's parameters when it starts
  - "<code>run.Get</code> waits for the result and decodes the receipt into a typed variable"
---

::code::

```go {all|2|6|14|15|16|17|17|21-22}
func main() {
	c, err := client.Dial(client.Options{HostPort: "localhost:7233"})
	if err != nil {
		log.Fatalln("unable to create client", err)
	}
	defer c.Close()

	transfer := transfers.Transfer{
		TransferID:  "tx-1001",
		FromAccount: "A-1",
		ToAccount:   "B-2",
		Amount:      250,
	}
	run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "transfer-" + transfer.TransferID,
		TaskQueue: "transfers",
	}, transfers.TransferWorkflow, transfer)
	if err != nil {
		log.Fatalln("unable to start the transfer", err)
	}
	var receipt transfers.Receipt
	if err := run.Get(context.Background(), &receipt); err != nil {
		log.Fatalln("transfer failed", err)
	}
	log.Println(receipt.Status)
}
```

<!--
Code: snippets/clientanatomy/main.go (compiled and vetted, never run: it would
dial localhost:7233). snippets/clientanatomy/main_test.go runs the same steps
against the test's dev server; the Workflow it starts is the anatomy slide's,
run in full by snippets/workflows/workflows_test.go. This is the
`starter/main.go` a learner runs with `go run ./starter`.

- **Build 1 - the connection.** `client.Dial` from `go.temporal.io/sdk/client`
  opens the gRPC connection and returns the Client in one call. It is the same
  call the Worker made: one type serves both. For Temporal Cloud, add your
  namespace's address, `Namespace`, and an API key or TLS in `client.Options`.
- **Build 2 - one per process.** Connections are expensive and a Client is safe
  for concurrent use, so create one at startup, share it, and `Close` it on the
  way out.
- **Build 3 - `ExecuteWorkflow`.** Despite the name, it returns as soon as the
  Service has recorded the start, with a `client.WorkflowRun`. The next section
  is what you do with that run.
- **Build 4 - the Workflow ID.** `transfer-tx-1001`, built from the business
  ID, so a retried request cannot move the money twice, and support can find
  the Workflow by transfer number. A second `ExecuteWorkflow` with the ID of a
  running Workflow starts nothing new; it hands back the run already going.
- **Build 5 - the Task Queue.** The only link between this Client and the
  Worker. A typo here and the Workflow waits forever on a queue nobody polls.
- **Build 6 - the Workflow function.** Passing the function rather than the
  string `"TransferWorkflow"` means a rename is a compile error, not a Workflow
  that never starts. The Client only reads its name; a Client in another
  codebase passes the string instead.
- **Build 7 - the input.** Everything after the Workflow is its arguments, as
  `...any`. The SDK checks them against the function's parameters before it
  sends the start, so a wrong type fails here, in the Client.
- **Build 8 - the result.** `run.Get` blocks until the Workflow completes and
  decodes its result into whatever you point it at, here a `Receipt`. If the
  Workflow failed, `Get` returns the error instead.
-->
