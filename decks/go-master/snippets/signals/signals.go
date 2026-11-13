// Package signals backs 09-signals.md.
package signals

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

type WireTransfer struct {
	WireID string
}

// slide: 09-signals.md "Go Signals"
const ApproveWireSignal = "approveWire"

func WireTransferWorkflow(ctx workflow.Context, wire WireTransfer) (string, error) {
	var approver string

	approvals := workflow.GetSignalChannel(ctx, ApproveWireSignal)
	approvals.Receive(ctx, &approver)

	return fmt.Sprintf("%s approved by %s", wire.WireID, approver), nil
}

// end slide

// AsyncWireTransferWorkflow is not on a slide: it backs the speaker note that
// a Signal can be handled in a coroutine that waits on a durable Timer before
// it changes state, while the Workflow waits for that state.
func AsyncWireTransferWorkflow(ctx workflow.Context, wire WireTransfer) (string, error) {
	approver := ""
	workflow.Go(ctx, func(ctx workflow.Context) {
		var officer string
		workflow.GetSignalChannel(ctx, ApproveWireSignal).Receive(ctx, &officer)
		_ = workflow.Sleep(ctx, time.Second)
		approver = officer
	})
	if err := workflow.Await(ctx, func() bool { return approver != "" }); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s approved by %s", wire.WireID, approver), nil
}
