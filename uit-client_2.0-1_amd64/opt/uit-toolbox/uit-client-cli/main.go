//go:build linux && amd64

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var clientConfig atomic.Pointer[ClientConfig]

const unixSocketPath = "/run/uit-client/uit-client-cli.sock"

func GetClientConfig() (*ClientConfig, error) {
	reqURL := &url.URL{
		Scheme:   "https",
		Host:     "10.0.0.1:31411",
		Path:     "/static/client/configs/uit-client",
		RawQuery: "json=true",
	}
	queries := url.Values{}
	queries.Set("json", "true")
	reqURL.RawQuery = queries.Encode()

	httpRequestConfig := new(HTTPRequestConfig)
	httpRequestConfig.URL = *reqURL
	httpRequestConfig.Method = "GET"

	httpRequest := &HTTPRequest{
		Config:  httpRequestConfig,
		Payload: nil,
	}

	jsonBody, err := sendRequest(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("error in GetClientConfig: %w", err)
	}
	if len(jsonBody) == 0 {
		return nil, fmt.Errorf("received nil or empty response body in GetClientConfig")
	}

	var configData ClientConfig
	if err := json.NewDecoder(bytes.NewReader(jsonBody)).Decode(&configData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GetClientConfig response: %w", err)
	}

	return &configData, nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	config, err := GetClientConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get client config: %v\n", err)
		os.Exit(1)
	}
	if config == nil || strings.TrimSpace(config.UIT_WEB_HTTPS_HOST) == "" || strings.TrimSpace(config.UIT_WEB_HTTPS_PORT) == "" {
		fmt.Fprintf(os.Stderr, "client config is invalid\n")
		os.Exit(1)
	}
	clientConfig.Store(config)

	listener, inherited, err := getUnixSocketListener()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to acquire unix socket listener: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = listener.Close()
		if !inherited {
			_ = os.Remove(unixSocketPath)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	var wg sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				wg.Wait()
				return
			}
			fmt.Fprintf(os.Stderr, "unix socket accept error: %v\n", err)
			continue
		}

		wg.Add(1)
		go handleConnection(ctx, conn, &wg)
	}
}

func getUnixSocketListener() (net.Listener, bool, error) {
	listener, err := inheritedUnixSocketListener()
	if err == nil {
		return listener, true, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, false, err
	}

	listener, err = net.Listen("unix", unixSocketPath)
	if err != nil {
		return nil, false, err
	}

	if err := os.Chmod(unixSocketPath, 0660); err != nil {
		_ = listener.Close()
		_ = os.Remove(unixSocketPath)
		return nil, false, fmt.Errorf("failed to chmod unix socket: %w", err)
	}

	return listener, false, nil
}

func inheritedUnixSocketListener() (net.Listener, error) {
	listenPID := os.Getenv("LISTEN_PID")
	listenFDs := os.Getenv("LISTEN_FDS")
	if listenPID == "" || listenFDs == "" {
		return nil, os.ErrNotExist
	}

	pid, err := strconv.Atoi(listenPID)
	if err != nil {
		return nil, fmt.Errorf("invalid LISTEN_PID: %w", err)
	}
	if pid != os.Getpid() {
		return nil, os.ErrNotExist
	}

	fds, err := strconv.Atoi(listenFDs)
	if err != nil {
		return nil, fmt.Errorf("invalid LISTEN_FDS: %w", err)
	}
	if fds < 1 {
		return nil, os.ErrNotExist
	}

	file := os.NewFile(uintptr(3), "systemd-unix-socket")
	if file == nil {
		return nil, fmt.Errorf("failed to access inherited systemd socket")
	}
	defer file.Close()

	listener, err := net.FileListener(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener from inherited socket: %w", err)
	}

	return listener, nil
}

func handleConnection(ctx context.Context, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		handleInput(ctx, scanner.Text())
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "unix socket read error: %v\n", err)
	}
}

