package encryption

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/temporalio/temporal-devdays-go/decks/go-master/snippets/internal/testenv"
)

func TestMain(m *testing.M) { testenv.Main(m) }

// runGreet runs Greet through a client and Worker that share the codec, and
// returns the result and the Workflow ID.
func runGreet(t *testing.T, codec Codec) (string, string) {
	t.Helper()
	tq := testenv.ID("encryption")
	c, w := Connect(testenv.HostPort(), tq, codec.AEAD)
	t.Cleanup(c.Close)
	w.RegisterWorkflow(Greet)
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(w.Stop)
	run, err := c.ExecuteWorkflow(testenv.Ctx(t), client.StartWorkflowOptions{ID: testenv.ID("greet"), TaskQueue: tq}, Greet, "Ada Lovelace")
	if err != nil {
		t.Fatal(err)
	}
	var out string
	if err := run.Get(testenv.Ctx(t), &out); err != nil {
		t.Fatal(err)
	}
	return out, run.GetID()
}

func newCodec(t *testing.T) Codec {
	aead, err := NewAEAD()
	if err != nil {
		t.Fatal(err)
	}
	return Codec{AEAD: aead}
}

func TestAClientWithTheCodecReadsPlaintextAndTheWorkflowIsUnchanged(t *testing.T) {
	if got, _ := runGreet(t, newCodec(t)); got != "Hello, Ada Lovelace" {
		t.Fatalf("got %q", got)
	}
}

func TestTheServerStoresOnlyCiphertext(t *testing.T) {
	_, id := runGreet(t, newCodec(t))
	// testenv.Client() has no codec, so it sees what the Server stores.
	iter := testenv.Client().GetWorkflowHistory(testenv.Ctx(t), id, "", false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		e, err := iter.Next()
		if err != nil {
			t.Fatal(err)
		}
		if a := e.GetWorkflowExecutionStartedEventAttributes(); a != nil {
			p := a.Input.Payloads[0]
			if enc := string(p.Metadata[converter.MetadataEncoding]); enc != "binary/encrypted" {
				t.Fatalf("encoding %q", enc)
			}
			if bytes.Contains(p.Data, []byte("Ada")) {
				t.Fatal("the stored input contains the plaintext name")
			}
			return
		}
	}
	t.Fatal("no WorkflowExecutionStarted event")
}

func TestAClientWithoutTheCodecCannotReadTheResult(t *testing.T) {
	_, id := runGreet(t, newCodec(t))
	var got string
	err := testenv.Client().GetWorkflow(testenv.Ctx(t), id, "").Get(testenv.Ctx(t), &got)
	if err == nil || !strings.Contains(err.Error(), "binary/encrypted") {
		t.Fatalf("got %v, %q; want an unknown-encoding error", err, got)
	}
}

func TestDecodePassesUnencryptedPayloadsThrough(t *testing.T) {
	plain, err := converter.GetDefaultDataConverter().ToPayload("hello")
	if err != nil {
		t.Fatal(err)
	}
	out, err := newCodec(t).Decode([]*commonpb.Payload{plain})
	if err != nil {
		t.Fatal(err)
	}
	if out[0] != plain {
		t.Fatal("an unencrypted payload was changed")
	}
}

// Backs the "Who can read what" notes: a Codec Server is the same codec
// behind converter.NewPayloadCodecHTTPHandler.
func TestACodecServerDecodesForTheTemporalUI(t *testing.T) {
	codec := newCodec(t)
	srv := httptest.NewServer(converter.NewPayloadCodecHTTPHandler(codec))
	defer srv.Close()

	plain, err := converter.GetDefaultDataConverter().ToPayload("Ada Lovelace")
	if err != nil {
		t.Fatal(err)
	}
	enc, err := codec.Encode([]*commonpb.Payload{plain})
	if err != nil {
		t.Fatal(err)
	}
	body, err := protojson.Marshal(&commonpb.Payloads{Payloads: enc})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(srv.URL+"/decode", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var decoded struct {
		Payloads []struct {
			Data []byte `json:"data"`
		} `json:"payloads"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("%v: %s", err, raw)
	}
	if len(decoded.Payloads) != 1 || string(decoded.Payloads[0].Data) != `"Ada Lovelace"` {
		t.Fatalf("decoded %s", raw)
	}
}
