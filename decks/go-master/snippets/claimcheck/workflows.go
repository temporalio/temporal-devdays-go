package claimcheck

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

var options = workflow.ActivityOptions{StartToCloseTimeout: time.Minute}

// slide: 21-claim-check.md "heading: The Workflow holds only the key"
func DocumentWorkflow(ctx workflow.Context, url string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, options)
	var key string
	if err := workflow.ExecuteActivity(ctx, DownloadDocument, url).Get(ctx, &key); err != nil {
		return "", err
	}
	defer func() {
		ctx, _ := workflow.NewDisconnectedContext(ctx)
		_ = workflow.ExecuteActivity(ctx, DeleteDocument, key).Get(ctx, nil)
	}()

	var summary string
	err := workflow.ExecuteActivity(ctx, SummarizeDocument, key).Get(ctx, &summary)
	return summary, err
}

// end slide

// ReportWorkflow's Activity returns 1 MB; External Storage keeps that out of
// the Event History.
func ReportWorkflow(ctx workflow.Context) (int, error) {
	ctx = workflow.WithActivityOptions(ctx, options)
	var report string
	err := workflow.ExecuteActivity(ctx, BuildReport).Get(ctx, &report)
	return len(report), err
}
