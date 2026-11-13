// ABOUTME: What the downstream systems actually did, in order. Reading the ledger
// is how you confirm compensation ran, and in which direction.
//
// In memory on purpose. It stands in for the inventory service and the payment
// gateway, so the lab can prove an effect happened without either. It lives in
// the Worker process, never in Workflow code.
package saga

import (
	"fmt"
	"strings"
	"sync"
)

var ledger struct {
	sync.Mutex
	entries []string
}

// record writes one effect to the ledger and prints it in the Worker terminal.
func record(entry string) {
	ledger.Lock()
	ledger.entries = append(ledger.entries, entry)
	ledger.Unlock()
	fmt.Printf("[ledger] %s\n", entry)
}

// ledgerFor returns what happened to one order, oldest first: "reserve",
// "charge", "refund" and so on.
func ledgerFor(orderID string) []string {
	ledger.Lock()
	defer ledger.Unlock()
	var out []string
	for _, e := range ledger.entries {
		if action, id, ok := strings.Cut(e, ":"); ok && id == orderID {
			out = append(out, action)
		}
	}
	return out
}

// isReserved is true while a reservation stands: reserved more times than released.
func isReserved(orderID string) bool {
	return count("reserve", orderID) > count("release", orderID)
}

// isCharged is true while a charge stands: charged more times than refunded.
func isCharged(orderID string) bool {
	return count("charge", orderID) > count("refund", orderID)
}

func count(action, orderID string) int {
	n := 0
	for _, a := range ledgerFor(orderID) {
		if a == action {
			n++
		}
	}
	return n
}
