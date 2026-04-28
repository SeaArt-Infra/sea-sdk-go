package service

import (
	"context"
	"encoding/json"
	"net/http"

	mmtypes "github.com/seaart/sa-go/internal/multimodal/types"
	"github.com/seaart/sa-go/internal/shared"
	"github.com/seaart/sa-go/internal/transport"
)

const (
	PathGeneration = "/v1/generation"
	PathTask       = "/v1/generation/task/"
)

func CreateTask(client *transport.Client, ctx context.Context, body any, headers http.Header) (*mmtypes.GenerationResponse, error) {
	status, payload, err := client.Request(ctx, http.MethodPost, PathGeneration, body, headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, httpError(status, payload)
	}

	var resp mmtypes.GenerationResponse
	if err := decode(payload, &resp); err != nil {
		return nil, err
	}
	if resp.ID == "" {
		return nil, &shared.Error{Kind: shared.ErrGeneral, Message: "API returned no task ID"}
	}
	return &resp, nil
}

func GetTask(client *transport.Client, ctx context.Context, taskID string, headers http.Header) (*mmtypes.TaskResponse, error) {
	status, payload, err := client.Request(ctx, http.MethodGet, PathTask+taskID, nil, headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, httpError(status, payload)
	}

	var resp mmtypes.TaskResponse
	if err := decode(payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func httpError(status int, payload []byte) error {
	var apiErr struct {
		Error *mmtypes.APIError `json:"error"`
	}
	_ = json.Unmarshal(payload, &apiErr)

	message := "HTTP error"
	if apiErr.Error != nil && apiErr.Error.ErrorMessage != "" {
		message = apiErr.Error.ErrorMessage
	} else {
		message = http.StatusText(status)
		if message == "" {
			message = "HTTP error"
		}
	}
	return shared.NewHTTPError(status, message)
}

func decode(payload []byte, out any) error {
	if err := json.Unmarshal(payload, out); err != nil {
		return &shared.Error{
			Kind:    shared.ErrGeneral,
			Message: "failed to decode response: " + err.Error(),
		}
	}
	return nil
}
