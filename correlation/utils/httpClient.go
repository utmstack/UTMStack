package utils

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func Download(url, file string) error {
	out, err := os.Create(file)
	if err != nil {
		h.Error("Could not create file: %v", err)
		return err
	}

	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		h.Error("Could not do request to the URL: %v", err)
		return err
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		h.Error("Could not save data to file: %v", err)
		return err
	}
	h.Debug("Downloaded %d bytes from %s", n, url)
	return nil
}

func DoPost(url, contentType string, body io.Reader) ([]byte, error) {
	res, err := http.Post(url, contentType, body)
	if err != nil {
		h.Error("Could not do request to the URL: %v", err)
		return []byte{}, err
	}

	defer res.Body.Close()

	response, err := ioutil.ReadAll(res.Body)
	if err != nil {
		h.Error("Could not read response: %v", err)
		return []byte{}, err
	}
	return response, nil
}
