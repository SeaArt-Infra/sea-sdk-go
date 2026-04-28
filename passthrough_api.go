package sa

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/seaart/sa-go/internal/shared"
)

// Request sends one vendor-compatible passthrough request with a JSON-encoded body.
func (p *PassthroughService) Request(ctx context.Context, method, path string, body any, opts ...RequestOption) (*PassthroughResponse, error) {
	var payload []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, &Error{
				Kind:    ErrGeneral,
				Message: "failed to marshal request: " + err.Error(),
			}
		}
		payload = b
	}
	return p.RequestRaw(ctx, method, path, payload, opts...)
}

// RequestRaw sends one vendor-compatible passthrough request using body as-is.
func (p *PassthroughService) RequestRaw(ctx context.Context, method, path string, body []byte, opts ...RequestOption) (*PassthroughResponse, error) {
	normalizedPath, err := normalizePassthroughPath(path)
	if err != nil {
		return nil, err
	}

	status, headers, payload, err := p.client.RequestRaw(ctx, method, normalizedPath, body, buildRequestOptions(opts).headers)
	if err != nil {
		return nil, err
	}

	return &PassthroughResponse{
		StatusCode: status,
		Headers:    headers,
		Body:       RawResponse(payload),
	}, nil
}

func (p *PassthroughService) Get(ctx context.Context, path string, opts ...RequestOption) (*PassthroughResponse, error) {
	return p.RequestRaw(ctx, http.MethodGet, path, nil, opts...)
}

func (p *PassthroughService) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*PassthroughResponse, error) {
	return p.Request(ctx, http.MethodPost, path, body, opts...)
}

func (p *PassthroughService) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*PassthroughResponse, error) {
	return p.Request(ctx, http.MethodPut, path, body, opts...)
}

func (p *PassthroughService) Delete(ctx context.Context, path string, body any, opts ...RequestOption) (*PassthroughResponse, error) {
	return p.Request(ctx, http.MethodDelete, path, body, opts...)
}

func normalizePassthroughPath(raw string) (string, error) {
	path := strings.TrimSpace(raw)
	if path == "" {
		return "", &shared.Error{Kind: shared.ErrGeneral, Message: "passthrough path is required"}
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return "", &shared.Error{Kind: shared.ErrGeneral, Message: "passthrough path must be relative"}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path, nil
}
