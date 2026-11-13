package serialization

import (
	"bytes"
	"encoding/json"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

// A payload converter that keeps every digit of a number held in an `any`.
// The Workflow serializes Activity arguments itself, so this runs inside
// Workflow code: it must be deterministic and do no I/O.

// slide: 19-serialization.md "heading: A custom payload converter"
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

// end slide
