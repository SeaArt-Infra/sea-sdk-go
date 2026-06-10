package service

import (
	"context"
	"encoding/json"
	"net/http"

	llmtypes "github.com/SeaArt-Infra/sea-sdk-go/internal/llm/types"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/shared"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/transport"
)

const (
	PathChatCompletions = "/chat/completions"
	PathMessages        = "/v1/messages"
	PathResponses       = "/responses"
	PathRerank          = "/rerank"
	PathEmbeddings      = "/v1/embeddings"
	PathModels          = "/v1/models"
)

func ChatCompletions(client *transport.Client, ctx context.Context, payload llmtypes.JSONMap, headers http.Header) (llmtypes.RawResponse, error) {
	if isStreaming(payload) {
		return nil, unsupportedStreamingError("ChatCompletions", "ChatCompletionsStream")
	}
	return doRawJSON(client, ctx, http.MethodPost, PathChatCompletions, payload, headers)
}

func Messages(client *transport.Client, ctx context.Context, payload llmtypes.JSONMap, headers http.Header) (llmtypes.RawResponse, error) {
	if isStreaming(payload) {
		return nil, unsupportedStreamingError("Messages", "MessagesStream")
	}
	return doRawJSON(client, ctx, http.MethodPost, PathMessages, payload, headers)
}

func Responses(client *transport.Client, ctx context.Context, payload llmtypes.JSONMap, headers http.Header) (llmtypes.RawResponse, error) {
	if isStreaming(payload) {
		return nil, unsupportedStreamingError("Responses", "ResponsesStream")
	}
	return doRawJSON(client, ctx, http.MethodPost, PathResponses, payload, headers)
}

func Rerank(client *transport.Client, ctx context.Context, payload llmtypes.JSONMap, headers http.Header) (llmtypes.RawResponse, error) {
	return doRawJSON(client, ctx, http.MethodPost, PathRerank, payload, headers)
}

func Embeddings(client *transport.Client, ctx context.Context, payload llmtypes.JSONMap, headers http.Header) (llmtypes.RawResponse, error) {
	return doRawJSON(client, ctx, http.MethodPost, PathEmbeddings, payload, headers)
}

func ListModels(client *transport.Client, ctx context.Context, headers http.Header) (llmtypes.RawResponse, error) {
	return doRawJSON(client, ctx, http.MethodGet, PathModels, nil, headers)
}

func doRawJSON(client *transport.Client, ctx context.Context, method, path string, body any, headers http.Header) (llmtypes.RawResponse, error) {
	status, payload, err := client.Request(ctx, method, path, body, headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, httpError(status, payload)
	}
	return llmtypes.RawResponse(payload), nil
}

func isStreaming(payload llmtypes.JSONMap) bool {
	stream, ok := payload["stream"].(bool)
	return ok && stream
}

func unsupportedStreamingError(methodName, streamMethod string) error {
	return &shared.Error{
		Kind:    shared.ErrGeneral,
		Message: "stream=true is not supported by " + methodName + "; use " + streamMethod + " instead",
	}
}

func httpError(status int, payload []byte) error {
	var apiErr struct {
		Message string `json:"message"`
		Error   *struct {
			Message      string `json:"message"`
			ErrorMessage string `json:"error_message"`
			Type         string `json:"type"`
		} `json:"error"`
	}
	_ = json.Unmarshal(payload, &apiErr)

	message := http.StatusText(status)
	switch {
	case apiErr.Error != nil && apiErr.Error.ErrorMessage != "":
		message = apiErr.Error.ErrorMessage
	case apiErr.Error != nil && apiErr.Error.Message != "":
		message = apiErr.Error.Message
	case apiErr.Message != "":
		message = apiErr.Message
	case message == "":
		message = "HTTP error"
	}

	return shared.NewHTTPError(status, message)
}
