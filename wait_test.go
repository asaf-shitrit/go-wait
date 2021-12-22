package wait

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func calcCheckTimesDelta(checkTimes []time.Time) []time.Duration {
	checkDurations := []time.Duration{}
	for i := range checkTimes {
		if i == 0 {
			continue
		}
		checkDurations = append(checkDurations, checkTimes[i].Sub(checkTimes[i-1]))
	}
	return checkDurations
}

func TestUntil(t *testing.T) {
	t.Parallel()

	t.Run("wait for 3 seconds", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
		options := &UntilOptions{
			Interval: time.Millisecond * 500,
			Jitter:   10, //10% of jitter in this case
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

		checkIntervals := calcCheckTimesDelta(checkTimes)
		didAnyIntervalJitter := false
		for _, interval := range checkIntervals {
			if interval != options.Interval {
				didAnyIntervalJitter = true
			}
		}

		assert.True(t, didAnyIntervalJitter)
	})

	t.Run("should timeout", func(t *testing.T) {
		t.Parallel()
		options := &UntilOptions{
			Interval: defaultUntilOptions.Interval,
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

func TestBackoff(t *testing.T){
	t.Parallel()

	t.Run("should backoff interval of checks", func (t *testing.T)  {
		t.Parallel()
		checkTimes := []time.Time{}
		counter := 0
		err := Backoff(context.Background(), func() (bool, error) {
			checkTimes = append(checkTimes, time.Now())
			counter++
			if counter == 10 {
				return true, nil
			}
			return false, nil
		})
		assert.Nil(t, err)

		checkDurations := calcCheckTimesDelta(checkTimes)
		for i := range checkDurations {
			if i == 0 {
				continue
			}
			assert.Greater(t, checkDurations[i], checkDurations[i-1])
		}
	})

	t.Run("should be cancelable", func (t *testing.T)  {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())

		go func (){
			<- time.After(time.Second*3)
			cancel()
		}()

		err := Backoff(ctx, func() (bool, error) {
			return false, nil
		})
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, canceledErr)
	})

	t.Run("should be able to timeout", func (t *testing.T)  {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		err := Backoff(ctx, func() (bool, error) {
			return false, nil
		})
		assert.NotNil(t, err)
		assert.ErrorIs(t, err, canceledErr)
	})

	t.Run("should not pass limit", func (t *testing.T)  {
		t.Parallel()
		expectedMaxInterval := time.Millisecond*200

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		checkTimes := []time.Time{}

		err := Backoff(ctx, func () (bool, error) {
			checkTimes = append(checkTimes, time.Now())
			return false, nil
		}, &BackoffOptions{
			BaselineDuration: time.Millisecond*100,
			Limit: expectedMaxInterval,
		})

		assert.NotNil(t, err)
		assert.ErrorIs(t, err, canceledErr)

		intervals := calcCheckTimesDelta(checkTimes)
		for _, d := range intervals {
			assert.LessOrEqual(t, d, expectedMaxInterval)
		}
	})
}

func TestJitterDuration(t *testing.T) {
	t.Parallel()
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
