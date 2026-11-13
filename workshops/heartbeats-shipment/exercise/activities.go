// ABOUTME: Packs a shipment, one item at a time. Long enough for a crash to land
// mid-job.
package shipment

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
)

// say prints one line in the Worker terminal, stamped with the time so a
// retry's delay is visible.
func say(format string, args ...any) {
	fmt.Printf("%s %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
}

// PackShipment packs input.ItemCount items, one every TimePerItem.
func PackShipment(ctx context.Context, input ShipmentInput) (ShipmentResult, error) {
	attempt := activity.GetInfo(ctx).Attempt

	// TODO 1c start: start from the count the last attempt Heartbeated, or 0 on a first attempt
	packed := 0
	// TODO 1c end

	say("packing %s, attempt %d, starting at item %d", input.ShipmentID, attempt, packed)

	for item := packed; item < input.ItemCount; item++ {
		select {
		case <-ctx.Done():
			// Cancellation reaches a running Activity through its Heartbeat:
			// the SDK cancels ctx. Stop and let the error out.
			say("interrupted at item %d", packed)
			return ShipmentResult{}, ctx.Err()
		case <-time.After(TimePerItem):
		}
		packed = item + 1
		say("packed item %d of %d", packed, input.ItemCount)

		// TODO 1b start: Heartbeat the number of items packed so far
		// TODO 1b end
	}

	return ShipmentResult{ShipmentID: input.ShipmentID, Packed: packed, FinishedOnAttempt: attempt}, nil
}
