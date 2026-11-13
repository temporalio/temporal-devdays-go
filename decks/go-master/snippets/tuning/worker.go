package tuning

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/sysinfo"
	"go.temporal.io/sdk/worker"
)

// NewTunedWorker builds the Worker on the auto-tuning slide.
func NewTunedWorker(c client.Client) (worker.Worker, error) {
	// slide: 26-worker-tuning.md "Execution slot auto-tuning"
	tuner, err := worker.NewResourceBasedTuner(worker.ResourceBasedTunerOptions{
		TargetMem:    0.8,
		TargetCpu:    0.8,
		InfoSupplier: sysinfo.SysInfoProvider(),
	})
	if err != nil {
		return nil, err
	}
	w := worker.New(c, "order-task-queue", worker.Options{
		Tuner: tuner,
	})
	w.RegisterWorkflow(OrderWorkflow)
	w.RegisterActivity(SubmitOrder)
	// end slide
	return w, nil
}
