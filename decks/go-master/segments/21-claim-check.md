---
layout: section
---

<ZiggyHero scene="claim">

# The Claim-Check Pattern

Large payloads do not belong in the Event History.

</ZiggyHero>

<!--
Source: PPT slides 65-67, `{{SLIDE:claim-check-upload-activity}}` and
`{{SLIDE:claim-check-fetch-activity}}`. The Workflow and External Storage
slides are new to the Go Deck.

Set it up before the code: every Activity argument and result is stored in the
Event History. Temporal rejects a single payload over 2 MB by default (fixed at
2 MB on Temporal Cloud), and terminates a Workflow whose Event History passes
50 MB. Put a 40 MB PDF through an Activity result and it never reaches the
Workflow: the completion is rejected and the Activity fails; put a 1 MB one
through twenty Activities and you are most of the way to the history limit.

The fix is the oldest one in messaging: store the thing, pass the receipt.
-->

---
layout: code-walkthrough
heading: Store the bytes, return the key
notes:
  - The key is the claim check. A small string that stands in for a large object
  - Store the bytes anywhere. S3 here, but a bucket, a blob store or a database will do
  - Return the key. This is the only thing that enters the Event History
---

::code::

```go {all|2|3-7|8}
func uploadToStorage(ctx context.Context, data []byte) (string, error) {
	key := uuid.NewString()
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	return key, err
}
```

<!--
Source: PPT slide 66, ported to the AWS SDK for Go v2's call shapes. The
tests behind this Deck run it against an in-memory stand-in for the S3 client.
Code: snippets/claimcheck/activities.go, run by
snippets/claimcheck/claimcheck_test.go.

Keep it simple: it is a thin wrapper around your blob storage client and it
should stay that way.

The trap in the original framing: this is *not* an Activity the Workflow calls
with the bytes. If the Workflow passed `data` as an Activity argument, the bytes
would already be in the Event History. That is why it is an unexported helper,
not an Activity: call it from inside the Activity that *produces* the bytes
(next-but-one slide), or from the Client before it starts the Workflow.

`uuid.NewString()` is fine here because this runs in an Activity, outside
Workflow code. Inside a Workflow it would be a different random value on every
replay; there you would wrap it in `workflow.SideEffect`. Worth pointing at:
it is the determinism rule from the design-patterns segment showing up in real
code.

The guarantee to name: the Activity around this gets Temporal's full retry
treatment, so a transient S3 error does not lose the data. The cost of a retry
is a fresh key and, if the first upload succeeded before the Worker crashed,
an orphaned object. A bucket lifecycle rule cleans those up.
-->

---
layout: code-walkthrough
heading: Fetch by key
notes:
  - Takes the claim check back
  - Fetches the object by its key
  - Reads the whole body back into a byte slice
---

::code::

```go {all|1|2-5|9-10}
func fetchFromStorage(ctx context.Context, key string) ([]byte, error) {
	out, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}
```

<!--
Source: PPT slide 67. The original reassembled the stream chunk by chunk;
`io.ReadAll` does the same.
Code: snippets/claimcheck/activities.go.

The mirror of the upload. Nothing clever.

If the object is too large to hold in memory, stream it instead: `out.Body` is
an `io.ReadCloser`, and the Activity can process it chunk by chunk,
heartbeating as it goes.
-->

---
layout: code-walkthrough
heading: The Workflow holds only the key
notes:
  - The Activity that produces the bytes stores them, and returns only the key
  - The Activity that needs the bytes fetches them by key
  - You own the object's lifecycle. Delete it on the way out
---

::code::

```go {all|4|13|7-10}
func DocumentWorkflow(ctx workflow.Context, url string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, options)
	var key string
	if err := workflow.ExecuteActivity(ctx, DownloadDocument, url).Get(ctx, &key); err != nil {
		return "", err
	}
	defer func() {
		ctx, _ := workflow.NewDisconnectedContext(ctx)
		_ = workflow.ExecuteActivity(ctx, DeleteDocument, key).Get(ctx, nil)
	}()

	var summary string
	err := workflow.ExecuteActivity(ctx, SummarizeDocument, key).Get(ctx, &summary)
	return summary, err
}
```

