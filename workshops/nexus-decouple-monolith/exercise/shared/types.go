// [GIVEN] The data that crosses team boundaries.
//
// Plain structs with JSON tags, and that is all they need to be. The default
// data converter is JSON, so a struct is enough to cross the Nexus boundary in
// both directions.
package shared

// Where each team runs. One Namespace per team is the whole point.
const (
	PaymentsNamespace   = "payments-namespace"
	ComplianceNamespace = "compliance-namespace"

	PaymentsTaskQueue = "payments-processing"
	// ComplianceTaskQueue MUST match the --target-task-queue of the Nexus
	// Endpoint, or the Endpoint routes work to a queue nobody is polling.
	ComplianceTaskQueue = "compliance-risk"

	// ComplianceEndpoint is the only thing Payments knows about where
	// Compliance runs. The Nexus Registry resolves it to a Namespace and a
	// Task Queue.
	ComplianceEndpoint = "compliance-endpoint"
)

// RiskLevel is the compliance verdict's severity.
type RiskLevel string

const (
	RiskLow    RiskLevel = "LOW"
	RiskMedium RiskLevel = "MEDIUM"
	RiskHigh   RiskLevel = "HIGH"
)

// PaymentStatus is how a payment Workflow ended.
type PaymentStatus string

const (
	StatusCompleted          PaymentStatus = "COMPLETED"
	StatusRejected           PaymentStatus = "REJECTED"
	StatusDeclinedCompliance PaymentStatus = "DECLINED_COMPLIANCE"
	StatusFailed             PaymentStatus = "FAILED"
)

// PaymentRequest is a payment transaction to be processed.
type PaymentRequest struct {
	TransactionID   string  `json:"transactionId"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	SenderCountry   string  `json:"senderCountry"`
	ReceiverCountry string  `json:"receiverCountry"`
	Description     string  `json:"description"`
	SenderAccount   string  `json:"senderAccount"`
	ReceiverAccount string  `json:"receiverAccount"`
}

// PaymentResult is the outcome of a payment Workflow.
type PaymentResult struct {
	Success            bool          `json:"success"`
	TransactionID      string        `json:"transactionId"`
	Status             PaymentStatus `json:"status"`
	RiskLevel          RiskLevel     `json:"riskLevel,omitempty"`
	Explanation        string        `json:"explanation,omitempty"`
	ConfirmationNumber string        `json:"confirmationNumber,omitempty"`
	Error              string        `json:"error,omitempty"`
}

// ComplianceRequest is the input to the compliance check. Sent by Payments to
// Compliance.
type ComplianceRequest struct {
	TransactionID   string  `json:"transactionId"`
	Amount          float64 `json:"amount"`
	SenderCountry   string  `json:"senderCountry"`
	ReceiverCountry string  `json:"receiverCountry"`
	Description     string  `json:"description"`
}

// ComplianceResult is the result of a compliance check. Returned by
// Compliance to Payments.
type ComplianceResult struct {
	TransactionID string    `json:"transactionId"`
	Approved      bool      `json:"approved"`
	RiskLevel     RiskLevel `json:"riskLevel"`
	Explanation   string    `json:"explanation"`
}

// ReviewRequest is a human review decision, submitted through Nexus.
type ReviewRequest struct {
	TransactionID string `json:"transactionId"`
	Approved      bool   `json:"approved"`
	Explanation   string `json:"explanation"`
}
