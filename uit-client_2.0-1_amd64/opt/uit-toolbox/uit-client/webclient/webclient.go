package webclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ClientLookup struct {
	Tagnumber    int64  `json:"tagnumber"`
	SystemSerial string `json:"system_serial"`
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

	resp, err := http.Get(reqURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	//log.Printf("Received response: %s", string(body))
	return body, nil
}
