package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/threatwinds/logger"
)

func ConnectionChecker(url string, h *logger.Logger) error {
	checkConn := func() error {
		if err := CheckConnection(url); err != nil {
			return fmt.Errorf("connection failed: %v", err)
		}
		return nil
	}

	if err := h.InfiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return err
	}

	return nil
}

func CheckConnection(url string) error {
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
