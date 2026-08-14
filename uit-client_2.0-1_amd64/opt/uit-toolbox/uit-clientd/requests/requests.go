package requests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	client *http.Client
	tr     *http.Transport
)

func initRequests() error {
	tr = &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client = &http.Client{
		Transport: tr,
	}
	return nil
}

func constructURL(u url.URL) url.URL {
	return url.URL{
		Scheme:   "https",
		Host:     "10.0.0.1",
		Path:     u.Path,
		RawQuery: u.RawQuery,
	}
}

func getRequest(ctx context.Context, u url.URL) ([]byte, error) {
	if client == nil || tr == nil {
		initRequests()
	}
	merged := constructURL(u)

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	resp, err := client.Get(merged.String())
	if err != nil {
		return nil, fmt.Errorf("error GETing request from '%s' (%d): %w", merged.String(), resp.StatusCode, err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func postRequest(ctx context.Context, u url.URL, contentType string, body io.Reader) error {
	if client == nil || tr == nil {
		initRequests()
	}
	merged := constructURL(u)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	resp, err := client.Post(merged.String(), contentType, body)
	if err != nil {
		return fmt.Errorf("error POSTing request to '%s' (%d): %w", merged.String(), resp.StatusCode, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("response from '%s' returned non-200 value %d (POST)", merged.String(), resp.StatusCode)
	}

	return nil
}
