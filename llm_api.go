package sa

import (
	"context"
	"net/http"

	llmservice "github.com/SeaArt-Infra/sea-sdk-go/internal/llm/service"
)

func (l *LLMService) ChatCompletions(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	body, headers, err := l.requestModel(payload, opts)
	if err != nil {
		return nil, err
	}
	return llmservice.ChatCompletions(l.client, ctx, body, headers)
}

func (l *LLMService) ChatCompletionsStream(ctx context.Context, payload JSONMap, opts ...RequestOption) (<-chan LLMStreamEvent, error) {
	body, headers, err := l.requestModel(payload, opts)
	if err != nil {
		return nil, err
	}
	return llmservice.ChatCompletionsStream(l.client, ctx, body, headers)
}

func (l *LLMService) Messages(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	body, headers, err := l.requestModel(payload, opts)
	if err != nil {
		return nil, err
	}
	return llmservice.Messages(l.client, ctx, body, headers)
}

func (l *LLMService) MessagesStream(ctx context.Context, payload JSONMap, opts ...RequestOption) (<-chan LLMStreamEvent, error) {
	body, headers, err := l.requestModel(payload, opts)
	if err != nil {
		return nil, err
	}
	return llmservice.MessagesStream(l.client, ctx, body, headers)
}

func (l *LLMService) Responses(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	body, headers, err := l.requestModel(payload, opts)
	if err != nil {
		return nil, err
	}
	return llmservice.Responses(l.client, ctx, body, headers)
}

func (l *LLMService) ResponsesStream(ctx context.Context, payload JSONMap, opts ...RequestOption) (<-chan LLMStreamEvent, error) {
	body, headers, err := l.requestModel(payload, opts)
	if err != nil {
		return nil, err
	}
	return llmservice.ResponsesStream(l.client, ctx, body, headers)
}

func (l *LLMService) Rerank(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	body, headers, err := l.requestModel(payload, opts)
	if err != nil {
		return nil, err
	}
	return llmservice.Rerank(l.client, ctx, body, headers)
}

func (l *LLMService) Embeddings(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	body, headers, err := l.requestModel(payload, opts)
	if err != nil {
		return nil, err
	}
	return llmservice.Embeddings(l.client, ctx, body, headers)
}

func (l *LLMService) ListModels(ctx context.Context, opts ...RequestOption) (RawResponse, error) {
	return llmservice.ListModels(l.client, ctx, buildRequestOptions(opts).headers)
}

func (l *LLMService) requestModel(payload JSONMap, opts []RequestOption) (JSONMap, http.Header, error) {
	return moveModelToHeader(payload, buildRequestOptions(opts).headers)
}
