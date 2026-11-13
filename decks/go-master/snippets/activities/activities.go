// Package activities backs 06-activities.md. The Activity calls a real HTTP
// API; the tests point OrdersAPI at a local server.
package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/workflow"
)

// OrdersAPI is the orders service the Activity posts to.
var OrdersAPI = "https://orders.example.com"

type Order struct {
	OrderID  string
	SKU      string
	Quantity int
}

type OrderResult struct {
	OrderID string
	Status  string
}

// slide: 06-activities.md "Go Activity anatomy"
// activities.go
func SubmitOrder(ctx context.Context, order Order) (OrderResult, error) {
	body, err := json.Marshal(order)
	if err != nil {
		return OrderResult{}, err
	}
	resp, err := http.Post(OrdersAPI+"/submit", "application/json", bytes.NewReader(body))
	if err != nil {
		return OrderResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return OrderResult{}, fmt.Errorf("orders API returned %d", resp.StatusCode)
	}
	var result OrderResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// end slide

// OrderWorkflow is not on a slide: the Workflow that schedules SubmitOrder,
// for the tests here and for the Worker on 07-workers.md.
func OrderWorkflow(ctx workflow.Context, order Order) (OrderResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})
	var result OrderResult
	err := workflow.ExecuteActivity(ctx, SubmitOrder, order).Get(ctx, &result)
	return result, err
}