func handleInput(ctx context.Context, stdinData string) {

	select {
	case <-ctx.Done():
		return
	default:
	}

	clean := strings.TrimSpace(stdinData)
	if clean == "" {
		return
	}

	fmt.Printf("received stdin data: %s\n", clean)

	httpRequest, err := MapInputToHTTPRequest(clean)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create array from input: %v\n", err)
		return
	}

	if _, err := sendRequest(httpRequest); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %v\n", err)
		return
	}
}

func MapInputToHTTPRequest(input string) (*HTTPRequest, error) {
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

	switch httpRequestPayload.Key {
	case "battery_charge_pcnt":
		httpRequestPayload.Key = "battery_charge_pcnt"
		batteryPcnt, err := strconv.ParseFloat(inputArr[2], 64)
		if err != nil || batteryPcnt < 0 || batteryPcnt > 110 {
			return nil, fmt.Errorf("invalid battery_charge_pcnt value: %w", err)
		}
		httpRequestPayload.Value = &BatteryData{
			Tagnumber: httpRequestPayload.Tagnumber,
			Percent:   &batteryPcnt,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/hardware/battery"}
		httpRequestConfig.Method = "POST"
	case "system_uptime":
		httpRequestPayload.Key = "system_uptime"
		uptimeSeconds, err := strconv.ParseInt(inputArr[2], 10, 64)
		if err != nil || uptimeSeconds < 0 {
			return nil, fmt.Errorf("invalid system_uptime value: %w", err)
		}
		httpRequestPayload.Value = &ClientUptime{
			Tagnumber:    httpRequestPayload.Tagnumber,
			SystemUptime: uptimeSeconds,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/uptime"}
		httpRequestConfig.Method = "POST"
	case "client_app_uptime":
		httpRequestPayload.Key = "client_app_uptime"
		uptimeSeconds, err := strconv.ParseInt(inputArr[2], 10, 64)
		if err != nil || uptimeSeconds < 0 {
			return nil, fmt.Errorf("invalid client_app_uptime value: %w", err)
		}
		httpRequestPayload.Value = &ClientUptime{
			Tagnumber:       httpRequestPayload.Tagnumber,
			ClientAppUptime: uptimeSeconds,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/uptime"}
		httpRequestConfig.Method = "POST"
	case "cpu_usage":
		httpRequestPayload.Key = "cpu_usage"
		cpuUsage, err := strconv.ParseFloat(inputArr[2], 64)
		if err != nil || cpuUsage < 0 || cpuUsage > 110 {
			return nil, fmt.Errorf("invalid cpu_usage value: %w", err)
		}
		httpRequestPayload.Value = &CPUData{
			Tagnumber:    httpRequestPayload.Tagnumber,
			UsagePercent: &cpuUsage,
		}
		httpRequestConfig.URL = url.URL{Path: "/api/client/cpu/usage"}
		httpRequestConfig.Method = "POST"
	default:
		return nil, fmt.Errorf("unsupported key: '%s'", httpRequestPayload.Key)
	}

	return &HTTPRequest{
		Config:  httpRequestConfig,
		Payload: httpRequestPayload,
	}, nil
}

func sendRequest(data *HTTPRequest) ([]byte, error) {
	if data == nil || data.Config == nil {
		return nil, fmt.Errorf("data variable and/or config is nil")
	}
	if data.Config.Method != "POST" && data.Config.Method != "GET" {
		return nil, fmt.Errorf("unsupported HTTP method: %s", data.Config.Method)
	}
	if strings.TrimSpace(data.Config.URL.Path) == "" {
		return nil, fmt.Errorf("relative URL cannot be empty")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
	}

	tr := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		TLSClientConfig:     tlsConfig,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	protocols := http.Protocols{}
	protocols.SetHTTP1(false)
	protocols.SetUnencryptedHTTP2(false)
	protocols.SetHTTP2(true)
	tr.Protocols = &protocols

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
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
	resp, err := client.Do(req)
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
