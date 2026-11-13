package features

import (
	"context"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// CreateDailyReport is the Schedules slide, wrapped in a function so the
// runtime test can hand it a Client.
func CreateDailyReport(ctx context.Context, c client.Client) (client.ScheduleHandle, error) {
	// slide: 22-more-features.md "heading: Schedules in Go"
	handle, err := c.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID: "daily-report",
		Action: &client.ScheduleWorkflowAction{
			Workflow:  ReportWorkflow,
			TaskQueue: "reports",
		},
		Spec: client.ScheduleSpec{
			CronExpressions: []string{"0 9 * * *"},
		},
		Overlap: enumspb.SCHEDULE_OVERLAP_POLICY_SKIP,
	})
	// end slide
	return handle, err
}
