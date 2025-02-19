package utils

import (
	"io"
	"log"
	"net/http"
)

func DoPost(url, contentType string, body io.Reader) ([]byte, error) {
	res, err := http.Post(url, contentType, body)
	if err != nil {
		log.Printf("Could not do request to the URL: %v", err)
		return []byte{}, err
	}

	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Could not read response: %v", err)
		return []byte{}, err
	}
	return response, nil
}
