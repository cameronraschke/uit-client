//go:build linux && amd64

package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"uitclient/config"
)

func GetClientConfig() (*config.ClientConfig, error) {
	reqURL := &url.URL{
		Scheme: "http",
		Path:   "/static/client/configs/uit-client",
	}
	queries := url.Values{}
	queries.Set("json", "true")
	reqURL.RawQuery = queries.Encode()

	resp, err := GetData(reqURL)
	if err != nil {
		return nil, fmt.Errorf("error in GetClientConfig: %v", err)
	}

	var configData config.ClientConfig
	if err := json.Unmarshal(resp, &configData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GetClientConfig response: %v", err)
	}
	return &configData, nil
}

func SerialLookup(s string) (*config.ClientLookup, error) {
	if s == "" {
		return nil, fmt.Errorf("serial number is empty in SerialLookup")
	}
	serial := strings.TrimSpace(s)
	reqURL := &url.URL{Path: "/api/lookup"}
	queries := url.Values{}
	queries.Add("system_serial", serial)
	reqURL.RawQuery = queries.Encode()

	resp, err := GetData(reqURL)
	if err != nil {
		return nil, fmt.Errorf("error in SerialLookup: %v", err)
	}
	clientLookup := &config.ClientLookup{}
	if err := json.Unmarshal(resp, clientLookup); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SerialLookup response: %v", err)
	}
	return clientLookup, nil
}

func TagnumberLookup(tagnumber int) (*config.ClientLookup, error) {
	if tagnumber <= 0 {
		return nil, fmt.Errorf("invalid tagnumber in TagnumberLookup")
	}
	reqURL := &url.URL{Path: "/api/lookup"}
	queries := url.Values{}
	queries.Add("tagnumber", fmt.Sprintf("%d", tagnumber))
	reqURL.RawQuery = queries.Encode()

	resp, err := GetData(reqURL)
	if err != nil {
		return nil, fmt.Errorf("error in TagnumberLookup: %v", err)
	}
	clientLookup := &config.ClientLookup{}
	if err := json.Unmarshal(resp, clientLookup); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TagnumberLookup response: %v", err)
	}
	return clientLookup, nil
}
