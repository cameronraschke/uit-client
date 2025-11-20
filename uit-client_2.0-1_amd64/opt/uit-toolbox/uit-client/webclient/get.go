package webclient

import (
	"fmt"
	"net/url"
)

func GetClientConfig() ([]byte, error) {
	reqURL := &url.URL{}
	reqURL.Scheme = "http"
	reqURL.Path = "/client/api/configs/uit-client"
	queries := url.Values{}
	queries.Add("json", "true")

	resp, err := CreateGETRequest(reqURL)
	if err != nil {
		return nil, fmt.Errorf("error in GetClientConfig: %v", err)
	}

	return resp, nil
}

func SerialLookup(serial string) ([]byte, error) {
	if serial == "" {
		return nil, fmt.Errorf("serial number is empty in SerialLookup")
	}
	reqURL := &url.URL{}
	reqURL.Scheme = "https"
	reqURL.Path = "/api/lookup"
	queries := url.Values{}
	queries.Add("system_serial", serial)
	reqURL.RawQuery = queries.Encode()

	resp, err := CreateGETRequest(reqURL)
	if err != nil {
		return nil, fmt.Errorf("error in SerialLookup: %v", err)
	}
	return resp, nil
}
