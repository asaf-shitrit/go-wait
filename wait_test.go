package wait

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUntil(t *testing.T){
	t.Run("wait for 3 seconds", func (t *testing.T){
		value := false
		mu := sync.Mutex{}

		go func ()  {
			<- time.After(time.Second*3)
			mu.Lock()
			defer mu.Unlock()
			value = true
		}()

		err := Until(context.Background(), func() (bool, error) {
			mu.Lock()
			defer mu.Unlock()
			return value, nil
		})

		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("jitter should produce non-repeating interval check times", func (t *testing.T)  {
		options := &Options{
			timeout: time.Minute,
			interval: time.Millisecond*500,
			jitter: 10, //10% of jitter in this case
		}

		value := false
		mu := sync.Mutex{}

		go func ()  {
			<- time.After(time.Second*5)
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

		expectedJitterOffset := time.Duration(int64(options.interval)/int64(options.jitter))

		expectedMinInterval := options.interval - expectedJitterOffset
		expectedMaxInterval := options.interval + expectedJitterOffset
		didAnyIntervalJitter := false

		for _, interval := range checkIntervals {
			assert.GreaterOrEqual(t, interval, expectedMinInterval)
			assert.LessOrEqual(t, interval, expectedMaxInterval)
		
			if interval != options.interval {
				didAnyIntervalJitter = true
			}
		}

		assert.True(t, didAnyIntervalJitter)
	})

	t.Run("should timeout", func(t *testing.T) {
		options := &Options{
			timeout: time.Second*2,
			interval: defaultOptions.interval,
		}
		returnFunc := func() (bool, error) {
			return false, nil
		}

		err := Until(context.Background(), returnFunc, options);
		if err == nil {
			t.Fail()
			return
		}

		if err != ctxErr {
			t.Errorf("not expected timeout err: %v", err)
			return
		}
	})
}