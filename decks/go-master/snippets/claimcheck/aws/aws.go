// Package aws stands in for the one helper the claim-check slides use from the
// AWS SDK.
package aws

// String returns a pointer to s.
func String(s string) *string { return &s }
