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
	"time"
	"uit-clientd/requests"
)

var (
	clientConfig atomic.Pointer[ClientConfig]
	systemSerial atomic.Pointer[string]
	tagnumber    atomic.Int64
	jobQueueData atomic.Pointer[requests.ClientJobQueueDataResponse]
)

const unixSocketPath = "/run/uit-client/uit-clientd.sock"

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

	jsonBody, err := sendHTTPRequest(context.Background(), httpRequest)
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

func handleInput(ctx context.Context, stdinData string) (string, error) {

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	clean := strings.TrimSpace(stdinData)
	if clean == "" {
		return "", fmt.Errorf("input cannot be empty or whitespace")
	}

	// fmt.Printf("received stdin data: %s\n", clean)

	httpRequest, err := MapInputToHTTPRequest(clean)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create array from input: %v\n", err)
		return "", err
	}

	res, err := sendHTTPRequest(ctx, httpRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request: %v\n", err)
		return "", err
	}
	if len(res) == 0 {
		return "", nil
	}

	return string(res), nil
}

func initListener(rootCtx context.Context, wg *sync.WaitGroup) error {
	listener, inherited, err := getUnixSocketListener()
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
		if !inherited {
			_ = os.Remove(unixSocketPath)
		}
	}()

	go func() {
		<-rootCtx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if rootCtx.Err() != nil || errors.Is(err, net.ErrClosed) {
				fmt.Fprintf(os.Stderr, "shutting down: %v\n", rootCtx.Err())
				return nil // no error on regular shutdown
			}
			fmt.Fprintf(os.Stderr, "unix socket accept error: %v\n", err)
			continue // no app shutdown if error isolated to specific socket connection
		}

		wg.Go((func() {
			if err := handleConnection(rootCtx, conn); err != nil {
				fmt.Fprintf(os.Stderr, "(handleConnection) %v", err)
			}
		}))
	}
}

func main() {
	rootCtx, rootCtxCancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGTERM,
	)
	defer rootCtxCancel()

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

	var wg sync.WaitGroup

	// System serial, set once
	for {
		if rootCtx.Err() != nil {
			fmt.Fprintf(os.Stdout, "(main - system serial # loop): %v", rootCtx.Err())
		}
		if systemSerial.Load() != nil && *systemSerial.Load() != "" {
			break
		}
		s, err := requests.GetSerial(rootCtx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to retrieve system serial, retrying: %v", err)
			continue
		}
		systemSerial.Store(&s)
		time.Sleep(1 * time.Second)
	}

	// Tag number, set once
	for {
		if rootCtx.Err() != nil {
			fmt.Fprintf(os.Stdout, "(main - tag # loop): %v", rootCtx.Err())
		}
		if tagnumber.Load() > 100000 && tagnumber.Load() < 999999 {
			break
		}
		tag, err := requests.GetTagFromSerial(rootCtx, *systemSerial.Load())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to retrieve tag number, retrrying: %v", err)
			continue
		}
		tagnumber.Store(tag)
		time.Sleep(1 * time.Second)
	}

	// Unix socket listener
	wg.Go(func() {
		if err := initListener(rootCtx, &wg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to acquire unix socket listener: %v\n", err)
		}
	})

	// Main app loop
	wg.Go(func() {
		for {
			jqd, err := requests.GetJobQueueData(rootCtx, tagnumber.Load())
			if err != nil {
				fmt.Fprintf(os.Stderr, "error retrieving client job queue data: %v", err)
			}
			jobQueueData.Store(&jqd)

			jobQueueBytes, err := json.Marshal(jobQueueData.Load())
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot unmarshal ClientJobQueueDataResponse JSON (main): %v", err)
			}
			if err := os.WriteFile("/root/job_queue_data", jobQueueBytes, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "error writing client job queue data to disk: %v", err)
			}
			time.Sleep(3 * time.Second)
		}
	})

	wg.Wait()
}
