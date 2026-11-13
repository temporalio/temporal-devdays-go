---
layout: section
---

<ZiggyHero scene="query">

# Go Queries

Reading state out of a running Workflow.

</ZiggyHero>

<!--
Source: PPT slides 80-85.

The mirror of Signals, and a short segment. Signals deliver data in, Queries
read data out.
-->

---
layout: code-walkthrough
heading: Go Queries
notes:
  - State the Query will expose. A plain local variable
  - The Workflow body updates it as it goes
  - "Same pattern as Signals: a name, <code>LoanStatusQuery</code>, shared by caller and Workflow"
  - "Unlike Signals, Queries <strong>return a value</strong>. <code>SetQueryHandler</code> registers a read-only view into Workflow state"
  - The return value can be anything serializable, not just a string. Queries can take arguments too
---

::code::

```go {all|4|12|1|5-7|6}
const LoanStatusQuery = "getLoanStatus"

func LoanWorkflow(ctx workflow.Context, loan Loan) (LoanResult, error) {
	status := "underwriting"
	err := workflow.SetQueryHandler(ctx, LoanStatusQuery, func() (string, error) {
		return status, nil
	})
	if err != nil {
		return LoanResult{}, err
	}

	status = "approved"
	return LoanResult{LoanID: loan.LoanID, Status: status}, nil
}
```

<!--
Source: PPT slides 81-85.
Code: snippets/queries/queries.go, run by snippets/queries/queries_test.go.

The scenario: a loan application moves through underwriting to a decision, and
the customer checks progress in the bank's app. Each refresh of that page is a
Query.

- **Build 1 - the state.** Any variable the handler can close over.
- **Build 2 - the update.** Nothing re-registers the handler. It is a closure
  over `status`, so it reads whatever the variable holds at the moment it is
  asked: a live reference, not a snapshot.
- **Build 3 - the name.** Same shape as a Signal: an exported constant the
  caller uses too. The difference is in the handler's signature, next.
- **Build 4 - the handler.** `SetQueryHandler` takes any function that returns
  a value and an `error`. It must not change state and must not block: no
  Activities, no Timers, no `workflow.Await`. Returning an error fails that
  Query for the caller, not the Workflow. Register it before the Workflow's
  first blocking call, so a Query never arrives before its handler.
- **Build 5 - the types.** A struct works as well as a string. And a Query may
  take arguments: `func(months int) (int, error)` registers a Query that takes
  a term in months and returns the payment. The caller passes the arguments
  after the Query name.

Why it matters: a Client can send a Query at any point during execution without
waiting for completion. The customer's app, a branch officer's screen, a
support tool. Queries
also answer after the Workflow has completed. And a Query costs nothing in
history. It is not an Event, so querying a Workflow a thousand times does not
grow its history by a byte.

Demo line: `temporal workflow query --workflow-id <id> --name getLoanStatus`.
-->
