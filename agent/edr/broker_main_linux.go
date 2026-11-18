//go:build linux

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/utmstack/UTMStack/agent/edr/fanotify"
)

// Dispatch before service/log/database startup. The broker never launches
// helpers, scans content, opens service state or owns the parent's Go runtime.
func runInternalBroker() bool {
	if len(os.Args) < 2 || os.Args[1] != "__fanotify-broker" {
		return false
	}
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "invalid fanotify broker invocation")
		os.Exit(2)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	f := os.NewFile(3, "fanotify-owner-channel")
	defer f.Close()
	if err := fanotify.ServeBroker(ctx, f); err != nil && ctx.Err() == nil && !errors.Is(err, io.EOF) && !errors.Is(err, syscall.ECONNRESET) && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return true
}
