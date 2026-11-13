// ABOUTME: Names, sizes and shapes shared by the Workflow, the Activities, the
// Worker, the starter and the tests.
package moneytransfer

// TaskQueue is the Task Queue this Workshop's Worker polls and its starter sends
// work to. One constant, so the two can never disagree: a typo here and the
// transfer waits forever.
const TaskQueue = "money-transfer"

// LargeDeposit is the size, in dollars, at which a deposit writes an audit note
// first.
const LargeDeposit = 1000

// PaymentDetails is the transfer's input, shared by the Workflow, the
// Activities and the starter.
type PaymentDetails struct {
	SourceAccount string `json:"sourceAccount"`
	TargetAccount string `json:"targetAccount"`
	Amount        int    `json:"amount"`
	ReferenceID   string `json:"referenceId"`
}
