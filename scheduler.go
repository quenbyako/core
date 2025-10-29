package core

import (
	"context"
	"errors"
	"sync"
)

// RunJobs concurrently executes the provided job functions, cancelling all
// remaining work as soon as any job returns a non-context error. Each job
// receives a shared derived context that is cancelled on the first failure
// (excluding correct context cancellations).
//
// Execution Model:
//   - All jobs start immediately in separate goroutines.
//   - A derived context is cancelled after the first non-context error.
//   - Subsequent jobs should observe cancellation and return promptly.
//
// Error Handling:
//   - Errors equal to [context.Canceled] or [context.DeadlineExceeded] are
//     suppressed.
//   - All other errors are collected and returned via [errors.Join].
//   - If no non-context errors occur, the function returns nil.
//
// Cancellation Semantics:
//   - Caller cancellation (ctx) propagates to all jobs.
//   - Internal cancellation (first failure) does not mask earlier successful
//     results.
//
// Ordering & Aggregation:
//   - Error order is not guaranteed.
//   - Multiple failures are joined; callers should inspect [errors.Is]/[errors.As].
//
// Usage Example:
//
//	err := RunJobs(ctx,
//	    func(c context.Context) error { return workA(c) },
//	    func(c context.Context) error { return workB(c) },
//	)
//	if err != nil {
//	    // handle joined errors
//	}
func RunJobs(ctx context.Context, jobs ...func(context.Context) error) error {
	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		errs    []error
		errsMux sync.Mutex
		wg      sync.WaitGroup
	)

	wg.Add(len(jobs))

	for i, job := range jobs {
		go func(_ int, job func(context.Context) error) {
			if err := omitContextErr(job(jobCtx)); err != nil {
				errsMux.Lock()

				errs = append(errs, err)

				errsMux.Unlock()

				cancel()
			}

			wg.Done()
		}(i, job)
	}

	wg.Wait()

	return errors.Join(errs...)
}

func omitContextErr(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}

	return err
}
