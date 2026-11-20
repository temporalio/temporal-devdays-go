// ABOUTME: Reads one order's Event History and says what Temporal recorded: each
// Activity and how it ended, which undos ran, and which undo is owed and missing.
// The starter prints this, so the Client tab shows the outcome without the Worker.
package saga

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// undoFor pairs each forward step with the Activity that undoes it. Dispatch is
// the last step, so nothing after it can fail and it has no undo.
var undoFor = map[string]string{
	"ReserveInventory": "ReleaseInventoryIfReserved",
	"ChargePayment":    "RefundPaymentIfCharged",
}

// Step is one Activity in the Event History and how it ended.
type Step struct {
	Activity string // e.g. "ChargePayment"
	Outcome  string // completed, failed, timed out, canceled, or never finished
	Detail   string // the Activity's result, or its failure message
}

// Record is what the Event History says one order did.
type Record struct {
	WorkflowID string
	Status     string // Completed, Failed, Canceled, ... ; Running if not closed
	Steps      []Step
	UndosRan   []string // undos that completed, in the order they completed
	UndosOwed  []string // "Undo, for Step": a step that ran, and no undo completed for it
}

// ReadRecord reads the Event History of one Workflow Execution.
func ReadRecord(ctx context.Context, c client.Client, workflowID, runID string) (Record, error) {
	rec := Record{WorkflowID: workflowID, Status: "Running"}
	scheduled := map[int64]int{} // scheduled event ID -> index in rec.Steps
	iter := c.GetWorkflowHistory(ctx, workflowID, runID, false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return rec, err
		}
		end := func(id int64, outcome, detail string) {
			if i, ok := scheduled[id]; ok {
				rec.Steps[i].Outcome, rec.Steps[i].Detail = outcome, detail
				if outcome == "completed" && isUndo(rec.Steps[i].Activity) {
					rec.UndosRan = append(rec.UndosRan, rec.Steps[i].Activity)
				}
			}
		}
		switch {
		case event.GetActivityTaskScheduledEventAttributes() != nil:
			scheduled[event.GetEventId()] = len(rec.Steps)
			rec.Steps = append(rec.Steps, Step{
				Activity: event.GetActivityTaskScheduledEventAttributes().GetActivityType().GetName(),
				Outcome:  "never finished",
			})
		case event.GetActivityTaskCompletedEventAttributes() != nil:
			a := event.GetActivityTaskCompletedEventAttributes()
			var result string
			_ = converter.GetDefaultDataConverter().FromPayloads(a.GetResult(), &result)
			end(a.GetScheduledEventId(), "completed", result)
		case event.GetActivityTaskFailedEventAttributes() != nil:
			a := event.GetActivityTaskFailedEventAttributes()
			end(a.GetScheduledEventId(), "failed", a.GetFailure().GetMessage())
		case event.GetActivityTaskTimedOutEventAttributes() != nil:
			end(event.GetActivityTaskTimedOutEventAttributes().GetScheduledEventId(), "timed out", "")
		case event.GetActivityTaskCanceledEventAttributes() != nil:
			end(event.GetActivityTaskCanceledEventAttributes().GetScheduledEventId(), "canceled", "")
		case event.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
			rec.Status = "Completed"
		case event.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
			rec.Status = "Failed"
		case event.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
			rec.Status = "Canceled"
		case event.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED:
			rec.Status = "Terminated"
		case event.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
			rec.Status = "TimedOut"
		}
	}

	// An order that did not complete owes an undo for every step that ran: a step
	// that ran may have landed its effect, even when it reported failure.
	if rec.Status != "Completed" && rec.Status != "Running" {
		for _, s := range rec.Steps {
			if undo, ok := undoFor[s.Activity]; ok && !slices.Contains(rec.UndosRan, undo) {
				rec.UndosOwed = append(rec.UndosOwed, undo+", for "+s.Activity)
			}
		}
	}
	return rec, nil
}

// Print writes the record for a learner: one line per Activity, then the undos.
func (r Record) Print(w io.Writer) {
	fmt.Fprintf(w, "Event History of %s:\n", r.WorkflowID)
	for _, s := range r.Steps {
		mark := "  "
		if isUndo(s.Activity) {
			mark = "<-" // an undo
		}
		fmt.Fprintf(w, "  %s %-27s %-9s %s\n", mark, s.Activity, s.Outcome, s.Detail)
	}
	fmt.Fprintf(w, "Workflow: %s\n", r.Status)
	if r.Status == "Completed" {
		return
	}
	fmt.Fprintf(w, "Undos completed, in order: %s\n", listOrNone(r.UndosRan))
	if len(r.UndosOwed) == 0 {
		fmt.Fprintln(w, "Undos owed but never run:  none")
		return
	}
	fmt.Fprintln(w, "Undos owed but never run:")
	for _, owed := range r.UndosOwed {
		fmt.Fprintf(w, "  MISSING %s\n", owed)
	}
}

func isUndo(activity string) bool {
	for _, undo := range undoFor {
		if undo == activity {
			return true
		}
	}
	return false
}

func listOrNone(list []string) string {
	if len(list) == 0 {
		return "none"
	}
	return strings.Join(list, ", ")
}
