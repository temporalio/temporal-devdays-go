// ABOUTME: The order, its result, and the switches that make fulfillment fail on
// purpose. Shared by the Workflow, the Activities, the Worker and the starter.
package saga

// TaskQueue is the Task Queue this Workshop's Worker polls and its starter targets.
const TaskQueue = "fulfillment-tasks"

// Order is one customer order, and the scenario it should fail in, if any.
type Order struct {
	OrderID     string
	SKU         string
	Quantity    int
	AmountCents int

	// FailReserve is a scenario switch. Inventory reservation fails BEFORE
	// recording anything, so the effect never landed. Its undo was registered
	// first, so it runs anyway with nothing to undo, and it has to cope.
	FailReserve bool

	// FailDispatch is a scenario switch. Courier dispatch fails, so the
	// compensations have to run. The ordinary failure: two steps landed, the
	// third did not.
	FailDispatch bool

	// ChargeFailsAfterEffect is a scenario switch. The charge lands and then
	// reports failure anyway, which is what a gateway timeout after a successful
	// capture looks like. The money moved and the Workflow was told it did not,
	// so an undo registered after the step would never have been recorded and
	// the charge would be stranded.
	ChargeFailsAfterEffect bool
}

// OrderResult is what a fulfilled order returns.
type OrderResult struct {
	OrderID        string
	TrackingNumber string
}
