// Package testenv is the shared harness for the runtime tests: one local dev
// server per test package, started from the `temporal` CLI on PATH, and a
// Worker on a fresh Task Queue for the length of one test.
//
// Each package that runs code against the Service wires it in with
//
//	func TestMain(m *testing.M) { testenv.Main(m) }
package testenv

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

var (
	server *testsuite.DevServer
	c      client.Client
)

// Quiet is a logger that keeps the SDK's own logging out of the test output.
// Failures still surface through the tests' own assertions.
var Quiet log.Logger = log.NewStructuredLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))

// Options lets a package start its dev server with extra flags or Search
// Attributes, or a client with its own Data Converter.
type Options struct {
	ExtraArgs     []string
	ClientOptions client.Options
	Configure     func(*testsuite.DevServerOptions)
}

// Main starts a dev server, runs the package's tests, and stops the server.
func Main(m *testing.M, opts ...Options) {
	os.Exit(run(m, opts...))
}

func run(m *testing.M, opts ...Options) int {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	path, err := exec.LookPath("temporal")
	if err != nil {
		fmt.Fprintln(os.Stderr, "the runtime tests need the temporal CLI on PATH:", err)
		return 1
	}
	co := o.ClientOptions
	co.Logger = Quiet
	dso := testsuite.DevServerOptions{
		ExistingPath:  path,
		ClientOptions: &co,
		LogLevel:      "error",
		ExtraArgs:     o.ExtraArgs,
		Stdout:        io.Discard,
		Stderr:        io.Discard,
	}
	if o.Configure != nil {
		o.Configure(&dso)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	server, err = testsuite.StartDevServer(ctx, dso)
	if err != nil {
		fmt.Fprintln(os.Stderr, "starting the dev server:", err)
		return 1
	}
	defer server.Stop()
	c = server.Client()
	return m.Run()
}

// Client is a client of the package's dev server.
func Client() client.Client { return c }

// HostPort is the dev server's frontend address, for tests that dial their own
// client.
func HostPort() string { return server.FrontendHostPort() }

// ID returns a fresh Workflow ID.
func ID(prefix string) string { return prefix + "-" + uuid.NewString() }

// Worker starts a Worker on a fresh Task Queue, lets register add its
// Workflows and Activities, and stops it when the test ends.
func Worker(t testing.TB, register func(worker.Registry), opts ...worker.Options) string {
	t.Helper()
	tq := ID("snippets")
	WorkerOn(t, tq, register, opts...)
	return tq
}

// WorkerOn starts a Worker on the given Task Queue and stops it when the test
// ends.
func WorkerOn(t testing.TB, taskQueue string, register func(worker.Registry), opts ...worker.Options) worker.Worker {
	t.Helper()
	var o worker.Options
	if len(opts) > 0 {
		o = opts[0]
	}
	w := worker.New(c, taskQueue, o)
	register(w)
	if err := w.Start(); err != nil {
		t.Fatalf("starting the Worker: %v", err)
	}
	t.Cleanup(w.Stop)
	return w
}

// Ctx is a context that ends with the test, bounded so a stuck Workflow fails
// the test instead of hanging it.
func Ctx(t testing.TB) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}
