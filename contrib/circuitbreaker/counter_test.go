package circuitbreaker

import (
	"testing"
)

//nolint:lll,exhaustruct // temporary solution
func TestCountsMethods(t *testing.T) {
	counts := Counts{}

	counts.onRequest()
	assertEqual(t, Counts{Requests: 1}, counts)

	counts.onSuccess()
	assertEqual(t, Counts{Requests: 1, TotalSuccesses: 1, ConsecutiveSuccesses: 1}, counts)

	counts.onRequest()
	assertEqual(t, Counts{Requests: 2, TotalSuccesses: 1, ConsecutiveSuccesses: 1}, counts)

	counts.onSuccess()
	assertEqual(t, Counts{Requests: 2, TotalSuccesses: 2, ConsecutiveSuccesses: 2}, counts)

	counts.onRequest()
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 2, ConsecutiveSuccesses: 2}, counts)

	counts.onFailure()
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, counts)

	counts.onRequest()
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, counts)

	counts.onFailure()
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 2}, counts)

	counts.onRequest()
	assertEqual(t, Counts{Requests: 5, TotalSuccesses: 2, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 2}, counts)

	counts.clear()
	assertEqual(t, Counts{}, counts)
}

func TestNewRollingCounts(t *testing.T) {
	counts := newRollingCounts(-1)
	assertEqual(t, uint64(0), counts.age)
	assertEqual(t, 0, len(counts.buckets))
	assertEqual(t, Counts{}, counts.Counts)

	counts = newRollingCounts(0)
	assertEqual(t, uint64(0), counts.age)
	assertEqual(t, 0, len(counts.buckets))
	assertEqual(t, Counts{}, counts.Counts)

	counts = newRollingCounts(5)
	assertEqual(t, uint64(0), counts.age)
	assertEqual(t, 5, len(counts.buckets))
	assertEqual(t, Counts{}, counts.Counts)
	for i := range counts.buckets {
		assertEqual(t, Counts{}, counts.buckets[i])
	}
}

func TestRollingCountsIndex(t *testing.T) {
	rc := newRollingCounts(0)
	assertEqual(t, uint64(0), rc.index(0))
	assertEqual(t, uint64(0), rc.index(1))

	rc = newRollingCounts(5)
	assertEqual(t, uint64(0), rc.index(0))
	assertEqual(t, uint64(1), rc.index(1))
	assertEqual(t, uint64(2), rc.index(2))
	assertEqual(t, uint64(3), rc.index(3))
	assertEqual(t, uint64(4), rc.index(4))
	assertEqual(t, uint64(0), rc.index(5))
	assertEqual(t, uint64(1), rc.index(6))
}

func TestRollingCountsCurrent(t *testing.T) {
	rc := newRollingCounts(0)
	assertEqual(t, uint64(0), rc.current())
	rc.roll()
	assertEqual(t, uint64(0), rc.current())

	rc = newRollingCounts(5)
	assertEqual(t, uint64(0), rc.current())
	rc.roll()
	assertEqual(t, uint64(1), rc.current())
	rc.roll()
	assertEqual(t, uint64(2), rc.current())
	rc.roll()
	assertEqual(t, uint64(3), rc.current())
	rc.roll()
	assertEqual(t, uint64(4), rc.current())
	rc.roll()
	assertEqual(t, uint64(0), rc.current())
	rc.roll()
	assertEqual(t, uint64(1), rc.current())
}

func TestRollingCountsMethods(t *testing.T) {
	rc := newRollingCounts(2)
	assertEqual(t, uint64(0), rc.age)
	assertEqual(t, Counts{}, rc.Counts)
	assertEqual(t, Counts{}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	rc.onRequest()
	assertEqual(t, Counts{Requests: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 1}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	rc.onSuccess(0)
	assertEqual(t, Counts{Requests: 1, TotalSuccesses: 1, ConsecutiveSuccesses: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 1, TotalSuccesses: 1, ConsecutiveSuccesses: 1}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	rc.onRequest()
	assertEqual(t, Counts{Requests: 2, TotalSuccesses: 1, ConsecutiveSuccesses: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 2, TotalSuccesses: 1, ConsecutiveSuccesses: 1}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	rc.onFailure(0)
	assertEqual(t, Counts{Requests: 2, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 2, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	rc.onRequest()
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	rc.onSuccess(0)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.Counts)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	rc.onRequest()
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.Counts)
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.Counts)
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	rc.roll()
	assertEqual(t, uint64(1), rc.age)
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.Counts)
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])

	rc.onRequest()
	assertEqual(t, Counts{Requests: 5, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.Counts)
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{Requests: 1}, rc.buckets[rc.index(1)])

	rc.onSuccess(1)
	assertEqual(t, Counts{Requests: 5, TotalSuccesses: 3, TotalFailures: 1, ConsecutiveSuccesses: 2, ConsecutiveFailures: 0}, rc.Counts)
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 2, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{Requests: 1, TotalSuccesses: 1, ConsecutiveSuccesses: 1}, rc.buckets[rc.index(1)])

	rc.roll()
	assertEqual(t, uint64(2), rc.age)
	assertEqual(t, Counts{Requests: 1, TotalSuccesses: 1, ConsecutiveSuccesses: 1}, rc.Counts)
	assertEqual(t, Counts{}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{Requests: 1, TotalSuccesses: 1, ConsecutiveSuccesses: 1}, rc.buckets[rc.index(1)])

	rc.clear()
	assertEqual(t, uint64(0), rc.age)
	assertEqual(t, Counts{}, rc.Counts)
	assertEqual(t, Counts{}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])
}

