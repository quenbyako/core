package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

var (
	errFailed          = errors.New("failed")
	errExcluded        = errors.New("excluded")
	ErrTooManyRequests = errors.New("too many requests")
)

// Test-only methods for CircuitBreaker
func (c *CircuitBreaker) Counts() Counts {
	return c.counts.Counts
}

func (c *CircuitBreaker) Name() string {
	return getTestName(c)
}

// Test name registry
var testNames = make(map[*CircuitBreaker]string)

func setTestName(cb *CircuitBreaker, name string) {
	testNames[cb] = name
}

func getTestName(cb *CircuitBreaker) string {
	return testNames[cb]
}

// pseudoSleep manipulates the internal state of CircuitBreaker to simulate time passing
func pseudoSleep(cb *CircuitBreaker, period time.Duration) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.start = cb.start.Add(-period)
	if !cb.expiry.IsZero() {
		cb.expiry = cb.expiry.Add(-period)
	}
}

func assertError(t *testing.T, v error) {
	t.Helper()

	if v == nil {
		t.Fatalf("expected error, got nil")
	}
}

func assertEqual[T comparable](t *testing.T, a, b T) {
	t.Helper()

	if a != b {
		t.Fatalf("expected %v, got %v", b, a)
	}
}

func assertNil(t *testing.T, v any) {
	t.Helper()

	if v != nil && !isNil(v) {
		t.Fatalf("expected nil, got %v", v)
	}
}

func assertTrue(t *testing.T, v bool) {
	t.Helper()

	if !v {
		t.Fatalf("expected true, got false")
	}
}

func assertFalse(t *testing.T, v bool) {
	t.Helper()

	if v {
		t.Fatalf("expected false, got true")
	}
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch v := i.(type) {
	case error:
		return v == nil
	}
	return false
}
