---
layout: section
---

<ZiggyHero scene="pack">

# Data Serialization

Everything that crosses the boundary gets converted.

</ZiggyHero>

<!--
Source: PPT slides 59-60, `{{BLOCK:module-serialization-demo}}`. The JSON
traps, custom payload converter and registration slides are new to the Go
Deck.

PORTED: the original slide leaned on serialization attributes on the data
class. Go has struct tags instead, and they are optional, which is itself part
of the point of this segment.
-->

---
layout: code-walkthrough
heading: Data serialization
notes:
  - Plain structs with exported fields. Tags pin the names on the wire
  - Workflow arguments are serialized by the Client and deserialized by the Worker
  - Activity arguments and results make the same trip
  - So does the Workflow result. Every value here makes a round trip through JSON
---

::code::

```go {all|1-5|7|10-11|13-14}
type Claim struct {
	PatientID         string `json:"patientId"`
	DrugNDC           string `json:"drugNdc"`
	BilledAmountCents int64  `json:"billedAmountCents"`
}

func ClaimWorkflow(ctx workflow.Context, claim Claim) (Result, error) {
	ctx = workflow.WithActivityOptions(ctx, opts)
	// claim was JSON on the wire a moment ago
	var result Result
	err := workflow.ExecuteActivity(ctx, ProcessClaim, claim).Get(ctx, &result)

	// and this crosses the wire on the way back out
	return result, err
}
```

<!--
Source: PPT slide 60.
Code: snippets/serialization/workflows.go, run by
snippets/serialization/serialization_test.go.

The default data converter is `encoding/json`. In Go that means the ergonomics
are good and the traps are quiet, so name the traps (next slide).

- **Build 1 - the shape.** A plain struct. Only exported fields are serialized.
  The `json` tags are optional, but they are the habit to teach: without them
  the wire name is the Go field name, so renaming a field in a refactor
  silently empties it for every in-flight Workflow whose Event History still
  has the old name.
- **Build 2 - arguments in.** Serialized by the Client, deserialized by the
  Worker, and they may be different processes on different machines. Use one
  struct argument rather than several positional ones: you can add a field
  later without changing the Workflow's signature.
- **Build 3 - Activity calls.** `ProcessClaim` is an Activity: its argument is
  serialized by the Workflow, stored in the Event History, and handed to
  whichever Worker picks up the Activity Task. Its result comes back the same
  way, decoded into the pointer you pass to `Get`.
- **Build 4 - results out.** Same trip in reverse, to whoever is waiting on the
  Workflow.

The default converter tries, in order: `nil`, `[]byte` (as raw bytes),
protobuf messages (as proto JSON), and then anything `encoding/json` can
handle.

Payload size matters: every argument and result is stored in the Event
History, and Temporal rejects a single payload over 2 MB by default. That is
the problem the Claim-Check segment already solved.
-->

---
layout: code-walkthrough
heading: What JSON does to your values
notes:
  - A time, a number inside an <code>any</code> and an unexported field go in
  - The time.Time comes back intact. RFC 3339 on the wire
  - The number comes back a float64, and one off
  - The unexported field is gone. Zero value
---

::code::

```go {all|2-4|12|13|14}
run, err := c.ExecuteWorkflow(ctx, options, Echo, Visit{
	At:    time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
	Claim: map[string]any{"amount": int64(9_007_199_254_740_993)},
	copay: 25,
})
if err != nil {
	return Visit{}, err
}
var received Visit
err = run.Get(ctx, &received)

fmt.Println(received.At)              // 2026-01-01 09:00:00 +0000 UTC
fmt.Println(received.Claim["amount"]) // 9.007199254740992e+15: float64, one off
fmt.Println(received.copay)           // 0: unexported fields are never sent
```

<!--
New to the Go Deck. `Echo` is a Workflow that returns its argument, and the
tests behind this Deck assert each of these three results.
Code: snippets/serialization/client.go.

- **Build 1 - the input.** Three ordinary values, all legal Go for the `Visit`
  type the Workflow declares.
- **Build 2 - the time.** `time.Time` implements `json.Marshaler`, so it goes
  over as an RFC 3339 string and comes back equal. One caveat: the monotonic
  clock reading is stripped, and a non-UTC location comes back as a fixed
  offset.
- **Build 3 - the number.** Inside an `any` (or a `map[string]any`), JSON has
  no integer type to decode into, so every number comes back a `float64`.
  2^53 + 1 does not fit in a float64's mantissa, so the ID is now wrong by one,
  and the compiler cannot warn you. The same is true inside the Workflow: the
  value it receives is already a float64. Typed fields (`int64`) do not have
  this problem.
- **Build 4 - unexported fields.** `encoding/json` cannot see them, so they are
  never sent. The receiver gets the zero value.

Two more traps that follow from "it is JSON":

- An interface-typed field arrives as a `map[string]any`, not as the concrete
  type you sent. Methods are gone and a type assertion fails.
