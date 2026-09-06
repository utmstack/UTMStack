package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const connectivityRetryDelay = 5 * time.Second

func waitForConnectivity(ctx context.Context, checker func(string) error, url string, retryDelay time.Duration) {
	for {
		if err := checker(url); err != nil {
			_ = catcher.Error("failed to connect with external service", err, map[string]any{"process": processName})
			if !sleepWithCancel(ctx, retryDelay) {
				return
			}
			continue
		}
		return
	}
}
