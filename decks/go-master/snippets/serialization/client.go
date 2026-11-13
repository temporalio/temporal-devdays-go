package serialization

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// JSONTraps starts Echo with a Visit and returns what came back.
func JSONTraps(ctx context.Context, c client.Client, options client.StartWorkflowOptions) (Visit, error) {
	// slide: 19-serialization.md "heading: What JSON does to your values"
	run, err := c.ExecuteWorkflow(ctx, options, Echo, Visit{
		At:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
		Claim: map[string]any{"amount": int64(9_007_199_254_740_993)},
		copay: 25,
	})
	if err != nil {
		return Visit{}, err
	}
	var received Visit
	err = run.Get(ctx, &received)

	fmt.Println(received.At)              // 2026-01-01 09:00:00 +0000 UTC
	fmt.Println(received.Claim["amount"]) // 9.007199254740992e+15: float64, one off
	fmt.Println(received.copay)           // 0: unexported fields are never sent
	// end slide
	return received, err
}

// Connect dials a client with the custom converter and builds a Worker from
// it.
func Connect(hostPort, taskQueue string) (client.Client, worker.Worker) {
	// slide: 19-serialization.md "heading: Register it on both sides"
	c, err := client.Dial(client.Options{
		HostPort:      hostPort,
		DataConverter: DataConverter,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}

	w := worker.New(c, taskQueue, worker.Options{})
	// end slide
	return c, w
}
