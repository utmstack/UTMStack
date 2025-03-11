package utils

import (
	"net/http"

	"github.com/threatwinds/logger"
)

func ConnectionChecker(url string, h *logger.Logger) error {
	checkConn := func() error {
		if err := CheckConnection(url); err != nil {
			return h.ErrorF("connection failed: %v", err)
		}
		return nil
	}

	if err := h.InfiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return h.ErrorF("error checking connection: %v", err)
	}

	return nil
}

func CheckConnection(url string) error {
	client := &http.Client{}

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
