package webclient

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func GetClientConfig() ([]byte, error) {
	reqURL := &url.URL{}
	reqURL.Scheme = "http"
	reqURL.Path = "/client/api/configs/uit-client"
	queries := url.Values{}
	queries.Set("json", "true")
	reqURL.RawQuery = queries.Encode()

	resp, err := CreateGETRequest(reqURL)
	if err != nil {
		return nil, fmt.Errorf("error in GetClientConfig: %v", err)
	}
	return resp, nil
}

func SerialLookup(serial string) (int64, error) {
	if serial == "" {
		return 0, fmt.Errorf("serial number is empty in SerialLookup")
	}
	reqURL := &url.URL{}
	reqURL.Scheme = "https"
	reqURL.Path = "/api/lookup"
	queries := url.Values{}
	queries.Add("system_serial", serial)
	reqURL.RawQuery = queries.Encode()

	resp, err := CreateGETRequest(reqURL)
	if err != nil {
		return 0, fmt.Errorf("error in SerialLookup: %v", err)
	}
	clientLookup := &ClientLookup{}
	if err := json.Unmarshal(resp, clientLookup); err != nil {
		return 0, fmt.Errorf("failed to unmarshal SerialLookup response: %v", err)
	}
	return clientLookup.Tagnumber, nil
}

func TagnumberLookup(tagnumber int) ([]byte, error) {
	return nil, nil
}
