package claimcheck

import (
	"fmt"
	"testing"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

func register(r worker.Registry) {
	r.RegisterWorkflow(DocumentWorkflow)
	r.RegisterWorkflow(ReportWorkflow)
	r.RegisterActivity(DownloadDocument)
	r.RegisterActivity(SummarizeDocument)
	r.RegisterActivity(DeleteDocument)
	r.RegisterActivity(BuildReport)
}

// largestEvent is the size in bytes of the biggest event in a Workflow's
// Event History, payloads included.
func largestEvent(t *testing.T, id string) int {
	t.Helper()
	largest := 0
	iter := testenv.Client().GetWorkflowHistory(testenv.Ctx(t), id, "", false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		largest = max(largest, proto.Size(e))
	}
	return largest
}

func runDocument(t *testing.T, url string) string {
	t.Helper()
	tq := testenv.Worker(t, register)
	id := testenv.ID("document")
	run, err := testenv.Client().ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: id, TaskQueue: tq}, DocumentWorkflow, url)
	if err != nil {
		t.Fatal(err)
	}
	var summary string
	if err := run.Get(testenv.Ctx(t), &summary); err != nil {
		t.Fatal(err)
	}
	if want := fmt.Sprintf("%d bytes", 3*1024*1024); summary != want {
		t.Fatalf("summary %q, want %q", summary, want)
	}
	return id
}

func TestHandRolledClaimCheckMoves3MBWithOnlyAKeyInTheEventHistory(t *testing.T) {
	id := runDocument(t, "https://example.com/report.pdf")
	if n := largestEvent(t, id); n >= 2048 {
		t.Fatalf("largest event is %d bytes; the document leaked into the Event History", n)
	}
}

func TestHandRolledClaimCheckDeletesTheObjectOnTheWayOut(t *testing.T) {
	runDocument(t, "https://example.com/a.pdf")
	if n := s3Client.Len(); n != 0 {
		t.Fatalf("%d objects left in the bucket", n)
	}
}

func TestExternalStorageOffloadsA1MBActivityResult(t *testing.T) {
	driver := NewMemoryDriver()
	tq := testenv.ID("extstore")
	c, w := Connect(testenv.HostPort(), tq, driver)
	defer c.Close()
	register(w)
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: testenv.ID("report"), TaskQueue: tq}, ReportWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	var length int
	if err := run.Get(testenv.Ctx(t), &length); err != nil {
		t.Fatal(err)
	}
	if length != 1024*1024 {
		t.Fatalf("the Workflow saw %d bytes", length)
	}
	if driver.Len() != 1 {
		t.Fatalf("the driver stored %d payloads; want the one oversized result", driver.Len())
	}
	if n := largestEvent(t, run.GetID()); n >= 2048 {
		t.Fatalf("largest event is %d bytes; the report was not offloaded", n)
	}
}
