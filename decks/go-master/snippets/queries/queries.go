// Package queries backs 10-queries.md.
package queries

import "go.temporal.io/sdk/workflow"

type Loan struct {
	LoanID string
}

type LoanResult struct {
	LoanID string
	Status string
}

// slide: 10-queries.md "Go Queries"
const LoanStatusQuery = "getLoanStatus"

func LoanWorkflow(ctx workflow.Context, loan Loan) (LoanResult, error) {
	status := "underwriting"
	err := workflow.SetQueryHandler(ctx, LoanStatusQuery, func() (string, error) {
		return status, nil
	})
	if err != nil {
		return LoanResult{}, err
	}

	status = "approved"
	return LoanResult{LoanID: loan.LoanID, Status: status}, nil
}

// end slide
