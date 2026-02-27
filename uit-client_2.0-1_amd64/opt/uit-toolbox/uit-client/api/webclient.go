//go:build linux && amd64

package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const requestRetryCount = 3

type ClientLookup struct {
	Tagnumber    int64  `json:"tagnumber"`
	SystemSerial string `json:"system_serial"`
}

func GetData(reqURL *url.URL) ([]byte, error) {
	if reqURL == nil {
		return nil, fmt.Errorf("input URL is nil")
	}

	if reqURL.Scheme == "" {
		reqURL.Scheme = "https"
	}

	if reqURL.Host == "" {
		reqURL.Host = "10.0.0.1"
	}

	switch reqURL.Scheme {
	case "https":
		reqURL.Host = reqURL.Host + ":31411"
	case "http":
		reqURL.Host = reqURL.Host + ":8080"
	default:
		reqURL.Host = reqURL.Host + ":31411"
	}

	resp, err := http.Get(reqURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %v", err)
	}
	if resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}