<!--
New to the Go Deck.
Code: snippets/claimcheck/workflows.go, run by
snippets/claimcheck/claimcheck_test.go.

- **Build 1 - produce.** `DownloadDocument` is an Activity: it downloads the
  file and calls `uploadToStorage` itself, so the bytes never leave the
  Activity. The Workflow gets a 36-character string.
- **Build 2 - consume.** `SummarizeDocument` calls `fetchFromStorage` with the
  key. Several downstream Activities can fetch with the same key: the bytes
  never pass through the Workflow, never enter the Event History, and are never
  duplicated. The Workflow holds a string and orchestrates gigabytes. The tests
  behind this Deck move a 3 MB document through this Workflow (over the 2 MB
  payload limit) and check that no event in its Event History is over 2 KB.
- **Build 3 - clean up.** Temporal will not delete the object. A `defer`
  registered as soon as the key exists runs on the failure path too. The
  disconnected context is what makes it run when the Workflow is cancelled: a
  cancelled `ctx` would cancel the delete with it. It is the same move as a
  saga compensation. Or rely on a bucket lifecycle policy.
-->

---
layout: code-walkthrough
heading: Or let the SDK do it
notes:
  - "External Storage: the claim check, built into the Client"
  - A driver for your store. Four methods, Store and Retrieve doing the work
  - Payloads at or over this size are offloaded. 256 KiB is the default
  - "<code>worker.New</code> inherits it from the Client. Same store on both sides"
---

::code::

```go {all|3-6|4|5|12}
c, err := client.Dial(client.Options{
	HostPort: hostPort,
	ExternalStorage: converter.ExternalStorage{
		Drivers:              []converter.StorageDriver{driver},
		PayloadSizeThreshold: 256 * 1024,
	},
})
if err != nil {
	log.Fatalln("Unable to create client", err)
}

w := worker.New(c, taskQueue, worker.Options{})
```

<!--
New to the Go Deck. External Storage is experimental in
go.temporal.io/sdk v1.51.0 (Public Preview on the Service): APIs and
configuration may change before General Availability. See
docs.temporal.io/external-storage.
Code: snippets/claimcheck/extstore.go, run by
snippets/claimcheck/claimcheck_test.go.

The hand-rolled version is worth understanding. The SDK can now do it for you,
transparently, for every payload.

- **Build 1 - `ExternalStorage`.** A field on `client.Options`. Any payload
  over the threshold is uploaded by a driver and replaced in the Event History
  by a small reference; the receiving side downloads it again before your code
  sees it. Workflow and Activity code do not change.
- **Build 2 - the driver.** `converter.StorageDriver` is four methods: `Name`
  and `Type` identify it, `Store` returns one claim per payload, and
  `Retrieve` turns claims back into payloads. The tests behind this Deck use
  an in-memory one; a real one wraps your bucket's client. With several
  drivers, a `DriverSelector` routes each payload.
- **Build 3 - the threshold.** 256 KiB by default; a zero value means the
  default.
- **Build 4 - both sides.** The Worker built from this Client offloads and
  retrieves with the same driver. A Client without it cannot read an
  offloaded result.

This segment comes before Serialization and Encryption, so the room has not
met the data converter yet. One sentence covers it: the data converter is the
SDK's hook for turning every argument and result into bytes on the way out and
back on the way in; the Serialization segment comes back to it.

External Storage runs after the payload codec, so with an encryption codec the
store only ever holds ciphertext.

The same caveat as the hand-rolled version: you own the bucket's lifecycle.
Keep objects for at least as long as the Workflow and its retention period.
-->
