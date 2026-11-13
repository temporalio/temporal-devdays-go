package claimcheck

import (
	"log"
	"sync"

	"github.com/google/uuid"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
)

// MemoryDriver is a StorageDriver that keeps payloads in memory. A real one
// stores them in a bucket.
type MemoryDriver struct {
	mu     sync.Mutex
	stored map[string]*commonpb.Payload
}

var _ converter.StorageDriver = (*MemoryDriver)(nil)

func NewMemoryDriver() *MemoryDriver {
	return &MemoryDriver{stored: map[string]*commonpb.Payload{}}
}

func (d *MemoryDriver) Name() string { return "memory" }

func (d *MemoryDriver) Type() string { return "example.memory" }

func (d *MemoryDriver) Store(ctx converter.StorageDriverStoreContext, payloads []*commonpb.Payload) ([]converter.StorageDriverClaim, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	claims := make([]converter.StorageDriverClaim, len(payloads))
	for i, p := range payloads {
		key := uuid.NewString()
		d.stored[key] = p
		claims[i] = converter.StorageDriverClaim{ClaimData: map[string]string{"key": key}}
	}
	return claims, nil
}

func (d *MemoryDriver) Retrieve(ctx converter.StorageDriverRetrieveContext, claims []converter.StorageDriverClaim) ([]*commonpb.Payload, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]*commonpb.Payload, len(claims))
	for i, c := range claims {
		out[i] = d.stored[c.ClaimData["key"]]
	}
	return out, nil
}

// Len is the number of payloads stored, for the tests.
func (d *MemoryDriver) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.stored)
}

// Connect dials a client with External Storage and builds a Worker from it.
func Connect(hostPort, taskQueue string, driver converter.StorageDriver) (client.Client, worker.Worker) {
	// slide: 21-claim-check.md "heading: Or let the SDK do it"
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
	// end slide
	return c, w
}
