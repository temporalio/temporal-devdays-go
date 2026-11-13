// Package antipatterns backs the ❌/✅ slides in
// 17b-patterns-and-anti-patterns.md and the idempotency key slide in
// 17-saga-07-related.md.
package antipatterns

import (
	"context"
	"errors"
	"sync"
)

type Debit struct {
	AccountID string
	Amount    int
}

type posting struct {
	key   string
	debit Debit
}

// fakeLedger is a stand-in core-banking ledger. It ignores a posting whose
// idempotency key it has already seen, and can be told to fail once after
// committing: the timeout-after-a-successful-write that makes a retry
// dangerous.
type fakeLedger struct {
	sync.Mutex
	postings            []posting
	requests            []string
	failuresAfterCommit int
}

var ledger = &fakeLedger{}

func (l *fakeLedger) reset() {
	l.Lock()
	defer l.Unlock()
	l.postings, l.requests, l.failuresAfterCommit = nil, nil, 0
}

func (l *fakeLedger) failOnceAfterCommit() {
	l.Lock()
	defer l.Unlock()
	l.failuresAfterCommit = 1
}

func (l *fakeLedger) Post(ctx context.Context, debit Debit, idempotencyKey string) (string, error) {
	l.Lock()
	defer l.Unlock()
	l.requests = append(l.requests, idempotencyKey)
	seen := false
	for _, p := range l.postings {
		seen = seen || p.key == idempotencyKey
	}
	if !seen {
		l.postings = append(l.postings, posting{idempotencyKey, debit})
	}
	if l.failuresAfterCommit > 0 {
		l.failuresAfterCommit--
		return "", errors.New("connection reset after commit")
	}
	return idempotencyKey, nil
}

// fakeBank releases held deposits for the ❌ Activity on the sleeping slide.
type fakeBank struct{}

var bank fakeBank

func (fakeBank) Release(ctx context.Context, depositID string) error {
	return ReleaseFunds(ctx, depositID)
}
