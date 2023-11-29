package utils

import "time"

func IncrementReconnectDelay(delay time.Duration, maxReconnectDelay time.Duration) time.Duration {
	delay *= 2
	if delay > maxReconnectDelay {
		delay = maxReconnectDelay
	}
	return delay
}

func IncrementReconnectTime(currentTime time.Duration, delay time.Duration, maxReconnectTime time.Duration) time.Duration {
	currentTime = currentTime + delay
	if currentTime >= maxReconnectTime {
		currentTime = maxReconnectTime
	}
	return currentTime
}
