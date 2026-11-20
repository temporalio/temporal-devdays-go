// ABOUTME: Runs one order, then prints what Temporal recorded for it: every
// Activity in its Event History, which undos ran and in which order, and any undo
// the order still owes. The Worker terminal shows the same story as [ledger] lines.
//
//	go run ./starter                     a clean order that succeeds
//	go run ./starter -fail-reserve       inventory fails before anything lands
//	go run ./starter -fail-dispatch      courier fails after two steps landed
//	go run ./starter -charge-times-out   the charge lands, then reports failure
package main

import (
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"go.temporal.io/sdk/client"
	tlog "go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"

	saga "github.com/temporalio/temporal-devdays-go/workshops/saga-order-fulfillment"
)

func main() {
	failReserve := flag.Bool("fail-reserve", false, "inventory fails before anything lands")
	failDispatch := flag.Bool("fail-dispatch", false, "courier fails after two steps landed")
	chargeTimesOut := flag.Bool("charge-times-out", false, "the charge lands, then reports failure")
	flag.Parse()

	order := saga.Order{
		OrderID:                "ORD-" + randomID(6),
		SKU:                    "widget",
		Quantity:               1,
		AmountCents:            4999,
		FailReserve:            *failReserve,
		FailDispatch:           *failDispatch,
		ChargeFailsAfterEffect: *chargeTimesOut,
	}

	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = client.DefaultHostPort
	}
	c, err := client.Dial(client.Options{
		HostPort: address,
		// Warnings and errors only, so the order's own lines stand out.
		Logger: tlog.NewStructuredLogger(slog.New(slog.NewTextHandler(os.Stderr,
			&slog.HandlerOptions{Level: slog.LevelWarn}))),
	})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal:", err)
	}
	defer c.Close()

	fmt.Printf("\nFulfilling %s...\n", order.OrderID)

	ctx := context.Background()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        "fulfillment-" + order.OrderID,
		TaskQueue: saga.TaskQueue,
	}, saga.OrderFulfillmentWorkflow, order)
	if err != nil {
		log.Fatalln("Unable to start the order:", err)
	}

	var result saga.OrderResult
	if err := run.Get(ctx, &result); err != nil {
		// A failed order is a business outcome here, not a crash of the starter.
		// The useful message is at the BOTTOM of the error chain: the Workflow
		// failure wraps the Activity failure, which wraps the Activity's own error.
		fmt.Printf("Order failed: %s\n", rootCause(err))
	} else {
		fmt.Printf("Order completed. Tracking number %s\n", result.TrackingNumber)
	}

	// The Workflow has closed. Read what Temporal recorded, not what any process
	// printed: the Event History is the same whichever Worker ran the steps.
	record, err := saga.ReadRecord(ctx, c, run.GetID(), run.GetRunID())
	if err != nil {
		log.Fatalln("Unable to read the Event History:", err)
	}
	fmt.Println()
	record.Print(os.Stdout)
	if scenario := failureScenario(*failReserve, *failDispatch, *chargeTimesOut); scenario != "" {
		fmt.Printf("\n%s fails on purpose, so Failed is the expected outcome. What\n", scenario)
		fmt.Println("matters is the undos: every step that ran needs its undo completed.")
	}
}

// failureScenario names the flag that makes this order fail, or "" for a clean order.
func failureScenario(failReserve, failDispatch, chargeTimesOut bool) string {
	switch {
	case failReserve:
		return "-fail-reserve"
	case failDispatch:
		return "-fail-dispatch"
	case chargeTimesOut:
		return "-charge-times-out"
	}
	return ""
}

// rootCause walks to the deepest error, where the Activity's real message lives.
func rootCause(err error) string {
	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Message()
	}
	return err.Error()
}

// randomID returns n characters of uppercase letters and digits.
func randomID(n int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}
