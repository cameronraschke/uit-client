package webclient

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

func SendGETRequest(reqURL *url.URL) (*http.Response, error) {
	if reqURL == nil {
		return nil, fmt.Errorf("input URL is empty")
	}

	resp, err := http.Get(reqURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %v", err)
	}
	defer resp.Body.Close()
	if resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	return resp, nil
}

func CreateGETRequest(reqURL *url.URL) ([]byte, error) {
	if reqURL == nil {
		return nil, fmt.Errorf("input URL is nil")
	}

	if reqURL.Host == "" {
		reqURL.Host = "10.0.0.1"
	}

	switch reqURL.Scheme {
	case "https":
		reqURL.Scheme = "https"
		reqURL.Host = reqURL.Host + ":31411"
	case "http":
		reqURL.Scheme = "http"
		reqURL.Host = reqURL.Host + ":8080"
	default:
		reqURL.Scheme = "https"
		reqURL.Host = reqURL.Host + ":31411"
	}

	queries := reqURL.Query()
	if queries != nil {
		for key, values := range reqURL.Query() {
			for _, value := range values {
				queries.Add(key, value)
			}
		}
		reqURL.RawQuery = queries.Encode()
	}

	var resp *http.Response
	var err error
	for i := range requestRetryCount {
		if i < requestRetryCount-1 {
			fmt.Printf("GET request attempt %d failed, retrying...\n", i+1)
		}
		resp, err = SendGETRequest(reqURL)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to complete GET request after %d attempts: %v", requestRetryCount, err)
	}
	defer resp.Body.Close()
	if resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}
