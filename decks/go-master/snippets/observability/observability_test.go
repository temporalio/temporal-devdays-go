package observability

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

// entry is one captured log line: its message and its key-value pairs.
type entry struct {
	msg    string
	fields map[string]any
}

// capture is a log.Logger that keeps every Info line.
type capture struct {
	mu      sync.Mutex
	entries []entry
}

func (c *capture) add(msg string, keyvals []any) {
	f := map[string]any{}
	for i := 0; i+1 < len(keyvals); i += 2 {
		f[fmt.Sprint(keyvals[i])] = keyvals[i+1]
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, entry{msg, f})
}

func (c *capture) Debug(string, ...any)             {}
func (c *capture) Info(msg string, keyvals ...any)  { c.add(msg, keyvals) }
func (c *capture) Warn(msg string, keyvals ...any)  { c.add(msg, keyvals) }
func (c *capture) Error(msg string, keyvals ...any) { c.add(msg, keyvals) }

func (c *capture) byMessage(msg string) (out []entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.entries {
		if e.msg == msg {
			out = append(out, e)
		}
	}
	return out
}

func expectFields(t *testing.T, e entry, want map[string]any) {
	t.Helper()
	for k, v := range want {
		if fmt.Sprint(e.fields[k]) != fmt.Sprint(v) {
			t.Errorf("%q: %s = %v, want %v (all fields: %v)", e.msg, k, e.fields[k], v, e.fields)
		}
	}
}

// The logging slides: Workflow and Activity log lines reach the client's
// Logger with the execution's identity attached, and Workflow lines are not
// repeated when the Workflow replays.
func TestLogsCarryTheExecutionsIdentityOnceEachDespiteReplay(t *testing.T) {
	logs := &capture{}
	c, err := client.Dial(client.Options{HostPort: testenv.HostPort(), Logger: logs})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	tq := testenv.ID("observability")
	w := worker.New(c, tq, worker.Options{})
	w.RegisterWorkflow(OrderWorkflow)
	w.RegisterActivity(ChargeCard)
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	id := testenv.ID("order")
	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: id, TaskQueue: tq}, OrderWorkflow, "A-1")
	if err != nil {
		t.Fatal(err)
	}
	var receipt string
	if err := run.Get(testenv.Ctx(t), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt != "receipt-A-1" {
		t.Fatalf("receipt %q", receipt)
	}

	for _, msg := range []string{"Order received", "Order charged", "Charging card", "Card charged"} {
		if n := len(logs.byMessage(msg)); n != 1 {
			t.Errorf("%q logged %d times, want 1", msg, n)
		}
	}
	if got := logs.byMessage("Order received"); len(got) == 1 {
		expectFields(t, got[0], map[string]any{
			"OrderID": "A-1", "WorkflowID": id, "WorkflowType": "OrderWorkflow", "TaskQueue": tq,
		})
	}
	if got := logs.byMessage("Charging card"); len(got) == 1 {
		expectFields(t, got[0], map[string]any{
			"OrderID": "A-1", "WorkflowID": id, "ActivityType": "ChargeCard", "Attempt": 1,
		})
	}

	// Replay the whole Event History, as a Worker does after a cache eviction
	// or on a different Worker: the Workflow code runs again from the top, and
	// logs nothing.
	history := &historypb.History{}
	iter := c.GetWorkflowHistory(testenv.Ctx(t), id, run.GetRunID(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		history.Events = append(history.Events, e)
	}
	replayLogs := &capture{}
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(OrderWorkflow)
	if err := replayer.ReplayWorkflowHistory(replayLogs, history); err != nil {
		t.Fatal(err)
	}
	for _, msg := range []string{"Order received", "Order charged"} {
		if n := len(replayLogs.byMessage(msg)); n != 0 {
			t.Errorf("replay logged %q %d times, want 0", msg, n)
		}
	}
}

// The SDK-metrics slide: a client dialled with the Prometheus-backed
// MetricsHandler serves /metrics on port 9000, and a Worker built from it
// reports the `temporal_`-prefixed names the SDK metric slides use. Timers get
// the `_seconds` suffix Prometheus expects.
func TestAWorkerOnTheClientExportsTheSDKMetrics(t *testing.T) {
	c, err := Dial(testenv.HostPort())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	tq := testenv.ID("observability")
	w := worker.New(c, tq, worker.Options{})
	w.RegisterWorkflow(OrderWorkflow)
	w.RegisterActivity(ChargeCard)
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: testenv.ID("order"), TaskQueue: tq}, OrderWorkflow, "A-2")
	if err != nil {
		t.Fatal(err)
	}
	if err := run.Get(testenv.Ctx(t), nil); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"temporal_worker_task_slots_available",
		"temporal_num_pollers",
		"temporal_sticky_cache_size",
		"temporal_activity_execution_latency_seconds",
		"temporal_workflow_task_execution_latency_seconds",
		"temporal_workflow_task_replay_latency_seconds",
		"temporal_workflow_task_schedule_to_start_latency_seconds",
		"temporal_activity_schedule_to_start_latency_seconds",
		"temporal_long_request_latency_seconds",
	}
	slots := []*regexp.Regexp{
		regexp.MustCompile(`temporal_worker_task_slots_available\{[^}]*worker_type="WorkflowWorker"`),
		regexp.MustCompile(`temporal_worker_task_slots_available\{[^}]*worker_type="ActivityWorker"`),
	}
	// The scope reports once a second, so give it a few.
	var body string
	var missing []string
	for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); time.Sleep(500 * time.Millisecond) {
		resp, err := http.Get("http://127.0.0.1:9000/metrics")
		if err != nil {
			continue
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		body = string(b)
		missing = missing[:0]
		for _, name := range want {
			if !strings.Contains(body, name) {
				missing = append(missing, name)
			}
		}
		for _, re := range slots {
			if !re.MatchString(body) {
				missing = append(missing, re.String())
			}
		}
		if len(missing) == 0 {
			return
		}
	}
	t.Fatalf("/metrics is missing %v", missing)
}
