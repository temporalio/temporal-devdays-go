package tuning

import (
	"time"

	"go.temporal.io/sdk/contrib/sysinfo"
	"go.temporal.io/sdk/worker"
)

// Every Worker option the tuning slides name, with the value each slide gives
// as the Go SDK's default. Compiled, and checked by tuning_test.go: if an
// option is renamed or removed, the build fails here before the slide lies.

// FixedSlots holds the compute and IO settings (fixed slots).
var FixedSlots = worker.Options{
	MaxConcurrentWorkflowTaskExecutionSize:  1_000,
	MaxConcurrentActivityExecutionSize:      1_000,
	MaxConcurrentLocalActivityExecutionSize: 1_000,
	WorkflowTaskPollerBehavior: worker.NewPollerBehaviorSimpleMaximum(worker.PollerBehaviorSimpleMaximumOptions{
		MaximumNumberOfPollers: 2,
	}),
	ActivityTaskPollerBehavior:   worker.NewPollerBehaviorAutoscaling(worker.PollerBehaviorAutoscalingOptions{}),
	TaskQueueActivitiesPerSecond: 50,
	WorkerActivitiesPerSecond:    10,
	// How long one Workflow coroutine may run without yielding.
	DeadlockDetectionTimeout: time.Second,
}

// StickyCacheSize is the memory setting: the process-wide Workflow cache.
const StickyCacheSize = 10_000

// SetCache applies it. Call it before any Worker in the process starts.
func SetCache() { worker.SetStickyWorkflowCacheSize(StickyCacheSize) }

// Composite is the tuner from the auto-tuning notes: resource-based for
// Activities, fixed for everything else.
func Composite() (worker.WorkerTuner, error) {
	fixed := func(n int) worker.SlotSupplier {
		s, err := worker.NewFixedSizeSlotSupplier(n)
		if err != nil {
			panic(err)
		}
		return s
	}
	controllerOpts := worker.DefaultResourceControllerOptions()
	controllerOpts.MemTargetPercent = 0.8
	controllerOpts.CpuTargetPercent = 0.8
	controllerOpts.InfoSupplier = sysinfo.SysInfoProvider()
	activities, err := worker.NewResourceBasedSlotSupplier(
		worker.NewResourceController(controllerOpts),
		worker.DefaultActivityResourceBasedSlotSupplierOptions(),
	)
	if err != nil {
		return nil, err
	}
	return worker.NewCompositeTuner(worker.CompositeTunerOptions{
		WorkflowSlotSupplier:      fixed(1_000),
		ActivitySlotSupplier:      activities,
		LocalActivitySlotSupplier: fixed(1_000),
		NexusSlotSupplier:         fixed(1_000),
	})
}
