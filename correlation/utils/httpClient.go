package utils

import (
	"io"
	"net/http"

	"github.com/threatwinds/go-sdk/catcher"
)

func DoPost(url, contentType string, body io.Reader) ([]byte, error) {
	res, err := http.Post(url, contentType, body)
	if err != nil {
		catcher.Error("Could not do request to the URL", err, nil)
		return []byte{}, err
	}

	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		catcher.Error("Could not read response", err, nil)
		return []byte{}, err
	}
	return response, nil
}
