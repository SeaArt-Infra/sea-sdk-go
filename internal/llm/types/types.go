package types

import (
	"encoding/json"
	"strings"
)

type JSONMap map[string]any
type RawResponse = json.RawMessage

type StreamEvent struct {
	Event string
	Data  RawResponse
	Done  bool
	Err   error
}

type LLMMessage struct {
	Role       string        `json:"role"`
	Content    any           `json:"content"`
	Name       string        `json:"name,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	ToolCalls  []LLMToolCall `json:"tool_calls,omitempty"`
}

type LLMToolCall struct {
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type,omitempty"`
	Function *LLMFunctionCall `json:"function,omitempty"`
}

type LLMFunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string                 `json:"id,omitempty"`
	Object  string                 `json:"object,omitempty"`
	Created int64                  `json:"created,omitempty"`
	Model   string                 `json:"model,omitempty"`
	Choices []ChatCompletionChoice `json:"choices,omitempty"`
	Usage   *LLMUsage              `json:"usage,omitempty"`
}

type ChatCompletionChoice struct {
	Index        int         `json:"index,omitempty"`
	Message      *LLMMessage `json:"message,omitempty"`
	Delta        *LLMMessage `json:"delta,omitempty"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

type MessagesResponse struct {
	ID         string                 `json:"id,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Role       string                 `json:"role,omitempty"`
	Model      string                 `json:"model,omitempty"`
	Content    []MessagesContentBlock `json:"content,omitempty"`
	StopReason string                 `json:"stop_reason,omitempty"`
	Choices    []ChatCompletionChoice `json:"choices,omitempty"`
	Usage      *LLMUsage              `json:"usage,omitempty"`
}

type MessagesContentBlock struct {
	Type   string         `json:"type,omitempty"`
	Text   string         `json:"text,omitempty"`
	ID     string         `json:"id,omitempty"`
	Name   string         `json:"name,omitempty"`
	Input  map[string]any `json:"input,omitempty"`
	Source any            `json:"source,omitempty"`
}

type MessagesStreamChunk struct {
	Type         string                      `json:"type,omitempty"`
	Index        int                         `json:"index,omitempty"`
	Message      *MessagesStreamMessage      `json:"message,omitempty"`
	ContentBlock *MessagesStreamContentBlock `json:"content_block,omitempty"`
	Delta        *MessagesStreamDelta        `json:"delta,omitempty"`
	Usage        *LLMUsage                   `json:"usage,omitempty"`
}

type MessagesStreamMessage struct {
	ID           string                 `json:"id,omitempty"`
	Type         string                 `json:"type,omitempty"`
	Role         string                 `json:"role,omitempty"`
	Model        string                 `json:"model,omitempty"`
	Content      []MessagesContentBlock `json:"content,omitempty"`
	StopReason   string                 `json:"stop_reason,omitempty"`
	StopSequence any                    `json:"stop_sequence,omitempty"`
	Usage        *LLMUsage              `json:"usage,omitempty"`
}

type MessagesStreamContentBlock struct {
	Type        string         `json:"type,omitempty"`
	Text        string         `json:"text,omitempty"`
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	Input       map[string]any `json:"input,omitempty"`
	Source      any            `json:"source,omitempty"`
	PartialJSON string         `json:"partial_json,omitempty"`
	Thinking    string         `json:"thinking,omitempty"`
	Signature   string         `json:"signature,omitempty"`
}

