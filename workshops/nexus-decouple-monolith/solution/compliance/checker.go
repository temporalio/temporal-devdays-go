// [GIVEN] Rule-based compliance checker. Deterministic, no API keys needed.
// Same logic whether it runs as an Activity or behind a Nexus handler.
//
// Rules:
//   - OFAC sanctioned country on either side   -> HIGH risk, blocked
//   - Amount over $50,000                       -> HIGH risk, blocked (TXN-C hits this)
//   - Over $10,000 or an unusual jurisdiction   -> MEDIUM risk (TXN-B hits this)
//   - Everything else                           -> LOW risk, approved
package compliance

import "github.com/temporalio/temporal-devdays-go/workshops/nexus-decouple-monolith/shared"

var (
	sanctioned = map[string]bool{"North Korea": true, "Iran": true, "Cuba": true, "Syria": true, "Venezuela": true}
	common     = map[string]bool{"US": true, "UK": true, "Canada": true, "Germany": true, "France": true, "Japan": true, "Australia": true}
)

// Check scores one transaction.
func Check(request shared.ComplianceRequest) shared.ComplianceResult {
	result := shared.ComplianceResult{TransactionID: request.TransactionID}

	switch {
	case sanctioned[request.ReceiverCountry] || sanctioned[request.SenderCountry]:
		result.RiskLevel = shared.RiskHigh
		result.Explanation = "Destination/source country is OFAC-sanctioned. Transaction blocked per regulatory requirements."

	case request.Amount > 50_000:
		result.RiskLevel = shared.RiskHigh
		result.Explanation = "Transaction amount exceeds $50,000 threshold. Requires enhanced due diligence review."

	case request.Amount > 10_000 ||
		(request.SenderCountry != request.ReceiverCountry && !common[request.ReceiverCountry]):
		result.Approved = true
		result.RiskLevel = shared.RiskMedium
		result.Explanation = "International transfer above $10K threshold. Approved with AML monitoring note."

	default:
		result.Approved = true
		result.RiskLevel = shared.RiskLow
		result.Explanation = "Routine domestic/standard international transfer. No regulatory concerns."
	}
	return result
}
