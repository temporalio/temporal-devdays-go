// Package encryption backs 20-encryption.md: an AES-GCM payload codec, held
// in memory here; a real one fetches its key from a KMS.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/proto"
)

// NewAEAD returns AES-256-GCM over a fresh random key.
func NewAEAD() (cipher.AEAD, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// seal encrypts with a random nonce and prepends it.
func (c Codec) seal(plain []byte) []byte {
	nonce := make([]byte, c.AEAD.NonceSize())
	_, _ = rand.Read(nonce)
	return c.AEAD.Seal(nonce, nonce, plain, nil)
}

func (c Codec) open(data []byte) ([]byte, error) {
	n := c.AEAD.NonceSize()
	if len(data) < n {
		return nil, errors.New("ciphertext too short")
	}
	return c.AEAD.Open(nil, data[:n], data[n:], nil)
}

// slide: 20-encryption.md "heading: A codec in practice"
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

// end slide
