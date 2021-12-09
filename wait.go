package wait

import (
	"context"
	"errors"
	"time"
)

type Options struct {
	interval time.Duration
	timeout time.Duration
}

var defaultOptions = &Options{
	interval: time.Millisecond*100,
	timeout: time.Second*10,
}

var ctxErr = errors.New("the operation was either canceled or had a timeout")

func Until(ctx context.Context, check func () (bool, error), o ...*Options) error {
	options := defaultOptions
	if len(o) != 0 {
		options = o[0]
	}

	tCtx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()
	t := time.NewTicker(options.interval)
	for {
		select {
		case <- t.C:
			res, err := check()
			if err != nil {
				return err
			}
			if !res {
				continue
			}
			return nil
		case <- tCtx.Done():
			return ctxErr
		}
	}
}