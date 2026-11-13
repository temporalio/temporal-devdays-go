package workflows

import (
	"fmt"
	"math/rand/v2"
	"time"

	"go.temporal.io/sdk/workflow"
)

type Packing struct {
	Started time.Time
	Box     string
	Label   string
	Lines   []string
}

// slide: 05-workflows.md "Go determinism rules"
func PackingWorkflow(ctx workflow.Context, items map[string]int) (Packing, error) {
	started := workflow.Now(ctx)

	packed := workflow.NewChannel(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		packed.Send(ctx, "box-1")
	})

	box := "none"
	sel := workflow.NewSelector(ctx)
	sel.AddReceive(packed, func(c workflow.ReceiveChannel, _ bool) { c.Receive(ctx, &box) })
	sel.AddFuture(workflow.NewTimer(ctx, time.Hour), func(workflow.Future) {})
	sel.Select(ctx)

	var label string
	err := workflow.SideEffect(ctx, func(workflow.Context) any {
		return fmt.Sprintf("PKG-%06d", rand.IntN(1_000_000))
	}).Get(&label)
	if err != nil {
		return Packing{}, err
	}

	var lines []string
	for _, sku := range workflow.DeterministicKeys(items) {
		lines = append(lines, fmt.Sprintf("%s x%d", sku, items[sku]))
	}
	return Packing{Started: started, Box: box, Label: label, Lines: lines}, nil
}

// end slide
