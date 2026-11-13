package versioning

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// Replay replays an Event History, exported with `temporal workflow show
// --output json`, against one version of the Workflow code. The version is a
// parameter so the tests can replay one history against each of them; the
// slide calls it MyWorkflow.
func Replay(historyFile string, MyWorkflow any) error {
	// slide: 18-versioning.md "Catch it before you deploy"
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(MyWorkflow)

	// Returns an error if the new code would make
	// different decisions from this history
	err := replayer.ReplayWorkflowHistoryFromJSONFile(nil, historyFile)
	// end slide
	return err
}

// NewVersionedWorker returns a Worker that joins the "orders" Worker
// Deployment as build 2.0.0.
func NewVersionedWorker(c client.Client, MyWorkflow any) worker.Worker {
	// slide: 18-versioning.md "Worker Versioning"
	w := worker.New(c, "orders", worker.Options{
		DeploymentOptions: worker.DeploymentOptions{
			UseVersioning: true,
			Version: worker.WorkerDeploymentVersion{
				DeploymentName: "orders",
				BuildID:        "2.0.0",
			},
			DefaultVersioningBehavior: workflow.VersioningBehaviorPinned,
		},
	})
	// end slide
	w.RegisterWorkflow(MyWorkflow)
	w.RegisterActivity(ActivityA)
	w.RegisterActivity(ActivityB)
	w.RegisterActivity(ActivityC)
	return w
}
