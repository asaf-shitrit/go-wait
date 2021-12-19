package wait

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

type Options struct {
	interval time.Duration
	timeout time.Duration
	jitter int
}

var defaultOptions = &Options{
	interval: time.Millisecond*100,
	timeout: time.Second*10,
	jitter: 0,
}

func init(){
	rand.Seed(time.Now().UnixNano())
}

var ctxErr = errors.New("the operation was either canceled or had a timeout")

func Until(ctx context.Context, check func () (bool, error), o ...*Options) error {
	options := defaultOptions
	if len(o) != 0 {
		options = o[0]
	}

	tCtx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()

	// we pre-calculate the jitter offset to reduce
	// overhead in each run of calculateNextInterval
	var maxTimeJitter int64
	if options.jitter != 0 {
		maxTimeJitter = int64(int64(options.interval)/int64(options.jitter))
	}

	calculateNextInterval := func () time.Duration  {
		if options.jitter == 0 {
			return options.interval
		}
		// we want to jitter to be in the range of 
		// [interval - jitter] ~ [interval + jitter]
		timeJitter := time.Duration(rand.Int63n(maxTimeJitter*2)-maxTimeJitter)
		return options.interval + timeJitter
	}

	t := time.NewTimer(calculateNextInterval())
	for {
		select {
		case <- t.C:
			res, err := check()
			if err != nil {
				return err
			}
			if !res {
				t.Reset(calculateNextInterval())
				continue
			}
			return nil
		case <- tCtx.Done():
			return ctxErr
		}
	}
}