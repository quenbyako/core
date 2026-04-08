package circuitbreaker

import (
	"errors"
	"sync"
	"time"

	"github.com/quenbyako/core"
)

type CircuitBreaker struct {
	now func() time.Time

	maxRequests  uint32
	interval     time.Duration
	bucketPeriod time.Duration
	timeout      time.Duration
	readyToTrip  func(counts Counts) bool
	callback     func(old, new core.State)

	mutex      sync.Mutex
	state      core.State
	generation uint64
	counts     *rollingCounts
	start      time.Time
	expiry     time.Time
}

const (
	defaultInterval = time.Duration(0) * time.Second
	defaultTimeout  = time.Duration(60) * time.Second
)

func defaultReadyToTrip(counts Counts) bool {
	return counts.ConsecutiveFailures > 5
}

var _ core.CircuitBreaker = (*CircuitBreaker)(nil)

type newParams struct {
	maxRequests  uint32
	interval     time.Duration
	bucketPeriod time.Duration
	timeout      time.Duration
	readyToTrip  func(counts Counts) bool
	now          func() time.Time
	callback     func(old, new core.State)
}

func (p *newParams) validate() error {
	if p.maxRequests == 0 {
		return errors.New("maxRequests must be greater than 0")
	}

	return nil
}

type NewOption func(*newParams)

func WithMaxRequests(maxRequests uint32) NewOption {
	return func(p *newParams) { p.maxRequests = maxRequests }
}

func WithInterval(interval time.Duration) NewOption {
	return func(p *newParams) { p.interval = interval }
}

func WithBucketPeriod(bucketPeriod time.Duration) NewOption {
	return func(p *newParams) { p.bucketPeriod = bucketPeriod }
}

func WithTimeout(timeout time.Duration) NewOption {
	return func(p *newParams) { p.timeout = timeout }
}

func WithReadyToTrip(readyToTrip func(counts Counts) bool) NewOption {
	return func(p *newParams) { p.readyToTrip = readyToTrip }
}

func WithNow(now func() time.Time) NewOption {
	return func(p *newParams) { p.now = now }
}

func WithStateChangeCallback(callback func(oldState, newState core.State)) NewOption {
	return func(p *newParams) { p.callback = callback }
}

func New(opts ...NewOption) (core.CircuitBreaker, error) {
	params := newParams{
		maxRequests:  1,
		interval:     defaultInterval,
		bucketPeriod: time.Duration(0) * time.Second,
		timeout:      defaultTimeout,
		readyToTrip:  defaultReadyToTrip,
		now:          time.Now,
		callback:     func(old, new core.State) {},
	}

	for _, opt := range opts {
		opt(&params)
	}

	if err := params.validate(); err != nil {
		return nil, err
	}

	cb := CircuitBreaker{
		now:         params.now,
		maxRequests: params.maxRequests,
		timeout:     params.timeout,
		readyToTrip: params.readyToTrip,
		callback:    params.callback,
	}

	var numBuckets int64
	if params.interval <= 0 {
		cb.interval = defaultInterval
		cb.bucketPeriod = cb.interval
		numBuckets = 1
	} else if params.bucketPeriod <= 0 {
		cb.interval = params.interval
		cb.bucketPeriod = cb.interval
		numBuckets = 1
	} else {
		cb.interval = (params.interval + params.bucketPeriod - 1) / params.bucketPeriod * params.bucketPeriod
		cb.bucketPeriod = params.bucketPeriod
		numBuckets = int64(cb.interval / cb.bucketPeriod)
	}

	cb.counts = newRollingCounts(numBuckets)

	cb.toNewGeneration(params.now())

	return &cb, nil
}

//nolint:ireturn // interface requirement
func (c *CircuitBreaker) Allow() (core.ResultReporter, bool) {
	if generation, age, ok := c.beforeRequest(); ok {
		return &callback{cb: c, generation: generation, age: age}, true
	}

	return nil, false
}

func (c *CircuitBreaker) State() core.State {
	state, _, _ := c.currentState(c.now())
	return state
}

type callback struct {
	cb         *CircuitBreaker
	generation uint64
	age        uint64
}

var _ core.ResultReporter = (*callback)(nil)

func (c *callback) Success() {
	c.cb.afterRequest(c.generation, c.age, true)
}

func (c *callback) Fail() {
	c.cb.afterRequest(c.generation, c.age, false)
}

func (c *CircuitBreaker) beforeRequest() (uint64, uint64, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := c.now()
	state, generation, age := c.currentState(now)

	if state == core.StateOpen {
		return 0, 0, false
	} else if state == core.StateHalfOpen && c.counts.validRequests() >= c.maxRequests {
		return 0, 0, false
	}

	c.counts.onRequest()
	return generation, age, true
}

func (c *CircuitBreaker) currentState(now time.Time) (core.State, uint64, uint64) {
	switch c.state {
	case core.StateClosed:
		if !c.expiry.IsZero() && c.expiry.Before(now) {
			c.toNewGeneration(now)
		} else if len(c.counts.buckets) >= 2 {
			c.counts.grow(c.age(now))
		}
	case core.StateOpen:
		if c.expiry.Before(now) {
			c.setState(core.StateHalfOpen, now)
		}
	}
	return c.state, c.generation, c.counts.age
}

func (c *CircuitBreaker) setState(state core.State, now time.Time) {
	if c.state == state {
		return
	}

	c.callback(c.state, state)
	c.state = state

	c.toNewGeneration(now)
}

func (c *CircuitBreaker) toNewGeneration(now time.Time) {
	c.generation++
	c.start = now
	c.counts.clear()

	var zero time.Time
	switch c.state {
	case core.StateClosed:
		if c.interval == 0 || len(c.counts.buckets) >= 2 {
			c.expiry = zero
		} else {
			c.expiry = now.Add(c.interval)
		}
	case core.StateOpen:
		c.expiry = now.Add(c.timeout)
	default: // StateHalfOpen
		c.expiry = zero
	}
}

func (c *CircuitBreaker) age(now time.Time) uint64 {
	if c.bucketPeriod == 0 {
		return 0
	}

	elapsed := now.Sub(c.start)
	age := int64(elapsed / c.bucketPeriod)
	if age < 0 {
		return 0
	}
	return uint64(age)
}

func (c *CircuitBreaker) afterRequest(previous uint64, age uint64, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := c.now()
	state, generation, _ := c.currentState(now)
	if generation != previous {
		return
	}

	if ok {
		c.onSuccess(state, age, now)
	} else {
		c.onFailure(state, age, now)
	}
}

func (c *CircuitBreaker) onSuccess(state core.State, age uint64, now time.Time) {
	switch state {
	case core.StateClosed:
		c.counts.onSuccess(age)
	case core.StateHalfOpen:
		c.counts.onSuccess(age)

		if c.counts.ConsecutiveSuccesses >= c.maxRequests {
			c.setState(core.StateClosed, now)
		}
	case core.StateOpen:
		// do nothing
	}
}

func (c *CircuitBreaker) onFailure(state core.State, age uint64, now time.Time) {
	switch state {
	case core.StateClosed:
		c.counts.onFailure(age)

		if c.readyToTrip(c.counts.Counts) {
			c.setState(core.StateOpen, now)
		}
	case core.StateHalfOpen:
		c.setState(core.StateOpen, now)

	case core.StateOpen:
		// do nothing
	}
}
