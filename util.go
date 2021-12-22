package wait

import (
	"math/rand"
	"time"
)

func jitterDuration(duration time.Duration, jitterPercentage int) time.Duration {
	maxTimeJitter := int64(duration) / int64(jitterPercentage)
	return time.Duration(int64(duration) + rand.Int63n(maxTimeJitter*2) - maxTimeJitter)
}
