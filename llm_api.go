package sa

import (
	"context"

	llmservice "github.com/seaart/sa-go/internal/llm/service"
)

func (l *LLMService) ChatCompletions(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	return llmservice.ChatCompletions(l.client, ctx, payload, buildRequestOptions(opts).headers)
}

func (l *LLMService) ChatCompletionsStream(ctx context.Context, payload JSONMap, opts ...RequestOption) (<-chan LLMStreamEvent, error) {
	return llmservice.ChatCompletionsStream(l.client, ctx, payload, buildRequestOptions(opts).headers)
}

func (l *LLMService) Messages(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	return llmservice.Messages(l.client, ctx, payload, buildRequestOptions(opts).headers)
}

func (l *LLMService) MessagesStream(ctx context.Context, payload JSONMap, opts ...RequestOption) (<-chan LLMStreamEvent, error) {
	return llmservice.MessagesStream(l.client, ctx, payload, buildRequestOptions(opts).headers)
}

func (l *LLMService) Responses(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	return llmservice.Responses(l.client, ctx, payload, buildRequestOptions(opts).headers)
}

func (l *LLMService) ResponsesStream(ctx context.Context, payload JSONMap, opts ...RequestOption) (<-chan LLMStreamEvent, error) {
	return llmservice.ResponsesStream(l.client, ctx, payload, buildRequestOptions(opts).headers)
}

func (l *LLMService) Rerank(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	return llmservice.Rerank(l.client, ctx, payload, buildRequestOptions(opts).headers)
}

func (l *LLMService) Embeddings(ctx context.Context, payload JSONMap, opts ...RequestOption) (RawResponse, error) {
	return llmservice.Embeddings(l.client, ctx, payload, buildRequestOptions(opts).headers)
}

func (l *LLMService) ListModels(ctx context.Context, opts ...RequestOption) (RawResponse, error) {
	return llmservice.ListModels(l.client, ctx, buildRequestOptions(opts).headers)
}
