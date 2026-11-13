// Package workflows backs 05-workflows.md: the Workflow anatomy walkthrough
// and the determinism rules.
package workflows

type Transfer struct {
	TransferID  string
	FromAccount string
	ToAccount   string
	Amount      int
}

type Receipt struct {
	Status string // "completed", "rejected" or "refunded"
	Reason string
}
