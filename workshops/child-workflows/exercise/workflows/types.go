// ABOUTME: Names and shapes shared by the Workflows, the Worker, the starter and the tests.
package workflows

// TaskQueue is the Task Queue this Workshop's Worker polls and its starter targets.
const TaskQueue = "child-workflows"

// Order is one customer order: an ID and the items to ship or pack.
type Order struct {
	OrderID string   `json:"orderId"`
	Items   []string `json:"items"`
}

// ChainRequest is the input of every level of a ChainWorkflow.
type ChainRequest struct {
	// Depth is how many levels of children are still to start below this one.
	Depth int `json:"depth"`
	// FailAtBottom makes the Workflow at the bottom of the chain fail instead of returning.
	FailAtBottom bool `json:"failAtBottom"`
}
