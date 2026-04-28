package sa

import (
	"encoding/json"

	llmtypes "github.com/seaart/sa-go/internal/llm/types"
)

type JSONMap = llmtypes.JSONMap
type RawResponse = json.RawMessage
type LLMStreamEvent = llmtypes.StreamEvent
type LLMMessage = llmtypes.LLMMessage
type LLMToolCall = llmtypes.LLMToolCall
type LLMFunctionCall = llmtypes.LLMFunctionCall
type ChatCompletionResponse = llmtypes.ChatCompletionResponse
type ChatCompletionChoice = llmtypes.ChatCompletionChoice
type MessagesResponse = llmtypes.MessagesResponse
type MessagesContentBlock = llmtypes.MessagesContentBlock
type MessagesStreamChunk = llmtypes.MessagesStreamChunk
type MessagesStreamMessage = llmtypes.MessagesStreamMessage
type MessagesStreamContentBlock = llmtypes.MessagesStreamContentBlock
type MessagesStreamDelta = llmtypes.MessagesStreamDelta
type MessagesStreamTextAssembler = llmtypes.MessagesStreamTextAssembler
type ResponsesResponse = llmtypes.ResponsesResponse
type ResponsesResponseStreamChunk = llmtypes.ResponsesResponseStreamChunk
type ResponsesStreamOutputItem = llmtypes.ResponsesStreamOutputItem
type ResponsesStreamContentPart = llmtypes.ResponsesStreamContentPart
type ResponsesStreamTextAssembler = llmtypes.ResponsesStreamTextAssembler
type ResponsesOutputItem = llmtypes.ResponsesOutputItem
type ResponsesContentItem = llmtypes.ResponsesContentItem
type RerankResponse = llmtypes.RerankResponse
type RerankResult = llmtypes.RerankResult
type RerankResponseMeta = llmtypes.RerankResponseMeta
type RerankBilledUnits = llmtypes.RerankBilledUnits
type RerankTokens = llmtypes.RerankTokens
type RerankUsage = llmtypes.RerankUsage
type EmbeddingsResponse = llmtypes.EmbeddingsResponse
type EmbeddingObject = llmtypes.EmbeddingObject
type LLMModelListResponse = llmtypes.LLMModelListResponse
type LLMModel = llmtypes.LLMModel
type LLMUsage = llmtypes.LLMUsage

func Decode[T any](raw RawResponse) (*T, error) {
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, &Error{
			Kind:    ErrGeneral,
			Message: "failed to decode response: " + err.Error(),
		}
	}
	return &out, nil
}