- A channel or a func cannot be encoded at all, and that fails before anything
  is sent: `ExecuteWorkflow` returns "unable to encode: json: unsupported
  type: chan int".

The rule to give them: pass data, not objects. Concrete structs with exported,
tagged fields; IDs as strings or typed integers, never through `any`.
-->

---
layout: code-walkthrough
heading: A custom payload converter
notes:
  - "<code>Encoding</code> names the format in each Payload's metadata"
  - "<code>ToPayload</code>: value to bytes. Plain JSON on the way out"
  - The encoding travels with the bytes
  - "<code>FromPayload</code>: bytes back to a value. <code>UseNumber</code> keeps every digit"
  - Tried in order. The first converter that accepts the value wins
---

::code::

```go {all|3|5-13|11|16-20|24-28}
type NumberJSONConverter struct{}

func (NumberJSONConverter) Encoding() string { return "json/number" }

func (c NumberJSONConverter) ToPayload(value any) (*commonpb.Payload, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return &commonpb.Payload{
		Metadata: map[string][]byte{converter.MetadataEncoding: []byte(c.Encoding())},
		Data:     data,
	}, nil
}

func (NumberJSONConverter) FromPayload(p *commonpb.Payload, valuePtr any) error {
	dec := json.NewDecoder(bytes.NewReader(p.Data))
	dec.UseNumber()
	return dec.Decode(valuePtr)
}

func (NumberJSONConverter) ToString(p *commonpb.Payload) string { return string(p.Data) }

var DataConverter = converter.NewCompositeDataConverter(
	converter.NewNilPayloadConverter(),
	converter.NewByteSlicePayloadConverter(),
	NumberJSONConverter{},
)
```

<!--
New to the Go Deck. Checked against go.temporal.io/sdk v1.51.0; the imports
are cut for room (`commonpb` is `go.temporal.io/api/common/v1`).
Code: snippets/serialization/converter.go.

When JSON is not enough, the payload converter is the extension point. This one
fixes the trap from the last slide: numbers inside an `any` keep every digit.

- **Build 1 - `Encoding`.** Every Payload carries its encoding in its metadata,
  so the receiving side knows which converter to hand it to.
- **Build 2 - `ToPayload`.** Value to bytes. A converter that cannot handle a
  value returns `nil, nil`, meaning "not mine, try the next converter"; this
  one is last in line and takes everything.
- **Build 3 - the metadata.** That `json/number` label is how the other side
  finds this converter again.
- **Build 4 - `FromPayload`.** Bytes back into the pointer the caller passed.
  `UseNumber` makes the decoder produce a `json.Number`, a string of digits,
  wherever it would have produced a `float64`. Typed fields decode exactly as
  before.
- **Build 5 - composition.** `NewCompositeDataConverter` tries converters in
  order on the way out, and picks by encoding on the way in. This one replaces
  the default JSON converter; the other two are the SDK's own. The default
  also has two protobuf converters; keep them in the list if you pass proto
  messages.

The Go-specific part: this runs *inside Workflow code*, because the Workflow
itself serializes Activity arguments and decodes their results. So it must be
deterministic and quick: no I/O, no clock, no randomness. Anything that needs
I/O, like encryption with a KMS, belongs in a payload codec instead (next
segment).

For most teams, the cheaper fix is the rule from the last slide: keep IDs out
of `any`.
-->

---
layout: code-walkthrough
heading: Register it on both sides
notes:
  - "A value, not a path. <code>DataConverter</code> is a field of the Client options"
  - The Client converts Workflow arguments and reads results
  - "The Worker is built from the Client: <code>worker.New</code> converts everything else with the same one"
---

::code::

```go {all|3|1-4|9}
c, err := client.Dial(client.Options{
	HostPort:      hostPort,
	DataConverter: DataConverter,
})
if err != nil {
	log.Fatalln("Unable to create client", err)
}

w := worker.New(c, taskQueue, worker.Options{})
```

<!--
New to the Go Deck.
Code: snippets/serialization/client.go, run by
snippets/serialization/serialization_test.go.

- **Build 1 - the converter.** An ordinary Go value on `client.Options`. There
  is nothing to bundle: Workflow code runs in the same process as the Worker,
  so the Worker hands it the very same converter.
- **Build 2 - the Client.** Workflow arguments, Signal and Update arguments,
  Query and Workflow results.
- **Build 3 - the Worker.** `worker.New` takes the Client, and with it the
  Client's converter: Activity arguments and results, and the Workflow's own
  inputs and outputs. There is no separate Worker setting to forget.

Register the same converter in every process that dials: the starter, the
Worker, any tool that reads results. A Client without it gets a Payload whose
encoding it does not recognise, and `Get` returns "encoding json/number:
payload encoding is not supported". At runtime, not compile time; the tests
behind this Deck check that too.
-->
