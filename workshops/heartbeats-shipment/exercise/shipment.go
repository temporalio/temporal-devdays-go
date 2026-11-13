// ABOUTME: Names, sizes and shapes shared by the Workflow, the Activity, the
// Worker, the starter and the tests.
package shipment

import "time"

// TaskQueue is the Task Queue this Workshop's Worker polls and its starter targets.
const TaskQueue = "shipment-tasks"

// BatchSize is the number of items in one shipment. Big enough that a crash
// lands mid-pack.
const BatchSize = 40

// TimePerItem is the simulated work per item. A whole shipment takes about
// 12 seconds.
const TimePerItem = 300 * time.Millisecond

// ShipmentInput is the shipment to pack.
type ShipmentInput struct {
	ShipmentID string
	ItemCount  int
}

// ShipmentResult is what packing produced. FinishedOnAttempt is the Attempt
// that finished the job: greater than 1 means a crash was survived.
type ShipmentResult struct {
	ShipmentID        string
	Packed            int
	FinishedOnAttempt int32
}
