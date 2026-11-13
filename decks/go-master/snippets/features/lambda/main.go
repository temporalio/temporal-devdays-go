// Command lambda is the Serverless Workers slide in 22-more-features.md.
// go vet compiles it against go.temporal.io/sdk/contrib/aws/lambdaworker; it
// never runs here, because it only makes sense inside AWS Lambda, invoked by
// Temporal.
package main

import (
	"go.temporal.io/sdk/contrib/aws/lambdaworker"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/features"
)

var (
	OrderWorkflow = features.OrderWorkflow
	ChargePayment = features.ChargePayment
)

// slide: 22-more-features.md "heading: Serverless Workers in Go"
func main() {
	lambdaworker.RunWorker(worker.WorkerDeploymentVersion{
		DeploymentName: "orders",
		BuildID:        "build-1",
	}, func(opts *lambdaworker.Options) error {
		opts.TaskQueue = "orders"
		opts.RegisterWorkflow(OrderWorkflow)
		opts.RegisterActivity(ChargePayment)
		return nil
	})
}

// end slide