func TestRollingCountsGrow(t *testing.T) {
	rc := newRollingCounts(2)

	rc.onRequest()
	rc.onSuccess(0)
	rc.onRequest()
	rc.onFailure(0)
	rc.onRequest()

	rc.grow(0) // no change
	assertEqual(t, uint64(0), rc.age)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))

	rc.grow(1)
	assertEqual(t, uint64(1), rc.age)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{}, rc.bucketAt(0))
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(1))

	rc.grow(0) // no change
	assertEqual(t, uint64(1), rc.age)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{}, rc.bucketAt(0))
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(1))

	rc.grow(2)
	assertEqual(t, uint64(2), rc.age)
	assertEqual(t, Counts{}, rc.Counts)
	assertEqual(t, Counts{}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))

	rc.onRequest()
	rc.onSuccess(0)
	rc.onRequest()
	rc.onFailure(0)
	rc.onRequest()

	assertEqual(t, uint64(2), rc.age)
	assertEqual(t, Counts{Requests: 3}, rc.Counts)
	assertEqual(t, Counts{Requests: 3}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))

	rc.onRequest()
	rc.onSuccess(3)
	rc.onRequest()
	rc.onFailure(3)
	rc.onRequest()

	assertEqual(t, uint64(2), rc.age)
	assertEqual(t, Counts{Requests: 6}, rc.Counts)
	assertEqual(t, Counts{Requests: 6}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))

	rc.onRequest()
	rc.onSuccess(1)
	rc.onRequest()
	rc.onFailure(1)
	rc.onRequest()

	assertEqual(t, uint64(2), rc.age)
	assertEqual(t, Counts{Requests: 9, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 9}, rc.bucketAt(0))
	assertEqual(t, Counts{TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(1))

	rc.onRequest()
	rc.onSuccess(2)
	rc.onRequest()
	rc.onFailure(2)
	rc.onRequest()

	assertEqual(t, uint64(2), rc.age)
	assertEqual(t, Counts{Requests: 12, TotalSuccesses: 2, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 12, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(0))
	assertEqual(t, Counts{TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(1))

	rc.grow(4)
	assertEqual(t, uint64(4), rc.age)
	assertEqual(t, Counts{}, rc.Counts)
	assertEqual(t, Counts{}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))
}

func TestRollingCountsBucketAt(t *testing.T) {
	rc := newRollingCounts(2)
	assertEqual(t, uint64(0), rc.age)
	assertEqual(t, Counts{}, rc.Counts)
	assertEqual(t, Counts{}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))

	rc.onRequest()
	assertEqual(t, Counts{Requests: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 1}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))

	rc.onFailure(0)
	assertEqual(t, Counts{Requests: 1, TotalFailures: 1, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 1, TotalFailures: 1, ConsecutiveFailures: 1}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))

	rc.onRequest()
	assertEqual(t, Counts{Requests: 2, TotalFailures: 1, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 2, TotalFailures: 1, ConsecutiveFailures: 1}, rc.bucketAt(2))
	assertEqual(t, Counts{}, rc.bucketAt(3))

	rc.onSuccess(0)
	assertEqual(t, Counts{Requests: 2, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.Counts)
	assertEqual(t, Counts{Requests: 2, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.bucketAt(-2))
	assertEqual(t, Counts{}, rc.bucketAt(-3))

	rc.onRequest()
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.Counts)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 1, ConsecutiveSuccesses: 1, ConsecutiveFailures: 0}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))

	rc.onFailure(0)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(0))
	assertEqual(t, Counts{}, rc.bucketAt(1))

	rc.roll()
	assertEqual(t, uint64(1), rc.age)
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{}, rc.bucketAt(0))
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(1))

	rc.onRequest()
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 1, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{Requests: 1}, rc.bucketAt(0))
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(1))

	rc.onFailure(1)
	assertEqual(t, Counts{Requests: 4, TotalSuccesses: 1, TotalFailures: 3, ConsecutiveSuccesses: 0, ConsecutiveFailures: 2}, rc.Counts)
	assertEqual(t, Counts{Requests: 1, TotalFailures: 1, ConsecutiveFailures: 1}, rc.bucketAt(0))
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(1))

	rc.onRequest()
	assertEqual(t, Counts{Requests: 5, TotalSuccesses: 1, TotalFailures: 3, ConsecutiveSuccesses: 0, ConsecutiveFailures: 2}, rc.Counts)
	assertEqual(t, Counts{Requests: 2, TotalFailures: 1, ConsecutiveFailures: 1}, rc.bucketAt(0))
	assertEqual(t, Counts{Requests: 3, TotalSuccesses: 1, TotalFailures: 2, ConsecutiveSuccesses: 0, ConsecutiveFailures: 1}, rc.bucketAt(1))

	rc.roll()
	assertEqual(t, uint64(2), rc.age)
	assertEqual(t, Counts{Requests: 2, TotalFailures: 1, ConsecutiveFailures: 1}, rc.Counts)
	assertEqual(t, Counts{}, rc.bucketAt(0))
	assertEqual(t, Counts{Requests: 2, TotalFailures: 1, ConsecutiveFailures: 1}, rc.bucketAt(1))

	rc.clear()
	assertEqual(t, uint64(0), rc.age)
	assertEqual(t, Counts{}, rc.Counts)
	assertEqual(t, Counts{}, rc.buckets[rc.index(0)])
	assertEqual(t, Counts{}, rc.buckets[rc.index(1)])
}

func TestEmptyRollingCounts(t *testing.T) {
	rc := newRollingCounts(0)
	rc.subtract(0) // no change
	assertEqual(t, Counts{}, rc.bucketAt(0))
}
