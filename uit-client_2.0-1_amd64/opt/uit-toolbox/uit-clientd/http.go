//go:build linux && amd64

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var sharedHTTPClient = newHTTPClient()

func newHTTPClient() *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
	}

	tr := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     5 * time.Minute,
		DisableCompression:  true,
		TLSClientConfig:     tlsConfig,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	protocols := http.Protocols{}
	protocols.SetHTTP1(false)
	protocols.SetUnencryptedHTTP2(false)
	protocols.SetHTTP2(true)
	tr.Protocols = &protocols

	return &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func sendHTTPRequest(data *HTTPRequest) ([]byte, error) {
	if data == nil || data.Config == nil {
		return nil, fmt.Errorf("data variable and/or config is nil")
	}
	if data.Config.Method != "POST" && data.Config.Method != "GET" {
		return nil, fmt.Errorf("unsupported HTTP method: %s", data.Config.Method)
	}
	if strings.TrimSpace(data.Config.URL.Path) == "" {
		return nil, fmt.Errorf("relative URL cannot be empty")
	}

	if data.Config.URL.Path == "" {
		return nil, fmt.Errorf("URL path cannot be empty")
	}

	requestURL := &url.URL{
		Scheme:   "https",
		Path:     data.Config.URL.Path,
		RawQuery: data.Config.URL.RawQuery,
	}

	if data.Config.URL.Host != "" {
		requestURL.Host = data.Config.URL.Host
	} else {
		tmpConfig := clientConfig.Load()
		if tmpConfig == nil {
			return nil, fmt.Errorf("client config is not loaded, cannot send request")
		}
		if strings.TrimSpace(tmpConfig.UIT_WEB_HTTPS_HOST) == "" || strings.TrimSpace(tmpConfig.UIT_WEB_HTTPS_PORT) == "" {
			return nil, fmt.Errorf("client config has invalid host or port for HTTPS")
		}
		requestURL.Host = fmt.Sprintf("%s:%s", tmpConfig.UIT_WEB_HTTPS_HOST, tmpConfig.UIT_WEB_HTTPS_PORT)
	}

	// HTTP body
	var bodyReader io.Reader = http.NoBody
	if data.Config.Method == "POST" {
		if data.Payload == nil {
			return nil, fmt.Errorf("payload cannot be nil")
		}
		if data.Payload.RequestType == "POST" && data.Payload.Value == nil {
			return nil, fmt.Errorf("payload value cannot be nil")
		}
		jsonData, err := json.Marshal(data.Payload.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	// HTTP request
	req, err := http.NewRequest(data.Config.Method, requestURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// HTTP headers
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "UIT-Client-CLI Daemon")

	// Server response
	resp, err := sharedHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned an HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func MapInputToPOSTRequest(input string) (*HTTPRequest, error) {
	// This will only create POST requests
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("input cannot be empty or whitespace")
	}

	inputPayload := new(HTTPRequestPayload)
	if err := json.Unmarshal([]byte(input), inputPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input into HTTPRequestPayload: %w", err)
	}

	// Tag number
	if inputPayload.Tagnumber <= 0 || inputPayload.Tagnumber > 999999 {
		return nil, fmt.Errorf("invalid tag number: %d", inputPayload.Tagnumber)
	}
	// Key
	if strings.TrimSpace(inputPayload.Key) == "" {
		return nil, fmt.Errorf("key is empty")
	}
	// Value
	if strings.TrimSpace(inputPayload.StringValue) == "" {
		return nil, fmt.Errorf("value is empty")
	}
	// UUID is optional, but if provided it cannot be empty
	if inputPayload.UUID != nil && strings.TrimSpace(*inputPayload.UUID) == "" {
		return nil, fmt.Errorf("UUID is empty")
	}

	httpRequestConfig := new(HTTPRequestConfig)

	httpRequestConfig.Method = inputPayload.RequestType

	switch inputPayload.Key {
	case "battery_charge_pcnt":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware/battery"}
		batteryPcnt, err := strconv.ParseFloat(inputPayload.StringValue, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing battery_charge_pcnt value: %w", err)
		}
		if batteryPcnt < 0 || batteryPcnt > 110 {
			return nil, fmt.Errorf("battery_charge_pcnt value out of range: %f", batteryPcnt)
		}
		inputPayload.Value = &BatteryData{
			Tagnumber: inputPayload.Tagnumber,
			Percent:   &batteryPcnt,
		}
	case "system_uptime":
		httpRequestConfig.URL = url.URL{Path: "/api/client/uptime"}
		uptimeSeconds, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse system_uptime value: %w", err)
		}
		if uptimeSeconds < 0 {
			return nil, fmt.Errorf("system_uptime value cannot be negative: %d", uptimeSeconds)
		}
		inputPayload.Value = &ClientUptime{
			Tagnumber:    inputPayload.Tagnumber,
			SystemUptime: &uptimeSeconds,
		}
	case "client_app_uptime":
		httpRequestConfig.URL = url.URL{Path: "/api/client/uptime"}
		uptimeSeconds, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse client_app_uptime value: %w", err)
		}
		if uptimeSeconds < 0 {
			return nil, fmt.Errorf("client_app_uptime value cannot be negative: %d", uptimeSeconds)
		}
		inputPayload.Value = &ClientUptime{
			Tagnumber:       inputPayload.Tagnumber,
			ClientAppUptime: &uptimeSeconds,
		}
	case "cpu_current_usage":
		httpRequestConfig.URL = url.URL{Path: "/api/client/cpu/usage"}
		cpuUsage, err := strconv.ParseFloat(inputPayload.StringValue, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu_current_usage value: %w", err)
		}
		if cpuUsage < 0 || cpuUsage > 110 {
			return nil, fmt.Errorf("cpu_current_usage value out of range: %f", cpuUsage)
		}
		inputPayload.Value = &CPUDataRequest{
			Tagnumber:    &inputPayload.Tagnumber,
			UsagePercent: &cpuUsage,
		}
	case "cpu_current_mhz":
		httpRequestConfig.URL = url.URL{Path: "/api/client/cpu/mhz"}
		cpuCurrentMHz, err := strconv.ParseFloat(inputPayload.StringValue, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu_current_mhz value: %w", err)
		}
		if cpuCurrentMHz <= 0 {
			return nil, fmt.Errorf("cpu_current_mhz value must be greater than 0: %f", cpuCurrentMHz)
		}
		inputPayload.Value = &CPUDataRequest{
			Tagnumber: &inputPayload.Tagnumber,
			MHz:       &cpuCurrentMHz,
		}
	case "cpu_millidegrees_c":
		httpRequestConfig.URL = url.URL{Path: "/api/client/cpu/temp"}
		cpuTempMilliC, err := strconv.ParseFloat(inputPayload.StringValue, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu_millidegrees_c value: %w", err)
		}
		if cpuTempMilliC < 0 {
			return nil, fmt.Errorf("cpu_millidegrees_c value cannot be negative: %f", cpuTempMilliC)
		}
		inputPayload.Value = &CPUDataRequest{
			Tagnumber:     &inputPayload.Tagnumber,
			MillidegreesC: &cpuTempMilliC,
		}
	case "memory_usage_kb":
		httpRequestConfig.URL = url.URL{Path: "/api/client/memory/usage"}
		memoryUsageKB, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse memory_usage_kb value: %w", err)
		}
		if memoryUsageKB <= 0 {
			return nil, fmt.Errorf("memory_usage_kb has to be greater than 0: %d", memoryUsageKB)
		}
		inputPayload.Value = &MemoryDataRequest{
			Tagnumber:    &inputPayload.Tagnumber,
			TotalUsageKB: &memoryUsageKB,
		}
	case "memory_capacity_kb":
		httpRequestConfig.URL = url.URL{Path: "/api/client/memory/capacity"}
		memoryCapacityKB, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse memory_capacity_kb value: %w", err)
		}
		if memoryCapacityKB <= 0 {
			return nil, fmt.Errorf("memory_capacity_kb has to be greater than 0: %d", memoryCapacityKB)
		}
		inputPayload.Value = &MemoryDataRequest{
			Tagnumber:       &inputPayload.Tagnumber,
			TotalCapacityKB: &memoryCapacityKB,
		}
	case "client_lookup_by_serial":
		httpRequestConfig.URL = url.URL{Path: "/api/client/lookup"}
		query := httpRequestConfig.URL.Query()
		query.Set("serial", inputPayload.StringValue)
		httpRequestConfig.URL.RawQuery = query.Encode()
	default:
		return nil, fmt.Errorf("unsupported key: '%s'", inputPayload.Key)
	}

	return &HTTPRequest{
		Config:  httpRequestConfig,
		Payload: inputPayload,
	}, nil
}
