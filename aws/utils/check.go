package utils

import (
	"fmt"
	"net/http"
	"time"
)

func ConnectionChecker(url string) error {
	checkConn := func() error {
		if err := checkPanelConnection(url); err != nil {
			return fmt.Errorf("connection failed: %v", err)
		}

		return nil
	}

	if err := Logger.InfiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return err
	}

	return nil
}

func checkPanelConnection(url string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
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
