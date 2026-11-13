// Package heartbeat backs 16-heartbeating.md.
package heartbeat

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.temporal.io/sdk/activity"
)

type ShipmentInput struct {
	ShipmentID string
	ItemCount  int
}

var (
	mu sync.Mutex
	// Every item packed, in order, across all attempts. The first attempt dies
	// on item 2, so a correct resume packs item 2 twice and never repeats 0 or 1.
	packed []int
	// One entry per time ProcessItems cleaned up after a cancellation.
	locksReleased []bool
)

func packItem(ctx context.Context, shipmentID string, index int) error {
	mu.Lock()
	packed = append(packed, index)
	mu.Unlock()
	if index == 2 && activity.GetInfo(ctx).Attempt == 1 {
		return errors.New("the Worker died mid-shipment")
	}
	return nil
}

// slide: 16-heartbeating.md "Heartbeats in Go"
func PackShipment(ctx context.Context, input ShipmentInput) (int, error) {
	info := activity.GetInfo(ctx)
	activity.GetLogger(ctx).Info("packing", "attempt", info.Attempt)
	start := 0
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &start); err != nil {
			return 0, err
		}
	}

	for index := start; index < input.ItemCount; index++ {
		if err := packItem(ctx, input.ShipmentID, index); err != nil {
			return 0, err
		}
		activity.RecordHeartbeat(ctx, index+1)
	}

	return input.ItemCount, nil
}

// end slide

type ProcessInput struct {
	Items []string
}

func processItem(ctx context.Context, item string) error {
	select {
	case <-time.After(200 * time.Millisecond):
	case <-ctx.Done():
	}
	return nil
}

func releaseLocks(input ProcessInput) {
	mu.Lock()
	defer mu.Unlock()
	locksReleased = append(locksReleased, true)
}

// slide: 16-heartbeating.md "Heartbeating and cancellation"
func ProcessItems(ctx context.Context, input ProcessInput) error {
	for index, item := range input.Items {
		if err := processItem(ctx, item); err != nil {
			return err
		}
		activity.RecordHeartbeat(ctx, index+1)

		if err := ctx.Err(); err != nil {
			releaseLocks(input)
			return err
		}
	}
	return nil
}

// end slide
