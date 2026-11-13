package encryption

import (
	"crypto/cipher"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// Greet knows nothing about encryption.
func Greet(ctx workflow.Context, patientName string) (string, error) {
	return "Hello, " + patientName, nil
}

// Connect dials a client that encrypts every payload, and builds a Worker
// from it.
func Connect(hostPort, taskQueue string, aead cipher.AEAD) (client.Client, worker.Worker) {
	// slide: 20-encryption.md "heading: Same codec, Client and Worker"
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
	// end slide
	return c, w
}
