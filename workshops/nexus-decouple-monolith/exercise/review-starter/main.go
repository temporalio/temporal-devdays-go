// [GIVEN] Approves TXN-B through Nexus: go run ./review-starter
//
// Rather than sending the Update to the Compliance Workflow directly, this
// starts a ReviewCallerWorkflow that calls the submitReview Nexus Operation.
// The Compliance team's sync handler forwards the decision to the running
// ComplianceWorkflow as an Update.
//
// Flip Approved to false below to see the denial path.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/payments"
	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

func main() {
	fmt.Println()
	fmt.Println("  REVIEW STARTER: submitting a review for TXN-B via Nexus")
	fmt.Println()

	c, err := client.Dial(client.Options{HostPort: hostPort(), Namespace: shared.PaymentsNamespace})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	request := shared.ReviewRequest{
		TransactionID: "TXN-B",
		Approved:      true,
		Explanation:   "Approved after manual review",
	}
	fmt.Println("  Approved:   ", request.Approved)
	fmt.Println("  Explanation:", request.Explanation)
	fmt.Println()

	run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "review-" + request.TransactionID,
		TaskQueue: shared.PaymentsTaskQueue,
	}, payments.ReviewCallerWorkflow, request)
	if err != nil {
		log.Fatalln("Unable to start the review:", err)
	}
	var result shared.ComplianceResult
	if err := run.Get(context.Background(), &result); err != nil {
		log.Fatalln("The review failed:", err)
	}

	decision := "DENIED"
	if result.Approved {
		decision = "APPROVED"
	}
	fmt.Println("  Review result:", decision)
	fmt.Println("  Risk level:   ", result.RiskLevel)
	fmt.Println("  Explanation:  ", result.Explanation)
	fmt.Println()
	fmt.Println("  TXN-B review submitted. The payment Workflow will now complete.")
}

func hostPort() string {
	if address := os.Getenv("TEMPORAL_ADDRESS"); address != "" {
		return address
	}
	return client.DefaultHostPort
}
