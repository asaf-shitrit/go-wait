package wait

import (
	"context"
	"sync"
	"testing"
	"time"
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