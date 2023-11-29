package utils

import "time"

func IncrementReconnectDelay(delay time.Duration, maxReconnectDelay time.Duration) time.Duration {
	delay *= 2
	if delay > maxReconnectDelay {
		delay = maxReconnectDelay
	}
	return delay
}
