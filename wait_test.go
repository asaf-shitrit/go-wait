package wait

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUntil(t *testing.T) {
	t.Run("wait for 3 seconds", func(t *testing.T) {
		value := false
		mu := sync.Mutex{}

		go func() {
			<-time.After(time.Second * 3)
			mu.Lock()
			defer mu.Unlock()
			value = true
		}()

		err := Until(context.Background(), func() (bool, error) {
			mu.Lock()
			defer mu.Unlock()
			return value, nil
		})

		assert.Nil(t, err)
	})

	t.Run("jitter should produce non-repeating interval check times", func(t *testing.T) {
		options := &UntilOptions{
			interval: time.Millisecond * 500,
			jitter:   10, //10% of jitter in this case
		}

		value := false
		mu := sync.Mutex{}

		go func() {
			<-time.After(time.Second * 5)
			mu.Lock()
			defer mu.Unlock()
			value = true
		}()

		checkTimes := make([]time.Time, 0)
		returnFunc := func() (bool, error) {
			checkTimes = append(checkTimes, time.Now())
			mu.Lock()
			defer mu.Unlock()
			return value, nil
		}

		err := Until(context.Background(), returnFunc, options)
		assert.Nil(t, err)

		checkIntervals := make([]time.Duration, len(checkTimes)-1)
		for i := range checkTimes {
			if i == 0 {
				continue
			}
			checkIntervals[i-1] = checkTimes[i].Sub(checkTimes[i-1])
		}
		didAnyIntervalJitter := false
		for _, interval := range checkIntervals {
			if interval != options.interval {
				didAnyIntervalJitter = true
			}
		}

		assert.True(t, didAnyIntervalJitter)
	})

	t.Run("should timeout", func(t *testing.T) {
		options := &UntilOptions{
			interval: defaultUntilOptions.interval,
		}
		returnFunc := func() (bool, error) {
			return false, nil
		}

		// we use context to 
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		err := Until(ctx, returnFunc, options)
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, canceledErr)
	})
}

func TestJitterDuration(t *testing.T) {
	d := time.Second*100
	jitter := 50
	offset := time.Duration(int64(d)/int64(jitter))
	expectedMinInterval := d-offset
	expectedMaxInterval := d+offset

	for i:=1;i<1000;i++{
		jd := jitterDuration(d, jitter)
		assert.True(t, expectedMinInterval <= jd && jd <= expectedMaxInterval)
	}
}
