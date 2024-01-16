package utils

import "time"

// IncrementReconnectDelay increments the delay for reconnecting to the server
func IncrementReconnectDelay(delay time.Duration, maxReconnectDelay time.Duration) time.Duration {
	delay *= 2
	if delay > maxReconnectDelay {
		delay = maxReconnectDelay
	}
	return delay
}