type MessagesStreamDelta struct {
	Type         string `json:"type,omitempty"`
	Text         string `json:"text,omitempty"`
	PartialJSON  string `json:"partial_json,omitempty"`
	Thinking     string `json:"thinking,omitempty"`
	Signature    string `json:"signature,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence any    `json:"stop_sequence,omitempty"`
}

func (c *MessagesStreamChunk) TextDelta() string {
	if c == nil || c.Delta == nil || c.Delta.Type != "text_delta" {
		return ""
	}
	return c.Delta.Text
}

func (c *MessagesStreamChunk) ThinkingDelta() string {
	if c == nil || c.Delta == nil || c.Delta.Type != "thinking_delta" {
		return ""
	}
	return c.Delta.Thinking
}

func (c *MessagesStreamChunk) InputJSONDelta() string {
	if c == nil || c.Delta == nil || c.Delta.Type != "input_json_delta" {
		return ""
	}
	return c.Delta.PartialJSON
}

type MessagesStreamTextAssembler struct {
	builder strings.Builder
}

func (a *MessagesStreamTextAssembler) Add(chunk *MessagesStreamChunk) {
	if a == nil || chunk == nil {
		return
	}
	if text := chunk.TextDelta(); text != "" {
		a.builder.WriteString(text)
	}
}

func (a *MessagesStreamTextAssembler) Text() string {
	if a == nil {
		return ""
	}
	return a.builder.String()
}

type ResponsesResponse struct {
	ID      string                 `json:"id,omitempty"`
	Object  string                 `json:"object,omitempty"`
	Model   string                 `json:"model,omitempty"`
	Status  string                 `json:"status,omitempty"`
	Output  []ResponsesOutputItem  `json:"output,omitempty"`
	Choices []ChatCompletionChoice `json:"choices,omitempty"`
	Usage   *LLMUsage              `json:"usage,omitempty"`
}

type ResponsesResponseStreamChunk struct {
	Type           string                      `json:"type,omitempty"`
	SequenceNumber int                         `json:"sequence_number,omitempty"`
	ResponseID     string                      `json:"response_id,omitempty"`
	OutputIndex    int                         `json:"output_index,omitempty"`
	ContentIndex   int                         `json:"content_index,omitempty"`
	ItemID         string                      `json:"item_id,omitempty"`
	Response       *ResponsesResponse          `json:"response,omitempty"`
	Item           *ResponsesStreamOutputItem  `json:"item,omitempty"`
	Part           *ResponsesStreamContentPart `json:"part,omitempty"`
	Delta          string                      `json:"delta,omitempty"`
	Text           string                      `json:"text,omitempty"`
	Annotation     map[string]any              `json:"annotation,omitempty"`
	Error          any                         `json:"error,omitempty"`
}

type ResponsesStreamOutputItem struct {
	ID        string                       `json:"id,omitempty"`
	Type      string                       `json:"type,omitempty"`
	Status    string                       `json:"status,omitempty"`
	Role      string                       `json:"role,omitempty"`
	Name      string                       `json:"name,omitempty"`
	CallID    string                       `json:"call_id,omitempty"`
	Arguments string                       `json:"arguments,omitempty"`
	Output    string                       `json:"output,omitempty"`
	Content   []ResponsesStreamContentPart `json:"content,omitempty"`
}

type ResponsesStreamContentPart struct {
	Type        string           `json:"type,omitempty"`
	Text        string           `json:"text,omitempty"`
	Annotations []map[string]any `json:"annotations,omitempty"`
}

func (c *ResponsesResponseStreamChunk) TextDelta() string {
	if c == nil || c.Type != "response.output_text.delta" {
		return ""
	}
	return c.Delta
}

func (c *ResponsesResponseStreamChunk) OutputText() string {
	if c == nil {
		return ""
	}
	switch c.Type {
	case "response.output_text.done":
		return c.Text
	case "response.content_part.added":
		if c.Part != nil && c.Part.Type == "output_text" {
			return c.Part.Text
		}
	case "response.content_part.done":
		if c.Part != nil && c.Part.Type == "output_text" {
			return c.Part.Text
		}
	}
	return ""
}

type ResponsesStreamTextAssembler struct {
	builder strings.Builder
}

func (a *ResponsesStreamTextAssembler) Add(chunk *ResponsesResponseStreamChunk) {
	if a == nil || chunk == nil {
		return
	}
	if text := chunk.TextDelta(); text != "" {
		a.builder.WriteString(text)
	}
}

func (a *ResponsesStreamTextAssembler) Text() string {
	if a == nil {
		return ""
	}
	return a.builder.String()
}

type ResponsesOutputItem struct {
	ID        string                 `json:"id,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Role      string                 `json:"role,omitempty"`
	Name      string                 `json:"name,omitempty"`
	CallID    string                 `json:"call_id,omitempty"`
	Arguments string                 `json:"arguments,omitempty"`
	Content   []ResponsesContentItem `json:"content,omitempty"`
}

type ResponsesContentItem struct {
	Type        string           `json:"type,omitempty"`
	Text        string           `json:"text,omitempty"`
	Annotations []map[string]any `json:"annotations,omitempty"`
}

type RerankResponse struct {
	ID      string              `json:"id,omitempty"`
	Results []RerankResult      `json:"results,omitempty"`
	Meta    *RerankResponseMeta `json:"meta,omitempty"`
	Usage   *RerankUsage        `json:"usage,omitempty"`
}

type RerankResult struct {
	Index          int     `json:"index,omitempty"`
	RelevanceScore float64 `json:"relevance_score,omitempty"`
	Document       any     `json:"document,omitempty"`
}

type RerankResponseMeta struct {
	APIVersion  JSONMap            `json:"api_version,omitempty"`
	BilledUnits *RerankBilledUnits `json:"billed_units,omitempty"`
	Tokens      *RerankTokens      `json:"tokens,omitempty"`
}

type RerankBilledUnits struct {
	SearchUnits int `json:"search_units,omitempty"`
	TotalTokens int `json:"total_tokens,omitempty"`
}

type RerankTokens struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
}

type RerankUsage struct {
	TotalTokens int `json:"total_tokens,omitempty"`
}

type EmbeddingsResponse struct {
	Object string            `json:"object,omitempty"`
	Data   []EmbeddingObject `json:"data,omitempty"`
	Model  string            `json:"model,omitempty"`
	Usage  *LLMUsage         `json:"usage,omitempty"`
}

type EmbeddingObject struct {
	Object    string `json:"object,omitempty"`
	Index     int    `json:"index,omitempty"`
	Embedding any    `json:"embedding,omitempty"`
}

type LLMModelListResponse struct {
	Object string     `json:"object,omitempty"`
	Data   []LLMModel `json:"data,omitempty"`
}

type LLMModel struct {
	ID       string         `json:"id,omitempty"`
	Object   string         `json:"object,omitempty"`
	Created  int64          `json:"created,omitempty"`
	OwnedBy  string         `json:"owned_by,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type LLMUsage struct {
	PromptTokens             int `json:"prompt_tokens,omitempty"`
	CompletionTokens         int `json:"completion_tokens,omitempty"`
	TotalTokens              int `json:"total_tokens,omitempty"`
	InputTokens              int `json:"input_tokens,omitempty"`
	OutputTokens             int `json:"output_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}
