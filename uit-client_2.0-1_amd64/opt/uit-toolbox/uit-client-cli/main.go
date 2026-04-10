package main

import (
	"bufio"
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

type InputData struct {
	TagNum int64
	Key    string
	Value  any
	UUID   *string
}

type HTTPRequestData struct {
	TagNum int64   `json:"tag_num"`
	Key    string  `json:"key"`
	Value  any     `json:"value"`
	UUID   *string `json:"uuid,omitempty"`
	URL    url.URL `json:"url"`
	Method string  `json:"http_method"`
}

var clientConfig atomic.Pointer[ClientConfig]

func GetClientConfig() (*ClientConfig, error) {
	reqURL := &url.URL{
		Host:   "10.0.0.1:8080",
		Scheme: "http",
		Path:   "/static/client/configs/uit-client",
	}
	queries := url.Values{}
	queries.Set("json", "true")
	reqURL.RawQuery = queries.Encode()

	resp, err := http.Get(reqURL.String())
	if err != nil {
		return nil, fmt.Errorf("error in GetClientConfig: %w", err)
	}
	defer resp.Body.Close()

	var configData ClientConfig
	if err := json.NewDecoder(resp.Body).Decode(&configData); err != nil {
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
		return
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

	httpRequestData, err := createArrayFromInput(clean)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create array from input: %v\n", err)
		return
	}

	if err := sendRequest(httpRequestData); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %v\n", err)
		return
	}
}

func createArrayFromInput(input string) (*HTTPRequestData, error) {
	httpRequestData := new(HTTPRequestData)

	inputArr := strings.Split(input, "|")
	if inputArr[0] == "" {
		return nil, fmt.Errorf("input missing tag number")
	}
	if inputArr[1] == "" {
		return nil, fmt.Errorf("input missing key")
	}
	if inputArr[2] == "" {
		return nil, fmt.Errorf("input missing value")
	}
	if len(inputArr) > 3 && inputArr[3] == "" {
		return nil, fmt.Errorf("input has empty UUID")
	}

	var err error
	httpRequestData.TagNum, err = strconv.ParseInt(inputArr[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TagNum: %w", err)
	}
	httpRequestData.Key = inputArr[1]
	if strings.TrimSpace(inputArr[3]) != "" {
		httpRequestData.UUID = &inputArr[3]
	}

	if httpRequestData.Key == "cpu_usage" {
		httpRequestData.Value, err = strconv.ParseFloat(inputArr[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid cpu_usage value: %w", err)
		}
		httpRequestData.URL = url.URL{Path: "/api/client/cpu/usage"}
		httpRequestData.Method = "POST"
	}

	return httpRequestData, nil
}

func sendRequest(data *HTTPRequestData) error {
	if data.Method != "POST" && data.Method != "GET" {
		return fmt.Errorf("unsupported HTTP method: %s", data.Method)
	}
	if strings.TrimSpace(data.Key) == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if strings.TrimSpace(data.URL.Path) == "" {
		return fmt.Errorf("relative URL cannot be empty")
	}
	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
	}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30,
		DisableCompression: true,
		TLSClientConfig:    tlsConfig,
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	requestURL := &url.URL{
		Scheme:   "https",
		Host:     fmt.Sprintf("%s:%s", clientConfig.Load().UIT_WEB_HTTPS_HOST, clientConfig.Load().UIT_WEB_HTTPS_PORT),
		Path:     data.URL.Path,
		RawQuery: data.URL.RawQuery,
	}

	req, err := http.NewRequest(data.Method, requestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	if data.Method == "POST" {
		req.Body = io.NopCloser(strings.NewReader(string(jsonData)))
	} else {
		req.Body = http.NoBody
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server returned an HTTP error: %d", resp.StatusCode)
	}

	return nil
}
