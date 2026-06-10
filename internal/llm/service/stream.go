package service

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	llmtypes "github.com/SeaArt-Infra/sea-sdk-go/internal/llm/types"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/shared"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/transport"
)

func ChatCompletionsStream(client *transport.Client, ctx context.Context, payload llmtypes.JSONMap, headers http.Header) (<-chan llmtypes.StreamEvent, error) {
	return doSSE(client, ctx, http.MethodPost, PathChatCompletions, ensureStreamingPayload(payload), headers)
}

func MessagesStream(client *transport.Client, ctx context.Context, payload llmtypes.JSONMap, headers http.Header) (<-chan llmtypes.StreamEvent, error) {
	return doSSE(client, ctx, http.MethodPost, PathMessages, ensureStreamingPayload(payload), headers)
}

func ResponsesStream(client *transport.Client, ctx context.Context, payload llmtypes.JSONMap, headers http.Header) (<-chan llmtypes.StreamEvent, error) {
	return doSSE(client, ctx, http.MethodPost, PathResponses, ensureStreamingPayload(payload), headers)
}

func doSSE(client *transport.Client, ctx context.Context, method, path string, body llmtypes.JSONMap, headers http.Header) (<-chan llmtypes.StreamEvent, error) {
	resp, err := client.RequestStream(ctx, method, path, body, headers)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		payload, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, &shared.Error{
				Kind:    shared.ErrGeneral,
				Message: "failed to read stream error response: " + readErr.Error(),
			}
		}
		return nil, httpError(resp.StatusCode, payload)
	}

	ch := make(chan llmtypes.StreamEvent, 8)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		eventName := ""
		dataLines := make([]string, 0, 4)

		emit := func() bool {
			if len(dataLines) == 0 && eventName == "" {
				return true
			}

			data := strings.Join(dataLines, "\n")
			event := llmtypes.StreamEvent{Event: eventName}
			switch data {
			case "":
			case "[DONE]":
				event.Done = true
			default:
				event.Data = llmtypes.RawResponse(data)
			}

			eventName = ""
			dataLines = dataLines[:0]

			if event.Done || len(event.Data) > 0 || event.Event != "" {
				return sendStreamEvent(ctx, ch, event)
			}
			return true
		}

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					emit()
					return
				}

				sendStreamEvent(ctx, ch, llmtypes.StreamEvent{
					Err: &shared.Error{Kind: shared.ErrNetwork, Message: "stream read failed: " + err.Error()},
				})
				return
			}

			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				if !emit() {
					return
				}
				continue
			}
			if strings.HasPrefix(line, ":") {
				continue
			}

			switch {
			case strings.HasPrefix(line, "event:"):
				eventName = strings.TrimSpace(line[len("event:"):])
			case strings.HasPrefix(line, "data:"):
				dataLines = append(dataLines, strings.TrimSpace(line[len("data:"):]))
			}
		}
	}()

	return ch, nil
}

func ensureStreamingPayload(payload llmtypes.JSONMap) llmtypes.JSONMap {
	cloned := make(llmtypes.JSONMap, len(payload)+1)
	for key, value := range payload {
		cloned[key] = value
	}
	cloned["stream"] = true
	return cloned
}

func sendStreamEvent(ctx context.Context, ch chan<- llmtypes.StreamEvent, event llmtypes.StreamEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- event:
		return true
	}
}
