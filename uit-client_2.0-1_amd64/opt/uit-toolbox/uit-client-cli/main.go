//go:build linux && amd64

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
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

	jsonBody, err := sendHTTPRequest(httpRequest)
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

	// fmt.Printf("received stdin data: %s\n", clean)

	httpRequest, err := MapInputToHTTPRequest(clean)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create array from input: %v\n", err)
		return
	}

	if _, err := sendHTTPRequest(httpRequest); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %v\n", err)
		return
	}
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
