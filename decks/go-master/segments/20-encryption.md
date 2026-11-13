---
layout: section
---

<ZiggyHero scene="lock">

# Encryption

Payloads leave your Worker. Encrypt them before they do.

</ZiggyHero>

<!--
Source: PPT slides 63-64, `{{SLIDE:encryption-codec-in-practice}}`. The
registration and Codec Server slides are new to the Go Deck.

PORTED: the Go interface is `converter.PayloadCodec`, with `Encode`/`Decode`
over `[]*commonpb.Payload`, plain blocking calls like every other Go call.
-->

---
layout: code-walkthrough
heading: A codec in practice
notes:
  - "<code>PayloadCodec</code>: two methods over slices of Payloads. This line proves Codec satisfies it"
  - "<code>Encode</code> runs after the payload converter, before anything leaves the process"
  - The whole Payload, metadata included, is marshaled and encrypted
  - A new encoding label, so the other side knows to decrypt
  - "<code>Decode</code> runs when data arrives back"
  - Not encrypted? Pass it through untouched
---

::code::

```go {all|3|7|10|15|22|25-28}
const EncodingEncrypted = "binary/encrypted"

var _ converter.PayloadCodec = Codec{}

type Codec struct{ AEAD cipher.AEAD }

func (c Codec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	out := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		plain, err := proto.Marshal(p)
		if err != nil {
			return nil, err
		}
		out[i] = &commonpb.Payload{
			Metadata: map[string][]byte{converter.MetadataEncoding: []byte(EncodingEncrypted)},
			Data:     c.seal(plain),
		}
	}
	return out, nil
}

func (c Codec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	out := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		if string(p.Metadata[converter.MetadataEncoding]) != EncodingEncrypted {
			out[i] = p
			continue
		}
		plain, err := c.open(p.Data)
		if err != nil {
			return nil, err
		}
		out[i] = &commonpb.Payload{}
		if err := proto.Unmarshal(plain, out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}
```

<!--
Source: PPT slide 64, ported to Go. The imports and the two small helpers are
cut for room: `seal`/`open` are AES-256-GCM from `crypto/cipher` with a random
nonce prepended to the ciphertext, and `proto` is
`google.golang.org/protobuf/proto`. This is the shape of Temporal's own Go
encryption sample.
Code: snippets/encryption/codec.go, run by snippets/encryption/encryption_test.go.

Walk the interface. Two methods, both over slices.

- **Build 1 - the interface.** `converter.PayloadCodec` is the whole contract:
  `Encode` and `Decode`. Go interfaces are satisfied implicitly, so the
  `var _ converter.PayloadCodec = Codec{}` line is there to make the compiler
  say so, and to fail the build if a signature drifts.
- **Build 2 - `Encode`.** Called on the way out, after the payload converter
  from the last segment has turned your values into Payloads, and before
  anything is sent to the Temporal Service. This is the last point at which
  the bytes are yours.
- **Build 3 - what gets encrypted.** The entire Payload, including its original
  encoding label, so nothing about the value leaks.
- **Build 4 - the label.** The outer Payload says only `binary/encrypted`. A
  real codec also records a key ID here, so keys can rotate.
- **Build 5 - `Decode`.** Called on the way in. Symmetric by construction.
- **Build 6 - pass-through.** Anything not encrypted, such as data written
  before you turned encryption on, is left alone.

The contrast with the payload converter, and the reason both exist: the
converter turns values into Payloads; a codec turns Payloads into other
Payloads, so it composes with any converter and stays out of your types. That
is where encryption and compression belong; the SDK ships
`converter.NewZlibCodec` for the second.

The Go-specific caveat: the codec runs wherever the data converter runs,
including inside Workflow code, which is under the deadlock detector's
one-second budget. A codec that calls a KMS on every payload can trip it.
Cache data keys in the codec, or wrap the converter with
`workflow.DataConverterWithoutDeadlockDetection`.
-->

---
layout: code-walkthrough
heading: Same codec, Client and Worker
notes:
  - "<code>NewCodecDataConverter</code> wraps the default converter with the codec"
  - The Client encodes Workflow arguments and decodes results
  - "<code>worker.New</code> inherits it from the Client. Workflow code does not change at all"
---

::code::

```go {all|3-6|1-7|12}
c, err := client.Dial(client.Options{
	HostPort: hostPort,
	DataConverter: converter.NewCodecDataConverter(
		converter.GetDefaultDataConverter(),
		Codec{AEAD: aead},
	),
})
if err != nil {
	log.Fatalln("Unable to create client", err)
}

w := worker.New(c, taskQueue, worker.Options{})
```

<!--
Source: PPT slide 64, the registration lines of the original, split out.
Code: snippets/encryption/client.go.

- **Build 1 - the codec chain.** `NewCodecDataConverter` takes the converter
  to wrap, then any number of codecs. Encoding applies them last to first, so
  the first one listed is the outermost: `NewCodecDataConverter(dc, encryption,
  compression)` compresses first, then encrypts.
- **Build 2 - the Client.** Miss it here and the Client cannot read the result:
  it gets a `binary/encrypted` Payload and `Get` returns "payload encoding is
  not supported". At runtime, not compile time.
- **Build 3 - the Worker.** `worker.New` takes the Client and its converter:
  same codec, same key.

The claim worth making explicitly, because it is the reason the pattern exists:
Workflow code does not change at all. Not one line. Encryption is a transport
concern and it stays one. The tests behind this Deck run a Workflow whose code
has never heard of encryption, and check that the Event History holds only
ciphertext.
-->

---
layout: default
---

# Who can read what

- **Temporal Service:** ciphertext only. It never has the key
- **Your Clients and Workers:** plaintext, through the codec
- **Temporal UI and CLI:** ciphertext, until you run a **Codec Server**

<!--
Source: PPT slide 64's speaker notes, given their own slide here.

The consequence somebody always spots: the Temporal Service now stores bytes it
cannot read, so the Temporal UI and the CLI show ciphertext.

The fix is a Codec Server: a small HTTP service you run that exposes the same
codec's `Decode` (and `Encode`). In Go it is one line:
`http.ListenAndServe(":8081", converter.NewPayloadCodecHTTPHandler(Codec{AEAD:
aead}))`, which serves `/encode` and `/decode`; the tests behind this Deck
decode a payload through it. The Temporal UI calls it from the user's browser
to decode for display, and the CLI takes `--codec-endpoint`. The keys stay on
your side; the Temporal Service never sees them. Put authentication and CORS
in front of it, because anyone who can call it can decrypt.

Mention it; do not demo it unless asked.
-->
