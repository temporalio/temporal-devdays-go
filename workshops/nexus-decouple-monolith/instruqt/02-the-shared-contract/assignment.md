---
slug: the-shared-contract
id: iru89bayjx82
type: challenge
title: 2. The shared contract
teaser: Write the contract both teams depend on. One file, two Operations.
notes:
- type: text
  contents: |-
    # How do two teams agree on a call neither one owns?

    Payments needs a compliance decision. Compliance owns the logic and refuses
    to hand over its database.

    An HTTP client would work until the network blips. Then you are writing
    retries, timeouts, and callback infrastructure by hand.
- type: text
  contents: |-
    # The four Nexus pieces

    Service is the contract. Operation is a method on it. Endpoint is the routing
    rule. Registry is where Endpoints live.

    You write the Service now. The other three come next.
tabs:
- id: jhlkcfobqscw
  title: Exercise
  type: service
  hostname: workshop
  path: /?folder=/root/workshop/nexus-decouple-monolith
  port: 8080
- id: baandcybypyx
  title: Temporal UI
  type: service
  hostname: workshop
  path: /
  port: 8233
- id: yqcfcixu3ot1
  title: Terminal
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
- id: bcozm7jx2ls0
  title: Payments Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
- id: xw6xranfoi4e
  title: Compliance Worker
  type: terminal
  hostname: workshop
  workdir: /root/workshop/nexus-decouple-monolith/exercise
difficulty: basic
timelimit: 900
enhanced_loading: null
---

# One contract, two teams

A Nexus Service is a contract both teams build against. Payments calls the Operations it
names. Compliance implements a handler for each one. Neither team sees the other's code.

It lives in the `shared` package because neither team owns it alone. It holds three
things: the Service name, and one reference per Operation that pairs a name with an input
type and an output type.

# Write it

Click the [button label="Exercise" background="#444CE7"](tab-0) tab, open
`exercise/shared/service.go`, and follow TODO 1.

<!-- snippet: exercise/shared/service.go TODO 1 -->
Lines 14-37 hold two placeholders, with no names and no types:

```go
// TODO 1 start
// Declare the two Operations this Service offers. Each one is a name paired
// with an input type and an output type, and carries no implementation: only
// what both teams agree on.
//
//	CheckComplianceOperation   "checkCompliance"   ComplianceRequest -> ComplianceResult
//	SubmitReviewOperation      "submitReview"      ReviewRequest     -> ComplianceResult
//
// Write each one as nexus.NewOperationReference[Input, Output]("name").
//
// The two below are placeholders, with no name and no types, so the code
// compiles. Replace them. The Compliance handlers name their Operations after
// these, and the Payments Workflow calls them, so everything downstream
// builds on what you write here.
//
// Stuck? "Ask AI" on https://docs.temporal.io:
//
//	"how do I define a Nexus Service contract in Go?"
var (
	CheckComplianceOperation = nexus.NewOperationReference[any, any]("")
	SubmitReviewOperation    = nexus.NewOperationReference[any, any]("")
)

// TODO 1 end
```

The editor saves as you type. There is no save button.

# See who depends on it

Open `exercise/compliance/handler.go`. The Compliance handlers are already written to
name themselves after your references, so they answer to whatever you declare:

```go,nocopy
var CheckComplianceOperation = temporalnexus.NewWorkflowRunOperation(
	shared.CheckComplianceOperation.Name(),
```

In challenge 4 the Payments Workflow calls `shared.CheckComplianceOperation` the same way.
One declaration, two teams building on it. That is the contract.

The compiler checks every use of the names. What it cannot check is that they are the
names and types both teams agreed on, so the Check does. Click the
[button label="Terminal" background="#444CE7"](tab-2) tab and run it yourself:

```bash,run
go test ./tests/ -run 'TestChallenge2_'
```

A placeholder still in place fails with the file, the TODO, and the line to write.

Click **Check** when the two Operations are declared.

# What you know now

- `nexus.NewOperationReference[Input, Output]("name")` declares an Operation: a name and
  two types, no implementation.
- The contract is shared. The implementation is not.
- The handler and the caller both build on the same references, so a name that drifts
  on one side drifts on both.

---

Tell us what worked in the **Feedback** tab. It takes a few seconds.
