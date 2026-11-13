package tests

import (
	"reflect"
	"testing"

	"github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"
)

// operation is what nexus.NewOperationReference returns, minus its type
// parameters, so both Operations can be checked by one function.
type operation interface {
	Name() string
	InputType() reflect.Type
	OutputType() reflect.Type
}

// The contract is only names and types. Both teams build against these exact
// values, so check each one.
func TestChallenge2_ContractDeclaresBothOperations(t *testing.T) {
	checkOperation(t, "CheckComplianceOperation", shared.CheckComplianceOperation,
		"checkCompliance", reflect.TypeFor[shared.ComplianceRequest](), reflect.TypeFor[shared.ComplianceResult]())
	checkOperation(t, "SubmitReviewOperation", shared.SubmitReviewOperation,
		"submitReview", reflect.TypeFor[shared.ReviewRequest](), reflect.TypeFor[shared.ComplianceResult]())
}

func checkOperation(t *testing.T, variable string, op operation, name string, in, out reflect.Type) {
	t.Helper()
	want := "nexus.NewOperationReference[" + in.Name() + ", " + out.Name() + "](\"" + name + "\")"
	if op.Name() != name {
		t.Fatalf("shared/service.go, TODO 1: %s must be named %q, not %q. Callers and the handler find an Operation by this name. Write it as %s.", variable, name, op.Name(), want)
	}
	if op.InputType() != in {
		t.Fatalf("shared/service.go, TODO 1: %s must take a %s as its input, not %v. Write it as %s.", variable, in.Name(), op.InputType(), want)
	}
	if op.OutputType() != out {
		t.Fatalf("shared/service.go, TODO 1: %s must return a %s, not %v. Write it as %s.", variable, out.Name(), op.OutputType(), want)
	}
}
