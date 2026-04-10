//go:build linux && amd64

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
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

type ClientConfig struct {
	UIT_CLIENT_DB_USER   string `json:"UIT_CLIENT_DB_USER"`
	UIT_CLIENT_DB_PASSWD string `json:"UIT_CLIENT_DB_PASSWD"`
	UIT_CLIENT_DB_NAME   string `json:"UIT_CLIENT_DB_NAME"`
	UIT_CLIENT_DB_HOST   string `json:"UIT_CLIENT_DB_HOST"`
	UIT_CLIENT_DB_PORT   string `json:"UIT_CLIENT_DB_PORT"`
	UIT_CLIENT_NTP_HOST  string `json:"UIT_CLIENT_NTP_HOST"`
	UIT_CLIENT_PING_HOST string `json:"UIT_CLIENT_PING_HOST"`
	UIT_SERVER_HOSTNAME  string `json:"UIT_SERVER_HOSTNAME"`
	UIT_WEB_HTTP_HOST    string `json:"UIT_WEB_HTTP_HOST"`
	UIT_WEB_HTTP_PORT    string `json:"UIT_WEB_HTTP_PORT"`
	UIT_WEB_HTTPS_HOST   string `json:"UIT_WEB_HTTPS_HOST"`
	UIT_WEB_HTTPS_PORT   string `json:"UIT_WEB_HTTPS_PORT"`
	UIT_WEBMASTER_NAME   string `json:"UIT_WEBMASTER_NAME"`
	UIT_WEBMASTER_EMAIL  string `json:"UIT_WEBMASTER_EMAIL"`
}

type HTTPRequest struct {
	Config  *HTTPRequestConfig
	Payload *HTTPRequestPayload
}

type HTTPRequestConfig struct {
	URL    url.URL
	Method string
}

type HTTPRequestPayload struct {
	Tagnumber int64   `json:"tagnumber"`
	Key       string  `json:"key"`
	Value     any     `json:"value"`
	UUID      *string `json:"uuid,omitempty"`
}

var clientConfig atomic.Pointer[ClientConfig]

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

	stdinCh := make(chan string)
	errCh := make(chan error, 1)

	go readStdinToChannel(ctx, os.Stdin, stdinCh, errCh)

	var wg sync.WaitGroup
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case err := <-errCh:
			if err != nil {
				fmt.Fprintf(os.Stderr, "stdin read error: %v\n", err)
			}
			wg.Wait()
			return
		case data, ok := <-stdinCh:
			if !ok {
				wg.Wait()
				return
			}

			wg.Add(1)
			go handleInput(ctx, data, &wg)
		}
	}
}

func readStdinToChannel(ctx context.Context, input *os.File, out chan<- string, errCh chan<- error) {
	defer close(out)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()

		select {
		case <-ctx.Done():
			return
		case out <- line:
		}
	}

	if err := scanner.Err(); err != nil {
		errCh <- err
	}
}

func handleInput(ctx context.Context, stdinData string, wg *sync.WaitGroup) {
	defer wg.Done()

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
	case "cpu_usage":
		httpRequestPayload.Key = "cpu_usage"
		httpRequestPayload.Value, err = strconv.ParseFloat(inputArr[2], 64)
		if err != nil || httpRequestPayload.Value.(float64) < 0 || httpRequestPayload.Value.(float64) > 110 {
			return nil, fmt.Errorf("invalid cpu_usage value: %w", err)
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
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	var bodyReader io.Reader = http.NoBody
	if data.Config.Method == "POST" {
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
