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

	"uit-clientd/keypolicy"
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

	q := requestURL.Query()
	q.Set("key", data.Payload.Key)
	q.Set("system_serial", data.Payload.SystemSerial)
	requestURL.RawQuery = q.Encode()

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

func MapInputToHTTPRequest(input string) (*HTTPRequest, error) {
	// Input arrives as a single JSON line over the unix socket.
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("input cannot be empty or whitespace")
	}

	inputPayload := new(HTTPRequestPayload)
	if err := json.Unmarshal([]byte(input), inputPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input into HTTPRequestPayload: %w", err)
	}

	// Key
	if strings.TrimSpace(inputPayload.Key) == "" {
		return nil, fmt.Errorf("key is empty")
	}

	method := strings.ToUpper(strings.TrimSpace(inputPayload.RequestType))
	if method == "" {
		method = "POST"
	}
	if method != "POST" && method != "GET" {
		return nil, fmt.Errorf("unsupported request_type: %s", inputPayload.RequestType)
	}
	rule, ok := keypolicy.Lookup(inputPayload.Key)
	if !ok {
		return nil, fmt.Errorf("unsupported key: '%s'", inputPayload.Key)
	}

	if rule.Method != "" && method != rule.Method {
		return nil, fmt.Errorf("key '%s' requires %s method", inputPayload.Key, rule.Method)
	}
	if strings.TrimSpace(inputPayload.SystemSerial) == "" {
		return nil, fmt.Errorf("system_serial is required")
	}

	if inputPayload.Tagnumber != 0 && (inputPayload.Tagnumber < 1 || inputPayload.Tagnumber > 999999) {
		return nil, fmt.Errorf("invalid tag number: %d", inputPayload.Tagnumber)
	}
	var tagnumber *int64
	if inputPayload.Tagnumber > 0 {
		tagnumber = &inputPayload.Tagnumber
	}
	systemSerial := &inputPayload.SystemSerial

	// Value
	if inputPayload.Key != "init" && inputPayload.Key != "client_lookup_by_serial" && strings.TrimSpace(inputPayload.StringValue) == "" {
		return nil, fmt.Errorf("value is empty")
	}
	if inputPayload.Key == "client_lookup_by_serial" && strings.TrimSpace(inputPayload.StringValue) == "" {
		inputPayload.StringValue = inputPayload.SystemSerial
	}
	// UUID is optional, but if provided it cannot be empty
	if inputPayload.TransactionUUID != nil && strings.TrimSpace(*inputPayload.TransactionUUID) == "" {
		return nil, fmt.Errorf("UUID is empty")
	}
	if rule.RequiresUUID && (inputPayload.TransactionUUID == nil || strings.TrimSpace(*inputPayload.TransactionUUID) == "") {
		return nil, fmt.Errorf("UUID is required for key '%s'", inputPayload.Key)
	}

	httpRequestConfig := new(HTTPRequestConfig)

	inputPayload.RequestType = method
	httpRequestConfig.Method = method

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
			Tagnumber:    tagnumber,
			SystemSerial: systemSerial,
			Percent:      &batteryPcnt,
		}
	case "battery_charge_cycles":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		batteryChargeCycles, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing battery_charge_cycles value: %w", err)
		}
		if batteryChargeCycles < 0 {
			return nil, fmt.Errorf("battery_charge_cycles value cannot be negative: %d", batteryChargeCycles)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:           tagnumber,
			SystemSerial:        systemSerial,
			TransactionUUID:     *inputPayload.TransactionUUID,
			BatteryChargeCycles: &batteryChargeCycles,
		}
	case "battery_current_max_capacity":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		batteryMaxCapacity, err := strconv.ParseFloat(inputPayload.StringValue, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing battery_current_max_capacity value: %w", err)
		}
		if batteryMaxCapacity < 0 {
			return nil, fmt.Errorf("battery_current_max_capacity value cannot be negative: %f", batteryMaxCapacity)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:                 tagnumber,
			SystemSerial:              systemSerial,
			TransactionUUID:           *inputPayload.TransactionUUID,
			BatteryCurrentMaxCapacity: &batteryMaxCapacity,
		}
	case "battery_design_capacity":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		batteryDesignCapacity, err := strconv.ParseFloat(inputPayload.StringValue, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing battery_design_capacity value: %w", err)
		}
		if batteryDesignCapacity < 0 {
			return nil, fmt.Errorf("battery_design_capacity value cannot be negative: %f", batteryDesignCapacity)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:             tagnumber,
			SystemSerial:          systemSerial,
			TransactionUUID:       *inputPayload.TransactionUUID,
			BatteryDesignCapacity: &batteryDesignCapacity,
		}
	case "battery_manufacture_date":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:              tagnumber,
			SystemSerial:           systemSerial,
			TransactionUUID:        *inputPayload.TransactionUUID,
			BatteryManufactureDate: &inputPayload.StringValue,
		}
	case "battery_manufacturer":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:           tagnumber,
			SystemSerial:        systemSerial,
			TransactionUUID:     *inputPayload.TransactionUUID,
			BatteryManufacturer: &inputPayload.StringValue,
		}
	case "battery_model":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			BatteryModel:    &inputPayload.StringValue,
		}
	case "battery_serial":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			BatterySerial:   &inputPayload.StringValue,
		}
	case "bios_firmware":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			BiosFirmware:    &inputPayload.StringValue,
		}
	case "bios_release_date":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			BiosReleaseDate: &inputPayload.StringValue,
		}
	case "bios_version":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			BiosVersion:     &inputPayload.StringValue,
		}

	case "chassis_type":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			ChassisType:     &inputPayload.StringValue,
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
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			ClientAppUptime: &uptimeSeconds,
		}
	case "client_lookup_by_serial":
		httpRequestConfig.URL = url.URL{Path: "/api/client/lookup"}
		query := httpRequestConfig.URL.Query()
		query.Set("system_serial", inputPayload.StringValue)
		httpRequestConfig.URL.RawQuery = query.Encode()
	case "cpu_core_count":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		cpuCoreCount, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu_core_count value: %w", err)
		}
		if cpuCoreCount <= 0 {
			return nil, fmt.Errorf("cpu_core_count value must be greater than 0: %d", cpuCoreCount)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			CPUCoreCount:    &cpuCoreCount,
		}
	case "cpu_thread_count":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		cpuThreadCount, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu_thread_count value: %w", err)
		}
		if cpuThreadCount <= 0 {
			return nil, fmt.Errorf("cpu_thread_count value must be greater than 0: %d", cpuThreadCount)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			CPUThreadCount:  &cpuThreadCount,
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
			Tagnumber:    tagnumber,
			SystemSerial: systemSerial,
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
			Tagnumber:    tagnumber,
			SystemSerial: systemSerial,
			MHz:          &cpuCurrentMHz,
		}
	case "cpu_manufacturer":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			CPUManufacturer: &inputPayload.StringValue,
		}
	case "cpu_max_speed_mhz":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		cpuMaxSpeedMHz, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse cpu_max_speed_mhz value: %w", err)
		}
		if cpuMaxSpeedMHz <= 0 {
			return nil, fmt.Errorf("cpu_max_speed_mhz value must be greater than 0: %d", cpuMaxSpeedMHz)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			CPUMaxSpeedMhz:  &cpuMaxSpeedMHz,
		}
	case "cpu_model":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			CPUModel:        &inputPayload.StringValue,
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
			Tagnumber:     tagnumber,
			SystemSerial:  systemSerial,
			MillidegreesC: &cpuTempMilliC,
		}
	case "disk_errors":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		diskErrors, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse disk_errors value: %w", err)
		}
		if diskErrors < 0 {
			return nil, fmt.Errorf("disk_errors value cannot be negative: %d", diskErrors)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			DiskErrors:      &diskErrors,
		}
	case "disk_firmware":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			DiskFirmware:    &inputPayload.StringValue,
		}
	case "disk_model":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			DiskModel:       &inputPayload.StringValue,
		}
	case "disk_power_cycles":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		diskPowerCycles, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse disk_power_cycles value: %w", err)
		}
		if diskPowerCycles < 0 {
			return nil, fmt.Errorf("disk_power_cycles value cannot be negative: %d", diskPowerCycles)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			DiskPowerCycles: &diskPowerCycles,
		}
	case "disk_power_on_hours":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		diskPowerOnHours, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse disk_power_on_hours value: %w", err)
		}
		if diskPowerOnHours < 0 {
			return nil, fmt.Errorf("disk_power_on_hours value cannot be negative: %d", diskPowerOnHours)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:        tagnumber,
			SystemSerial:     systemSerial,
			TransactionUUID:  *inputPayload.TransactionUUID,
			DiskPowerOnHours: &diskPowerOnHours,
		}
	case "disk_reads_kb":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		diskReadsKB, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse disk_reads_kb value: %w", err)
		}
		if diskReadsKB < 0 {
			return nil, fmt.Errorf("disk_reads_kb value cannot be negative: %d", diskReadsKB)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			DiskReadsKB:     &diskReadsKB,
		}
	case "disk_serial":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			DiskSerial:      &inputPayload.StringValue,
		}
	case "disk_size_kb":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		diskSizeKB, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse disk_size_kb value: %w", err)
		}
		if diskSizeKB <= 0 {
			return nil, fmt.Errorf("disk_size_kb value must be greater than 0: %d", diskSizeKB)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			DiskSize:        &diskSizeKB,
		}
	case "disk_type":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			DiskType:        &inputPayload.StringValue,
		}
	case "disk_writes_kb":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		diskWritesKB, err := strconv.ParseInt(inputPayload.StringValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse disk_writes_kb value: %w", err)
		}
		if diskWritesKB < 0 {
			return nil, fmt.Errorf("disk_writes_kb value cannot be negative: %d", diskWritesKB)
		}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			DiskWritesKB:    &diskWritesKB,
		}
	case "ethernet_mac":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			EthernetMAC:     &inputPayload.StringValue,
		}
	case "init":
		httpRequestConfig.URL = url.URL{Path: "/api/client/init"}
		inputPayload.Value = &ClientInitRequest{
			Tagnumber:       tagnumber,
			SystemSerial:    &inputPayload.SystemSerial,
			TransactionUUID: inputPayload.TransactionUUID,
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
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:        tagnumber,
			SystemSerial:     systemSerial,
			TransactionUUID:  *inputPayload.TransactionUUID,
			MemoryCapacityKB: &memoryCapacityKB,
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
			Tagnumber:    tagnumber,
			SystemSerial: systemSerial,
			TotalUsageKB: &memoryUsageKB,
		}
	case "motherboard_manufacturer":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:               tagnumber,
			SystemSerial:            systemSerial,
			TransactionUUID:         *inputPayload.TransactionUUID,
			MotherboardManufacturer: &inputPayload.StringValue,
		}
	case "motherboard_serial":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:         tagnumber,
			SystemSerial:      systemSerial,
			TransactionUUID:   *inputPayload.TransactionUUID,
			MotherboardSerial: &inputPayload.StringValue,
		}

	case "system_manufacturer":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:          tagnumber,
			SystemSerial:       systemSerial,
			TransactionUUID:    *inputPayload.TransactionUUID,
			SystemManufacturer: &inputPayload.StringValue,
		}
	case "system_model":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			SystemModel:     &inputPayload.StringValue,
		}
	case "system_sku":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			SystemSKU:       &inputPayload.StringValue,
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
			Tagnumber:    tagnumber,
			SystemSerial: systemSerial,
			SystemUptime: &uptimeSeconds,
		}
	case "system_uuid":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			SystemUUID:      &inputPayload.StringValue,
		}
	case "tpm_version":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			TPMVersion:      &inputPayload.StringValue,
		}
	case "wifi_mac":
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware"}
		inputPayload.Value = &ClientHardwareView{
			Tagnumber:       tagnumber,
			SystemSerial:    systemSerial,
			TransactionUUID: *inputPayload.TransactionUUID,
			WiFiMAC:         &inputPayload.StringValue,
		}
	default:
		return nil, fmt.Errorf("unsupported key: '%s'", inputPayload.Key)
	}

	return &HTTPRequest{
		Config:  httpRequestConfig,
		Payload: inputPayload,
	}, nil
}
