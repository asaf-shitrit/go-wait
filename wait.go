package wait

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	canceledErr            = errors.New("the operation was canceled")
	invalidBackoffLimitErr = errors.New("the provided backoff limit is lower then the baseline")
)

type UntilOptions struct {
	interval time.Duration
	jitter            int
}

func (o *UntilOptions) jitterDefined() bool {
	return o.jitter != -1 && o.jitter != 0
}

var defaultUntilOptions = &UntilOptions{
	interval: time.Millisecond * 100,
	jitter:   0,
}

// Until allows for a predictable interval based waiting mechanism until
// the given bool based check is satisfied. 
func Until(ctx context.Context, check func() (bool, error), o ...*UntilOptions) error {
	options := defaultUntilOptions
	if len(o) != 0 {
		options = o[0]
	}

	calculateNextInterval := func() time.Duration {
		if !options.jitterDefined() {
			return options.interval
		}
		return jitterDuration(options.interval, options.jitter)
	}

	t := time.NewTimer(calculateNextInterval())
	for {
		select {
		case <-t.C:
			res, err := check()
			if err != nil {
				return err
			}
			if !res {
				t.Reset(calculateNextInterval())
				continue
			}
			return nil
		case <-ctx.Done():
			return canceledErr
		}
	}
}

type BackoffOptions struct {
	jitter                  int
	baselineDuration, limit time.Duration
	multiplier              int64
}

func (o * BackoffOptions) jitterDefined() bool {
	return o.jitter != -1 && o.jitter != 0
}

var defaultBackoffOptions = &BackoffOptions{
	baselineDuration: time.Millisecond,
	limit:            500 * time.Millisecond,
	multiplier:       2,
	jitter:           0,
}

// Backoff is a waiting mechanism that allows for better CPU load as the interval
// starts from a given baseline and then backs off until it reaches the provided
// limit.
//
// Note: this is partially bases off of http.Server implementation of their
// Shutdown polling mechanism.
func Backoff(ctx context.Context, check func() (bool, error), o ...*BackoffOptions) error {

	options := defaultBackoffOptions
	if len(o) != 0 {
		options = o[0]
	}

	// make sure limit is greater then the given duration
	if options.limit < options.baselineDuration {
		return invalidBackoffLimitErr
	}

	duration := options.baselineDuration
	t := time.NewTimer(duration)

	calcNewDuration := func(previous time.Duration) time.Duration {
		d := time.Duration(int64(previous) * int64(options.multiplier))
		if !options.jitterDefined() {
			return d
		}
		return jitterDuration(d, options.jitter)
	}

	for {
		select {
		case <-ctx.Done():
			return canceledErr
		case <-t.C:

			res, err := check()
			if err != nil {
				return err
			}

			if res {
				return nil
			}

			if duration < options.limit {
				duration = calcNewDuration(duration)
			} else {
				// we cap the timer duration to the limit
				duration = options.limit
			}
			t.Reset(duration)
		}
	}
}
