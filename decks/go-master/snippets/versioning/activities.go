// Package versioning backs 18-versioning.md. The Workflow's four versions
// live in the subpackages before, patched, deprecated and final; each declares
// MyWorkflow, so all four register under the one Workflow Type the replay
// tests need.
package versioning

import "context"

// The stub Activities each return their own name, so a test can read the
// Event History and see which branch a run took.

func ActivityA(ctx context.Context) (string, error) { return "A", nil }

func ActivityB(ctx context.Context) (string, error) { return "B", nil }

func ActivityC(ctx context.Context) (string, error) { return "C", nil }
