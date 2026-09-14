package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

// HTTPClient is a small wrapper around http.Client for making GET and POST
// requests against a common base URL.
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient creates an HTTPClient that prefixes every request with baseURL.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// Get issues a GET request to path, applying params as URL query parameters
// and headers as request headers. It returns the raw response, the response
// body, and any error encountered.
func (h *HTTPClient) Get(ctx context.Context, path string, params map[string]string, headers map[string]string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.baseURL+path, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("building GET request: %w", err)
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	req.URL.RawQuery = q.Encode()

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.do(req)
}

// Post issues a POST request to path, applying headers as request headers
// and JSON-encoding data as the request body. It returns the raw response,
// the response body, and any error encountered.
func (h *HTTPClient) Post(ctx context.Context, path string, headers map[string]string, data map[string]interface{}) (*http.Response, []byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding POST body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("building POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.do(req)
}

func logOutgoingRequest(req *http.Request) {
	headers := make([]string, 0, len(req.Header))
	for key, values := range req.Header {
		if strings.EqualFold(key, "Authorization") {
			headers = append(headers, key+": [redacted]")
			continue
		}
		headers = append(headers, key+": "+strings.Join(values, ", "))
	}

	sort.Strings(headers)
	log.Printf("httpclient: %s %s headers=[%s]", req.Method, req.URL.String(), strings.Join(headers, "; "))
}

// do executes req and reads the full response body.
func (h *HTTPClient) do(req *http.Request) (*http.Response, []byte, error) {
	logOutgoingRequest(req)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("executing %s request to %s: %w", req.Method, req.URL, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("reading response body for %s %s: %w", req.Method, req.URL, err)
	}

	return resp, respBody, nil
}
