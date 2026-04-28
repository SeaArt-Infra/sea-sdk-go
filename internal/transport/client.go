package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/seaart/sa-go/internal/shared"
)

// Client owns the shared HTTP transport state used by service packages.
type Client struct {
	APIKey     string
	BaseURL    string
	Project    string
	UserAgent  string
	HTTPClient *http.Client
}

func (c *Client) buildRequest(ctx context.Context, method, path string, body any, headers http.Header) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, &shared.Error{
				Kind:    shared.ErrGeneral,
				Message: "failed to marshal request: " + err.Error(),
			}
		}
		bodyReader = bytes.NewReader(b)
	}

	return c.newRequest(ctx, method, path, bodyReader, headers)
}

func (c *Client) buildRawRequest(ctx context.Context, method, path string, body []byte, headers http.Header) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := c.newRequest(ctx, method, path, bodyReader, headers)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader, headers http.Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return nil, &shared.Error{
			Kind:    shared.ErrGeneral,
			Message: "failed to build request: " + err.Error(),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", c.UserAgent)
	if c.Project != "" {
		req.Header.Set("X-Project", c.Project)
	}
	for key, values := range headers {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return req, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, headers http.Header) (*http.Response, error) {
	req, err := c.buildRequest(ctx, method, path, body, headers)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &shared.Error{
			Kind:    shared.ErrNetwork,
			Message: "request failed: " + err.Error(),
		}
	}

	return resp, nil
}

func (c *Client) doRawRequest(ctx context.Context, method, path string, body []byte, headers http.Header) (*http.Response, error) {
	req, err := c.buildRawRequest(ctx, method, path, body, headers)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &shared.Error{
			Kind:    shared.ErrNetwork,
			Message: "request failed: " + err.Error(),
		}
	}

	return resp, nil
}

// Request executes one HTTP request and returns the raw status code and body.
func (c *Client) Request(ctx context.Context, method, path string, body any, headers http.Header) (int, []byte, error) {
	resp, err := c.doRequest(ctx, method, path, body, headers)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, &shared.Error{
			Kind:    shared.ErrGeneral,
			Message: "failed to read response: " + err.Error(),
		}
	}

	return resp.StatusCode, payload, nil
}

// RequestRaw executes one HTTP request using body as-is and returns the raw status code, headers, and body.
func (c *Client) RequestRaw(ctx context.Context, method, path string, body []byte, headers http.Header) (int, http.Header, []byte, error) {
	resp, err := c.doRawRequest(ctx, method, path, body, headers)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, &shared.Error{
			Kind:    shared.ErrGeneral,
			Message: "failed to read response: " + err.Error(),
		}
	}

	return resp.StatusCode, resp.Header.Clone(), payload, nil
}

// RequestStream executes one HTTP request and returns the open response body for streaming callers.
func (c *Client) RequestStream(ctx context.Context, method, path string, body any, headers http.Header) (*http.Response, error) {
	return c.doRequest(ctx, method, path, body, headers)
}
