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
		if data.Payload.Value == nil {
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
	req.Header.Set("User-Agent", "UIT-Client-CLI")

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

	inputArr := strings.Split(input, "|")
	if len(inputArr) < 3 {
		return nil, fmt.Errorf("input must have at least 3 parts separated by '|', got %d", len(inputArr))
	}
	if inputArr[0] == "" {
		return nil, fmt.Errorf("input missing tag number")
	}
	if inputArr[1] == "" {
		return nil, fmt.Errorf("input missing key")
	}
	if inputArr[2] == "" {
		return nil, fmt.Errorf("input missing value")
	}
	if len(inputArr) == 4 && inputArr[3] == "" {
		return nil, fmt.Errorf("input has empty UUID")
	}

	httpRequestConfig := new(HTTPRequestConfig)
	httpRequestPayload := new(HTTPRequestPayload)

	var err error
	httpRequestPayload.Tagnumber, err = strconv.ParseInt(inputArr[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TagNum: %w", err)
	}
	httpRequestPayload.Key = inputArr[1]
	if len(inputArr) == 4 && strings.TrimSpace(inputArr[3]) != "" {
		httpRequestPayload.UUID = &inputArr[3]
	}

	httpRequestConfig.Method = "POST"

	switch httpRequestPayload.Key {
	case "battery_charge_pcnt":
		httpRequestPayload.Key = "battery_charge_pcnt"
		batteryPcnt, err := strconv.ParseFloat(inputArr[2], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing battery_charge_pcnt value: %w", err)
		}
		if batteryPcnt < 0 || batteryPcnt > 110 {
			return nil, fmt.Errorf("battery_charge_pcnt value out of range: %f", batteryPcnt)
		}
		httpRequestPayload.Value = &BatteryData{
			Tagnumber: httpRequestPayload.Tagnumber,
			Percent:   &batteryPcnt,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware/battery"}
	case "system_uptime":
		httpRequestPayload.Key = "system_uptime"
		uptimeSeconds, err := strconv.ParseInt(inputArr[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse system_uptime value: %w", err)
		}
		if uptimeSeconds < 0 {
			return nil, fmt.Errorf("system_uptime value cannot be negative: %d", uptimeSeconds)
		}
		httpRequestPayload.Value = &ClientUptime{
			Tagnumber:    httpRequestPayload.Tagnumber,
			SystemUptime: &uptimeSeconds,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/uptime"}
		httpRequestConfig.Method = "POST"
	case "client_app_uptime":
		httpRequestPayload.Key = "client_app_uptime"
		uptimeSeconds, err := strconv.ParseInt(inputArr[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse client_app_uptime value: %w", err)
		}
		if uptimeSeconds < 0 {
			return nil, fmt.Errorf("client_app_uptime value cannot be negative: %d", uptimeSeconds)
		}
		httpRequestPayload.Value = &ClientUptime{
			Tagnumber:       httpRequestPayload.Tagnumber,
			ClientAppUptime: &uptimeSeconds,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/uptime"}
		httpRequestConfig.Method = "POST"
	case "cpu_current_usage":
		httpRequestPayload.Key = "cpu_current_usage"
		cpuUsage, err := strconv.ParseFloat(inputArr[2], 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu_current_usage value: %w", err)
		}
		if cpuUsage < 0 || cpuUsage > 110 {
			return nil, fmt.Errorf("cpu_current_usage value out of range: %f", cpuUsage)
		}
		httpRequestPayload.Value = &CPUDataRequest{
			Tagnumber:    &httpRequestPayload.Tagnumber,
			UsagePercent: &cpuUsage,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/cpu/usage"}
		httpRequestConfig.Method = "POST"
	case "cpu_current_mhz":
		httpRequestPayload.Key = "cpu_current_mhz"
		cpuUsageMHz, err := strconv.ParseFloat(inputArr[2], 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu_current_mhz value: %w", err)
		}
		if cpuUsageMHz < 0 || cpuUsageMHz > 110 {
			return nil, fmt.Errorf("cpu_current_mhz value out of range: %f", cpuUsageMHz)
		}
		httpRequestPayload.Value = &CPUDataRequest{
			Tagnumber: &httpRequestPayload.Tagnumber,
			MHz:       &cpuUsageMHz,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/cpu/mhz"}
		httpRequestConfig.Method = "POST"
	case "cpu_millidegrees_c":
		httpRequestPayload.Key = "cpu_millidegrees_c"
		cpuTempMilliC, err := strconv.ParseFloat(inputArr[2], 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu_millidegrees_c value: %w", err)
		}
		if cpuTempMilliC < 0 {
			return nil, fmt.Errorf("cpu_millidegrees_c value cannot be negative: %f", cpuTempMilliC)
		}
		httpRequestPayload.Value = &CPUDataRequest{
			Tagnumber:     &httpRequestPayload.Tagnumber,
			MillidegreesC: &cpuTempMilliC,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/cpu/temp"}
		httpRequestConfig.Method = "POST"
	case "memory_usage_kb":
		httpRequestPayload.Key = "memory_usage_kb"
		memoryUsageKB, err := strconv.ParseInt(inputArr[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse memory_usage_kb value: %w", err)
		}
		if memoryUsageKB <= 0 {
			return nil, fmt.Errorf("memory_usage_kb has to be greater than 0: %d", memoryUsageKB)
		}
		httpRequestPayload.Value = &MemoryDataRequest{
			Tagnumber:    &httpRequestPayload.Tagnumber,
			TotalUsageKB: &memoryUsageKB,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/memory/usage"}
		httpRequestConfig.Method = "POST"
	case "memory_capacity_kb":
		httpRequestPayload.Key = "memory_capacity_kb"
		memoryCapacityKB, err := strconv.ParseInt(inputArr[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse memory_capacity_kb value: %w", err)
		}
		if memoryCapacityKB <= 0 {
			return nil, fmt.Errorf("memory_capacity_kb has to be greater than 0: %d", memoryCapacityKB)
		}
		httpRequestPayload.Value = &MemoryDataRequest{
			Tagnumber:       &httpRequestPayload.Tagnumber,
			TotalCapacityKB: &memoryCapacityKB,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/memory/capacity"}
		httpRequestConfig.Method = "POST"
	default:
		return nil, fmt.Errorf("unsupported key: '%s'", httpRequestPayload.Key)
	}

	return &HTTPRequest{
		Config:  httpRequestConfig,
		Payload: httpRequestPayload,
	}, nil
}
