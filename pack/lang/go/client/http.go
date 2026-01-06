// Package client provides HTTP client for the {{camel .ServiceName}} service.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"micromanager/pack/lang/go/api"
	"net/http"

	"{{snake .ServiceName}}/api"
)

type HttpClient struct {
	baseURL string
	http    *http.Client
}

var _ api.Service = (*HttpClient)(nil)

type HttpClientOption func(*HttpClient)

func NewHttpClient(baseURL string, opts ...HttpClientOption) *HttpClient {
	c := &HttpClient{
		baseURL: baseURL,
		http:    http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *HttpClient) SayHello(ctx context.Context, req api.HelloRequest) (*api.HelloResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := c.baseURL + "/{{snake .ServiceName}}/v1/hello"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, parseErrorResponse(resp)
	}

	var result api.HelloResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func parseErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
}
