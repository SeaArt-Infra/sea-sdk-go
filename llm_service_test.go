package sa_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	sa "github.com/seaart/sa-go"
)

func liveLLMClient(t *testing.T) *sa.Client {
	t.Helper()

	if os.Getenv("SA_RUN_LIVE_LLM_TESTS") != "1" {
		t.Skip("set SA_RUN_LIVE_LLM_TESTS=1 to run live LLM gateway tests")
	}

	apiKey := os.Getenv("SA_LIVE_LLM_API_KEY")
	if apiKey == "" {
		t.Fatal("SA_LIVE_LLM_API_KEY is required for live LLM gateway tests")
	}

	baseURL := os.Getenv("SA_LIVE_LLM_BASE_URL")
	if baseURL == "" {
		t.Fatal("SA_LIVE_LLM_BASE_URL is required for live LLM gateway tests")
	}

	client, err := sa.New(&sa.ClientConfig{
		APIKey:     apiKey,
		LLMBaseURL: baseURL,
	})
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	return client
}

func TestLLMChatCompletions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["model"] != "gpt-4o-mini" {
			t.Fatalf("unexpected model: %v", body["model"])
		}
		if body["reasoning_effort"] != "low" {
			t.Fatalf("missing extra field: %v", body["reasoning_effort"])
		}

		writeJSON(w, 200, sa.ChatCompletionResponse{
			ID:    "chat_123",
			Model: "gpt-4o-mini",
			Choices: []sa.ChatCompletionChoice{
				{
					Index: 0,
					Message: &sa.LLMMessage{
						Role:    "assistant",
						Content: "hello",
					},
					FinishReason: "stop",
				},
			},
		})
	})

	raw, err := client.LLM.ChatCompletions(context.Background(), sa.JSONMap{
		"model":            "gpt-4o-mini",
		"messages":         []sa.JSONMap{{"role": "user", "content": "hi"}},
		"max_tokens":       16,
		"reasoning_effort": "low",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := sa.Decode[sa.ChatCompletionResponse](raw)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Message == nil {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestLLMMessages(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/messages" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["max_tokens"] != float64(32) {
			t.Fatalf("unexpected max_tokens: %v", body["max_tokens"])
		}

		writeJSON(w, 200, sa.MessagesResponse{
			ID:    "msg_123",
			Model: "claude-3-5-sonnet",
			Role:  "assistant",
			Content: []sa.MessagesContentBlock{
				{Type: "text", Text: "hello from claude"},
			},
		})
	})

	raw, err := client.LLM.Messages(context.Background(), sa.JSONMap{
		"model":      "claude-3-5-sonnet",
		"messages":   []sa.JSONMap{{"role": "user", "content": "hi"}},
		"max_tokens": 32,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := sa.Decode[sa.MessagesResponse](raw)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if len(resp.Content) != 1 || resp.Content[0].Text != "hello from claude" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestLLMResponses(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/responses" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["input"] != "hello" {
			t.Fatalf("unexpected input: %v", body["input"])
		}

		writeJSON(w, 200, sa.ResponsesResponse{
			ID:     "resp_123",
			Model:  "gpt-4.1-mini",
			Status: "completed",
			Output: []sa.ResponsesOutputItem{
				{
					ID:     "out_1",
					Type:   "message",
					Role:   "assistant",
					Status: "completed",
					Content: []sa.ResponsesContentItem{
						{Type: "output_text", Text: "hello from responses"},
					},
				},
			},
		})
	})

	raw, err := client.LLM.Responses(context.Background(), sa.JSONMap{
		"model":             "gpt-4.1-mini",
		"input":             "hello",
		"max_output_tokens": 16,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := sa.Decode[sa.ResponsesResponse](raw)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if len(resp.Output) != 1 || len(resp.Output[0].Content) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestLLMRerank(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rerank" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["query"] != "mountain lake" {
			t.Fatalf("unexpected query: %v", body["query"])
		}
		if body["priority"] != "latency" {
			t.Fatalf("missing extra field: %v", body["priority"])
		}

		writeJSON(w, 200, sa.RerankResponse{
			ID: "rerank_123",
			Results: []sa.RerankResult{
				{
					Index:          1,
					RelevanceScore: 0.98,
					Document:       map[string]any{"text": "a mountain lake at sunrise"},
				},
			},
			Usage: &sa.RerankUsage{TotalTokens: 24},
		})
	})

	raw, err := client.LLM.Rerank(context.Background(), sa.JSONMap{
		"model":            "qwen3-rerank",
		"query":            "mountain lake",
		"documents":        []string{"city skyline", "mountain lake at sunrise"},
		"top_n":            1,
		"return_documents": true,
		"priority":         "latency",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := sa.Decode[sa.RerankResponse](raw)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if len(resp.Results) != 1 || resp.Results[0].Index != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestLLMEmbeddings(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/embeddings" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["model"] != "text-embedding-3-small" {
			t.Fatalf("unexpected model: %v", body["model"])
		}

		writeJSON(w, 200, sa.EmbeddingsResponse{
			Object: "list",
			Model:  "text-embedding-3-small",
			Data: []sa.EmbeddingObject{
				{Object: "embedding", Index: 0, Embedding: []float64{0.1, 0.2}},
			},
		})
	})

	raw, err := client.LLM.Embeddings(context.Background(), sa.JSONMap{
		"model": "text-embedding-3-small",
		"input": "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := sa.Decode[sa.EmbeddingsResponse](raw)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestLLMListModels(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		writeJSON(w, 200, sa.LLMModelListResponse{
			Object: "list",
			Data: []sa.LLMModel{
				{ID: "gpt-4o-mini", Object: "model", OwnedBy: "seaart"},
			},
		})
	})

	raw, err := client.LLM.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := sa.Decode[sa.LLMModelListResponse](raw)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != "gpt-4o-mini" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestLLMChatCompletionsCustomHeaders(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Trace-Id"); got != "trace-123" {
			t.Fatalf("unexpected X-Trace-Id header: %q", got)
		}
		if got := r.Header.Get("X-Tenant-Id"); got != "tenant-a" {
			t.Fatalf("unexpected X-Tenant-Id header: %q", got)
		}
		writeJSON(w, 200, sa.ChatCompletionResponse{
			ID:    "chat_123",
			Model: "gpt-4o-mini",
			Choices: []sa.ChatCompletionChoice{
				{
					Index: 0,
					Message: &sa.LLMMessage{
						Role:    "assistant",
						Content: "hello",
					},
					FinishReason: "stop",
				},
			},
		})
	})

	_, err := client.LLM.ChatCompletions(
		context.Background(),
		sa.JSONMap{
			"model":    "gpt-4o-mini",
			"messages": []sa.JSONMap{{"role": "user", "content": "hi"}},
		},
		sa.WithHeader("X-Trace-Id", "trace-123"),
		sa.WithHeaders(http.Header{
			"X-Tenant-Id": []string{"tenant-a"},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLLMChatCompletionsLowercaseWithHeaders(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Request-Id"); got != "1234567890001" {
			t.Fatalf("unexpected X-Request-Id header: %q", got)
		}
		writeJSON(w, 200, sa.ChatCompletionResponse{
			ID:    "chat_123",
			Model: "gpt-4o-mini",
			Choices: []sa.ChatCompletionChoice{
				{
					Index: 0,
					Message: &sa.LLMMessage{
						Role:    "assistant",
						Content: "hello",
					},
					FinishReason: "stop",
				},
			},
		})
	})

	_, err := client.LLM.ChatCompletions(
		context.Background(),
		sa.JSONMap{
			"model":    "gpt-4o-mini",
			"messages": []sa.JSONMap{{"role": "user", "content": "hi"}},
		},
		sa.WithHeaders(http.Header{
			"x-request-id": {"1234567890001"},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLLMListModelsCustomHeaders(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Region"); got != "cn" {
			t.Fatalf("unexpected X-Region header: %q", got)
		}
		writeJSON(w, 200, sa.LLMModelListResponse{
			Object: "list",
			Data:   []sa.LLMModel{{ID: "gpt-4o-mini", Object: "model"}},
		})
	})

	_, err := client.LLM.ListModels(context.Background(), sa.WithHeader("X-Region", "cn"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLLMUsesDedicatedBaseURL(t *testing.T) {
	multimodalCalled := false
	multimodalSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		multimodalCalled = true
		t.Fatalf("LLM request should not hit multimodal server: %s %s", r.Method, r.URL.Path)
	}))
	defer multimodalSrv.Close()

	llmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected llm path: %s", r.URL.Path)
		}
		writeJSON(w, 200, sa.LLMModelListResponse{
			Object: "list",
			Data:   []sa.LLMModel{{ID: "claude-3-5-sonnet", Object: "model"}},
		})
	}))
	defer llmSrv.Close()

	client, err := sa.New(
		&sa.ClientConfig{
			APIKey:       "test-key",
			ModelBaseURL: multimodalSrv.URL,
			LLMBaseURL:   llmSrv.URL,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := client.LLM.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := sa.Decode[sa.LLMModelListResponse](raw)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if multimodalCalled {
		t.Fatal("llm request unexpectedly hit multimodal base URL")
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != "claude-3-5-sonnet" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestLLMErrorClassification(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 401, map[string]any{
			"error": map[string]any{
				"message": "invalid api key",
			},
		})
	})

	_, err := client.LLM.ListModels(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	sdkErr, ok := err.(*sa.Error)
	if !ok {
		t.Fatalf("expected *sa.Error, got %T", err)
	}
	if sdkErr.Kind != sa.ErrAuth {
		t.Fatalf("expected ErrAuth, got %s", sdkErr.Kind)
	}
}

func TestLLMChatCompletionsStream(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["stream"] != true {
			t.Fatalf("expected stream=true, got %v", body["stream"])
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}

		fmt.Fprint(w, "event: message\n")
		fmt.Fprint(w, "data: {\"id\":\"chatcmpl_1\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"hello\"}}]}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	})

	ch, err := client.LLM.ChatCompletionsStream(context.Background(), sa.JSONMap{
		"model":    "gpt-4o-mini",
		"messages": []sa.JSONMap{{"role": "user", "content": "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first, ok := <-ch
	if !ok {
		t.Fatal("expected first stream event")
	}
	if first.Err != nil {
		t.Fatalf("unexpected stream error: %v", first.Err)
	}
	if first.Event != "message" {
		t.Fatalf("unexpected event name: %s", first.Event)
	}

	resp, err := sa.Decode[sa.ChatCompletionResponse](first.Data)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if len(resp.Choices) != 1 || resp.Choices[0].Delta == nil || resp.Choices[0].Delta.Content != "hello" {
		t.Fatalf("unexpected stream response: %+v", resp)
	}

	done, ok := <-ch
	if !ok {
		t.Fatal("expected done stream event")
	}
	if !done.Done {
		t.Fatalf("expected done event, got %+v", done)
	}
}

func TestLLMChatCompletionsStreamCustomHeaders(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Trace-Id"); got != "stream-123" {
			t.Fatalf("unexpected X-Trace-Id header: %q", got)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}

		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	})

	ch, err := client.LLM.ChatCompletionsStream(
		context.Background(),
		sa.JSONMap{
			"model":    "gpt-4o-mini",
			"messages": []sa.JSONMap{{"role": "user", "content": "hi"}},
		},
		sa.WithHeader("X-Trace-Id", "stream-123"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	done, ok := <-ch
	if !ok {
		t.Fatal("expected done stream event")
	}
	if !done.Done {
		t.Fatalf("expected done event, got %+v", done)
	}
}

func TestLLMMessagesStreamChunkParsing(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/messages" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}

		fmt.Fprint(w, "data: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_1\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"claude-3-5-sonnet\",\"usage\":{\"input_tokens\":7}}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"hello\"}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":5}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	})

	ch, err := client.LLM.MessagesStream(context.Background(), sa.JSONMap{
		"model":      "claude-3-5-sonnet",
		"messages":   []sa.JSONMap{{"role": "user", "content": "hi"}},
		"max_tokens": 32,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var text sa.MessagesStreamTextAssembler
	var sawMessageStart bool
	var sawMessageStop bool

	for event := range ch {
		if event.Err != nil {
			t.Fatalf("unexpected stream event error: %v", event.Err)
		}
		if event.Done {
			break
		}

		chunk, err := sa.Decode[sa.MessagesStreamChunk](event.Data)
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		switch chunk.Type {
		case "message_start":
			sawMessageStart = true
			if chunk.Message == nil || chunk.Message.Role != "assistant" {
				t.Fatalf("unexpected message_start chunk: %+v", chunk)
			}
		case "content_block_delta":
			text.Add(chunk)
		case "message_delta":
			if chunk.Delta == nil || chunk.Delta.StopReason != "end_turn" {
				t.Fatalf("unexpected message_delta chunk: %+v", chunk)
			}
			if chunk.Usage == nil || chunk.Usage.OutputTokens != 5 {
				t.Fatalf("unexpected message_delta usage: %+v", chunk)
			}
		case "message_stop":
			sawMessageStop = true
		}
	}

	if !sawMessageStart || !sawMessageStop {
		t.Fatalf("unexpected stream lifecycle: start=%v stop=%v", sawMessageStart, sawMessageStop)
	}
	if text.Text() != "hello" {
		t.Fatalf("unexpected assembled text: %q", text.Text())
	}
}

func TestLLMResponsesStreamChunkParsing(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/responses" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}

		fmt.Fprint(w, "data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_1\",\"object\":\"response\",\"model\":\"gpt-4.1-mini\",\"status\":\"in_progress\",\"output\":[]}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"type\":\"response.output_item.added\",\"output_index\":0,\"item\":{\"id\":\"msg_1\",\"type\":\"message\",\"status\":\"in_progress\",\"role\":\"assistant\",\"content\":[]}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"type\":\"response.content_part.added\",\"item_id\":\"msg_1\",\"output_index\":0,\"content_index\":0,\"part\":{\"type\":\"output_text\",\"text\":\"\",\"annotations\":[]}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"item_id\":\"msg_1\",\"output_index\":0,\"content_index\":0,\"delta\":\"hello\"}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",\"item_id\":\"msg_1\",\"output_index\":0,\"content_index\":0,\"delta\":\" world\"}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\",\"object\":\"response\",\"model\":\"gpt-4.1-mini\",\"status\":\"completed\",\"output\":[],\"usage\":{\"input_tokens\":7,\"output_tokens\":2,\"total_tokens\":9}}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	})

	ch, err := client.LLM.ResponsesStream(context.Background(), sa.JSONMap{
		"model": "gpt-4.1-mini",
		"input": "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var text sa.ResponsesStreamTextAssembler
	var sawCreated bool
	var sawCompleted bool

	for event := range ch {
		if event.Err != nil {
			t.Fatalf("unexpected stream event error: %v", event.Err)
		}
		if event.Done {
			break
		}

		chunk, err := sa.Decode[sa.ResponsesResponseStreamChunk](event.Data)
		if err != nil {
			t.Fatalf("unexpected decode error: %v", err)
		}

		switch chunk.Type {
		case "response.created":
			sawCreated = true
			if chunk.Response == nil || chunk.Response.ID != "resp_1" {
				t.Fatalf("unexpected response.created chunk: %+v", chunk)
			}
		case "response.output_text.delta":
			text.Add(chunk)
		case "response.completed":
			sawCompleted = true
			if chunk.Response == nil || chunk.Response.Status != "completed" {
				t.Fatalf("unexpected response.completed chunk: %+v", chunk)
			}
			if chunk.Response.Usage == nil || chunk.Response.Usage.TotalTokens != 9 {
				t.Fatalf("unexpected response.completed usage: %+v", chunk)
			}
		}
	}

	if !sawCreated || !sawCompleted {
		t.Fatalf("unexpected stream lifecycle: created=%v completed=%v", sawCreated, sawCompleted)
	}
	if text.Text() != "hello world" {
		t.Fatalf("unexpected assembled text: %q", text.Text())
	}
}

func TestLLMStreamingErrorsPointToStreamMethods(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("streaming request should fail before reaching the server")
	})

	testCases := []struct {
		name       string
		call       func() error
		wantMethod string
	}{
		{
			name: "chat",
			call: func() error {
				_, err := client.LLM.ChatCompletions(context.Background(), sa.JSONMap{
					"model":    "gpt-4o-mini",
					"messages": []sa.JSONMap{{"role": "user", "content": "hi"}},
					"stream":   true,
				})
				return err
			},
			wantMethod: "ChatCompletionsStream",
		},
		{
			name: "messages",
			call: func() error {
				_, err := client.LLM.Messages(context.Background(), sa.JSONMap{
					"model":      "claude-3-5-sonnet",
					"messages":   []sa.JSONMap{{"role": "user", "content": "hi"}},
					"max_tokens": 32,
					"stream":     true,
				})
				return err
			},
			wantMethod: "MessagesStream",
		},
		{
			name: "responses",
			call: func() error {
				_, err := client.LLM.Responses(context.Background(), sa.JSONMap{
					"model":  "gpt-4.1-mini",
					"input":  "hello",
					"stream": true,
				})
				return err
			},
			wantMethod: "ResponsesStream",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantMethod) {
				t.Fatalf("unexpected error message: %v", err)
			}
		})
	}
}

func TestLLmChat(t *testing.T) {
	client := liveLLMClient(t)

	raw, err := client.LLM.ChatCompletions(context.Background(), sa.JSONMap{
		"model":    "gpt-5",
		"messages": []sa.JSONMap{{"role": "user", "content": "你好"}},
		// "stream":   true,
	})
	if err != nil {
		log.Fatal("unexpected chat completions error: ", err)
	}
	resp, err := sa.Decode[sa.ChatCompletionResponse](raw)
	if err != nil {
		log.Fatal("unexpected chat completions decode error: ", err)
	}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(">>>>> response: ", string(respJSON))
}

func TestLLmChatStream(t *testing.T) {
	client := liveLLMClient(t)

	stream, err := client.LLM.ChatCompletionsStream(context.Background(), sa.JSONMap{
		"model":    "gpt-5",
		"messages": []sa.JSONMap{{"role": "user", "content": "你好"}},
	})
	if err != nil {
		log.Fatal("unexpected chat completions stream error: ", err)
	}

	for event := range stream {
		if event.Err != nil {
			log.Fatal("unexpected stream event error: ", event.Err)
		}
		if event.Done {
			break
		}

		resp, err := sa.Decode[sa.ChatCompletionResponse](event.Data)
		if err != nil {
			log.Fatal("unexpected chat completions decode error: ", err)
		}
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(">>>>> response: ", string(respJSON))
	}
}

func TestLLmMessages(t *testing.T) {
	client := liveLLMClient(t)

	raw, err := client.LLM.Messages(context.Background(), sa.JSONMap{
		"model":      "gpt-5",
		"messages":   []sa.JSONMap{{"role": "user", "content": "你好"}},
		"max_tokens": 1024,
	})
	if err != nil {
		log.Fatal("unexpected messages error: ", err)
	}
	resp, err := sa.Decode[sa.MessagesResponse](raw)
	if err != nil {
		log.Fatal("unexpected messages decode error: ", err)
	}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(">>>>> response: ", string(respJSON))
}

func TestLLmMessagesStream(t *testing.T) {
	client := liveLLMClient(t)

	stream, err := client.LLM.MessagesStream(context.Background(), sa.JSONMap{
		"model":      "gpt-5",
		"max_tokens": 1024,
		"messages":   []sa.JSONMap{{"role": "user", "content": "你好"}},
	})
	if err != nil {
		log.Fatal("unexpected messages stream error: ", err)
	}

	for event := range stream {
		if event.Err != nil {
			log.Fatal("unexpected stream event error: ", event.Err)
		}
		if event.Done {
			break
		}

		resp, err := sa.Decode[sa.MessagesStreamChunk](event.Data)
		if err != nil {
			log.Fatal("unexpected messages decode error: ", err)
		}
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(">>>>> response: ", string(respJSON))
	}
}

func TestLLmResponses(t *testing.T) {
	client := liveLLMClient(t)

	raw, err := client.LLM.Responses(context.Background(), sa.JSONMap{
		"model": "gemini-2.0-flash",
		"input": "你好",
	})
	if err != nil {
		log.Fatal("unexpected responses error: ", err)
	}
	resp, err := sa.Decode[sa.ResponsesResponse](raw)
	if err != nil {
		log.Fatal("unexpected responses decode error: ", err)
	}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(">>>>> response: ", string(respJSON))
}

func TestLLmResponsesStream(t *testing.T) {
	client := liveLLMClient(t)

	stream, err := client.LLM.ResponsesStream(context.Background(), sa.JSONMap{
		"model": "gemini-2.0-flash",
		"input": "你好",
	})
	if err != nil {
		log.Fatal("unexpected responses stream error: ", err)
	}

	for event := range stream {
		if event.Err != nil {
			log.Fatal("unexpected stream event error: ", event.Err)
		}
		if event.Done {
			break
		}

		resp, err := sa.Decode[sa.ResponsesResponseStreamChunk](event.Data)
		if err != nil {
			log.Fatal("unexpected responses decode error: ", err)
		}
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(">>>>> response: ", string(respJSON))
	}
}

func TestLLmRerank(t *testing.T) {
	client := liveLLMClient(t)

	raw, err := client.LLM.Rerank(context.Background(), sa.JSONMap{
		"model":     "qwen3-rerank",
		"query":     "What is the capital of France?",
		"documents": []string{"Paris is the capital of France.", "Berlin is the capital of Germany.", "Madrid is the capital of Spain."},
	})
	if err != nil {
		log.Fatal("unexpected rerank error: ", err)
	}
	resp, err := sa.Decode[sa.RerankResponse](raw)
	if err != nil {
		log.Fatal("unexpected rerank decode error: ", err)
	}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(">>>>> response: ", string(respJSON))
}

func TestLLmEmbeddings(t *testing.T) {
	client := liveLLMClient(t)

	raw, err := client.LLM.Embeddings(context.Background(), sa.JSONMap{
		"model": "text-embedding-3-large",
		"input": "Hello world",
	})
	if err != nil {
		log.Fatal("unexpected embeddings error: ", err)
	}
	resp, err := sa.Decode[sa.EmbeddingsResponse](raw)
	if err != nil {
		log.Fatal("unexpected embeddings decode error: ", err)
	}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(">>>>> response: ", string(respJSON))
}

func TestLLmListModels(t *testing.T) {
	client := liveLLMClient(t)

	raw, err := client.LLM.ListModels(context.Background())
	if err != nil {
		log.Fatal("unexpected list models error: ", err)
	}
	resp, err := sa.Decode[sa.LLMModelListResponse](raw)
	if err != nil {
		log.Fatal("unexpected list models decode error: ", err)
	}
	respJSON, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(">>>>> response: ", string(respJSON))
}
