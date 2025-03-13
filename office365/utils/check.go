package utils

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const connectionTimeout = 5 * time.Second

func ConnectionChecker(url string) error {
	checkConn := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()

		if err := checkConnection(url, ctx); err != nil {
			return fmt.Errorf("connection failed")
		}
		return nil
	}

	if err := Logger.InfiniteRetryIfXError(checkConn, "connection failed"); err != nil {
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
	defer resp.Body.Close()

	return nil
}
