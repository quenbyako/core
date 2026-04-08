package core

// State represents the current state of a [CircuitBreaker].
type State int

const (
	// StateClosed allows requests to proceed normally and tracks failures.
	StateClosed State = iota
	// StateHalfOpen allows a limited number of requests to test if the
	// dependency has recovered.
	StateHalfOpen
	// StateOpen fails requests immediately until a timeout expires.
	StateOpen
)

// CircuitBreaker abstracts the common functionality of a circuit breaker.
// It uses a two-step approach: first, it checks if a request is allowed,
// and second, it reports the result of the request.
type CircuitBreaker interface {
	// Allow reports whether the request is permitted to proceed. If allowed (ok
	// == true), the returned [ResultReporter] MUST be called exactly once when
	// the request completes: Success on a good outcome, Fail on a bad one. If
	// not allowed (ok == false), r is nil; inspect [State] for the reason.
	Allow() (ResultReporter, bool)

	// State returns the current state of the circuit breaker.
	State() State
}

type ResultReporter interface {
	Success()
	Fail()
}

// NoopCircuitBreaker returns a [CircuitBreaker] implementation that always
// allows requests and does nothing when 'done' is called.
//
//nolint:ireturn // interface is required to make an abstraction.
func NoopCircuitBreaker() CircuitBreaker {
	return noopCircuitBreaker{}
}

type noopCircuitBreaker struct{}

var _ CircuitBreaker = noopCircuitBreaker{}

//nolint:ireturn // interface requirement
func (noopCircuitBreaker) Allow() (ResultReporter, bool) {
	return noopReporter{}, true
}

func (noopCircuitBreaker) State() State {
	return StateClosed
}

type noopReporter struct{}

func (noopReporter) Success() {}
func (noopReporter) Fail()    {}
