// Package saga backs 17-saga-03-implementation.md: the Saga Workshop's order
// fulfillment Workflow (workshops/saga-order-fulfillment), its five
// Activities, and the in-memory systems they change. needsresult.go backs the
// "register after" side of 17-saga-04-when-to-register.md.
package saga

// Order is the Saga Workshop's order, with its three scenario switches.
type Order struct {
	OrderID     string
	SKU         string
	Quantity    int
	AmountCents int
	// FailReserve makes the reservation fail before it records anything.
	FailReserve bool
	// FailDispatch makes courier dispatch fail after two steps landed.
	FailDispatch bool
	// ChargeFailsAfterEffect makes the charge land, then report failure anyway.
	ChargeFailsAfterEffect bool
}

type OrderResult struct {
	OrderID        string
	TrackingNumber string
}
