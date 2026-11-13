// Package sagacompensation backs "A compensation Activity" on
// 17-saga-04-when-to-register.md: a refund that calls the payment API with an
// idempotency key, and Activity Options that give the forward step a deadline
// and the undo none. The payment API is an in-memory fake; tests make the
// refund endpoint fail a set number of times, as if it were down.
package sagacompensation

import (
	"context"
	"errors"
	"sync"
)

type Order struct {
	OrderID     string
	AmountCents int
}

type RefundParams struct {
	IdempotencyKey string
}

type fakePayments struct {
	sync.Mutex
	charged            map[string]bool
	refunds            map[string]string // idempotency key -> order id
	refundCalls        int
	refundFailuresLeft int
}

var payments = &fakePayments{charged: map[string]bool{}, refunds: map[string]string{}}

func (p *fakePayments) Charge(ctx context.Context, orderID string) error {
	p.Lock()
	defer p.Unlock()
	p.charged[orderID] = true
	return nil
}

func (p *fakePayments) Refund(ctx context.Context, orderID string, opts RefundParams) error {
	p.Lock()
	defer p.Unlock()
	p.refundCalls++
	if p.refundFailuresLeft > 0 {
		p.refundFailuresLeft--
		return errors.New("payment API unavailable")
	}
	// The same key twice is one refund, not two.
	if _, ok := p.refunds[opts.IdempotencyKey]; !ok {
		p.refunds[opts.IdempotencyKey] = orderID
	}
	return nil
}

func ChargePayment(ctx context.Context, order Order) error {
	return payments.Charge(ctx, order.OrderID)
}

// part: 17-saga-04-when-to-register.md "heading: A compensation Activity" 1
// activities.go
func RefundPayment(ctx context.Context, order Order) error {
	return payments.Refund(ctx, order.OrderID, RefundParams{
		IdempotencyKey: "refund-" + order.OrderID,
	})
}

// end part
