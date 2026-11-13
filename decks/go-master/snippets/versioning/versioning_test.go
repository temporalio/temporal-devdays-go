package versioning_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/versioning"
	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/versioning/before"
	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/versioning/deprecated"
	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/versioning/final"
	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/versioning/patched"
)

func TestMain(m *testing.M) { testenv.Main(m) }

// run executes one version of MyWorkflow to completion and returns its ID.
func run(t *testing.T, wf any) string {
	t.Helper()
	tq := testenv.Worker(t, func(r worker.Registry) {
		r.RegisterWorkflow(wf)
		r.RegisterActivity(versioning.ActivityA)
		r.RegisterActivity(versioning.ActivityB)
		r.RegisterActivity(versioning.ActivityC)
	})
	id := testenv.ID("versioning")
	wr, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: id, TaskQueue: tq}, "MyWorkflow")
	if err != nil {
		t.Fatal(err)
	}
	if err := wr.Get(testenv.Ctx(t), nil); err != nil {
		t.Fatal(err)
	}
	return id
}

func events(t *testing.T, id string) []*historypb.HistoryEvent {
	t.Helper()
	var out []*historypb.HistoryEvent
	iter := testenv.Client().GetWorkflowHistory(testenv.Ctx(t), id, "", false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, e)
	}
	return out
}

func activitiesRun(t *testing.T, id string) []string {
	var names []string
	for _, e := range events(t, id) {
		if a := e.GetActivityTaskScheduledEventAttributes(); a != nil {
			names = append(names, a.ActivityType.Name)
		}
	}
	return names
}

// replay exports a run's Event History with the CLI, as the slide's notes
// say to, and replays it against one version of MyWorkflow.
func replay(t *testing.T, id string, wf any) error {
	t.Helper()
	file := filepath.Join(t.TempDir(), "history.json")
	out, err := exec.Command("temporal", "workflow", "show", "--address", testenv.HostPort(),
		"--workflow-id", id, "--output", "json").Output()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, out, 0o600); err != nil {
		t.Fatal(err)
	}
	return versioning.Replay(file, wf)
}

func TestPatching(t *testing.T) {
	beforeRun := run(t, before.MyWorkflow)   // started on the original code: A, B
	patchedRun := run(t, patched.MyWorkflow) // started after the patch: A, marker, C

	t.Run("the original code runs A, then B", func(t *testing.T) {
		if got := activitiesRun(t, beforeRun); !slices.Equal(got, []string{"ActivityA", "ActivityB"}) {
			t.Fatalf("got %v", got)
		}
	})
	t.Run("swapping B for C without GetVersion breaks replay of a run that ran B", func(t *testing.T) {
		if err := replay(t, beforeRun, final.MyWorkflow); err == nil {
			t.Fatal("replay succeeded; want a non-determinism error")
		}
	})
	t.Run("the versioned code replays an old run down the old branch", func(t *testing.T) {
		if err := replay(t, beforeRun, patched.MyWorkflow); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("a new run takes the new branch and records a marker", func(t *testing.T) {
		if got := activitiesRun(t, patchedRun); !slices.Equal(got, []string{"ActivityA", "ActivityC"}) {
			t.Fatalf("got %v", got)
		}
		var markers []string
		for _, e := range events(t, patchedRun) {
			if a := e.GetMarkerRecordedEventAttributes(); a != nil {
				markers = append(markers, a.MarkerName)
			}
		}
		if !slices.Equal(markers, []string{"Version"}) {
			t.Fatalf("markers %v", markers)
		}
	})
	t.Run("raising the minimum to 1 replays runs that have the marker", func(t *testing.T) {
		if err := replay(t, patchedRun, deprecated.MyWorkflow); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("raising the minimum to 1 refuses a run still on the old branch, by name", func(t *testing.T) {
		err := replay(t, beforeRun, deprecated.MyWorkflow)
		if err == nil || !strings.Contains(err.Error(), "removed support of version") {
			t.Fatalf("got %v; want the unsupported-version error", err)
		}
	})
	t.Run("once only marked runs remain, GetVersion can be deleted", func(t *testing.T) {
		// The replayer skips a Version marker no GetVersion call consumes.
		if err := replay(t, patchedRun, final.MyWorkflow); err != nil {
			t.Fatal(err)
		}
		deprecatedRun := run(t, deprecated.MyWorkflow)
		if err := replay(t, deprecatedRun, final.MyWorkflow); err != nil {
			t.Fatal(err)
		}
	})
}

func TestWorkerVersioningPinsANewRun(t *testing.T) {
	c := testenv.Client()
	w := versioning.NewVersionedWorker(c, patched.MyWorkflow)
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// The version exists once its Worker has polled; until then this fails.
	ctx := testenv.Ctx(t)
	handle := c.WorkerDeploymentClient().GetHandle("orders")
	var err error
	for range 100 {
		if _, err = handle.SetCurrentVersion(ctx, client.WorkerDeploymentSetCurrentVersionOptions{BuildID: "2.0.0"}); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}

	wr, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: testenv.ID("pinned"), TaskQueue: "orders"}, "MyWorkflow")
	if err != nil {
		t.Fatal(err)
	}
	if err := wr.Get(ctx, nil); err != nil {
		t.Fatal(err)
	}
	desc, err := c.DescribeWorkflowExecution(ctx, wr.GetID(), wr.GetRunID())
	if err != nil {
		t.Fatal(err)
	}
	info := desc.WorkflowExecutionInfo.VersioningInfo
	if info.GetBehavior() != enumspb.VERSIONING_BEHAVIOR_PINNED {
		t.Fatalf("behavior %v", info.GetBehavior())
	}
	if got := info.GetDeploymentVersion().GetBuildId(); got != "2.0.0" {
		t.Fatalf("build ID %q", got)
	}
}
