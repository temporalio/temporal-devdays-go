package saga

import (
	"strings"
	"sync"
)

// ledger is what the downstream systems actually did, in order: the inventory
// service and the payment gateway the lab's Activities stand in for. It lives
// in the Worker process, never in Workflow code, so the tests can read it
// after a run.
var ledger struct {
	sync.Mutex
	entries []string
}

func record(entry string) {
	ledger.Lock()
	defer ledger.Unlock()
	ledger.entries = append(ledger.entries, entry)
}

// ForOrder returns every entry for one order, oldest first, without the order
// id.
func ForOrder(orderID string) []string {
	ledger.Lock()
	defer ledger.Unlock()
	out := []string{}
	for _, e := range ledger.entries {
		if op, ok := strings.CutSuffix(e, ":"+orderID); ok {
			out = append(out, op)
		}
	}
	return out
}

// isReserved is true while a reservation stands: reserved more times than
// released.
func isReserved(orderID string) bool {
	return count("reserve:"+orderID) > count("release:"+orderID)
}

// isCharged is true while a charge stands: charged more times than refunded.
func isCharged(orderID string) bool {
	return count("charge:"+orderID) > count("refund:"+orderID)
}

func count(entry string) int {
	ledger.Lock()
	defer ledger.Unlock()
	n := 0
	for _, e := range ledger.entries {
		if e == entry {
			n++
		}
	}
	return n
}
