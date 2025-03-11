package utils

import (
	"net/http"
)

func ConnectionChecker(url string) error {
	checkConn := func() error {
		if err := checkPanelConnection(url); err != nil {
			return Logger.ErrorF("connection failed: %v", err)
		}

		return nil
	}

	if err := Logger.InfiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return Logger.ErrorF("error checking connection: %v", err)
	}

	return nil
}

func checkPanelConnection(url string) error {
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
