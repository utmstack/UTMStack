package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const (
	connectionTimeout = 5 * time.Second
	wait              = 3 * time.Second
)

const CHECKCON = "https://id.sophos.com"

func ConnectionChecker(url string) error {
	checkConn := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()

		if err := checkConnection(url, ctx); err != nil {
			return fmt.Errorf("connection failed")
		}
		return nil
	}

	if err := infiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return err
	}

	return nil
}

func checkConnection(url string, ctx context.Context) error {
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			_ = catcher.Error("error closing response body: %v", err, nil)
		}
	}()

	return nil
}

func infiniteRetryIfXError(f func() error, exception string) error {
	var xErrorWasLogged bool

	for {
		err := f()
		if err != nil && is(err, exception) {
			if !xErrorWasLogged {
				_ = catcher.Error("An error occurred (%s), will keep retrying indefinitely...", err, nil)
				xErrorWasLogged = true
			}
			time.Sleep(wait)
			continue
		}

		return err
	}
}

func is(e error, args ...string) bool {
	for _, arg := range args {
		if strings.Contains(e.Error(), arg) {
			return true
		}
	}
	return false
}
