package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/types"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/shared"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/transport"
)

const (
	// PathGenerationSync is the synchronous delivery endpoint: it submits the
	// task and waits for its terminal result. The response representation is
	// chosen with the Accept header.
	PathGenerationSync = "/v1/generation/sync"
	// PathTaskStream is the task subscription endpoint, followed by the task id
	// and "/stream".
	PathTaskStream = "/v1/generation/task/"

	acceptJSON        = "application/json"
	acceptEventStream = "text/event-stream"
	syncTimeoutCode   = "SYNC_TIMEOUT"
)

// CreateSync submits a task and blocks until it reaches a terminal state.
//
// Long-running tasks should not use this: the wait is silent, so an intermediary
// can drop the connection at its idle timeout. Poll with GetTask/WaitTask or use
// CreateStream instead.
func CreateSync(client *transport.Client, ctx context.Context, body any, headers http.Header) (*types.TaskResponse, error) {
	if headers == nil {
		headers = http.Header{}
	}
	headers.Set("Accept", acceptJSON)

	status, payload, err := client.Request(ctx, http.MethodPost, PathGenerationSync, body, headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, syncDeliveryError(status, payload)
	}

	var task types.TaskResponse
	if err := decode(payload, &task); err != nil {
		return nil, err
	}
	if strings.ToLower(task.Status) == StatusFailed {
		return nil, taskFailedError(&task, task.ID)
	}
	return &task, nil
}

// CreateStream submits a task and streams its output as it is produced.
func CreateStream(client *transport.Client, ctx context.Context, body any, headers http.Header) (<-chan types.TaskStreamEvent, error) {
	if headers == nil {
		headers = http.Header{}
	}
	headers.Set("Accept", acceptEventStream)

	resp, err := client.RequestStream(ctx, http.MethodPost, PathGenerationSync, body, headers)
	if err != nil {
		return nil, err
	}
	return readTaskStream(ctx, resp)
}

// Subscribe streams an existing task's incremental output, starting after cursor.
//
// It works for running tasks, finished tasks (chunks replay from cursor) and
// tasks created elsewhere, which makes a dropped stream resumable without
// resubmitting the work.
func Subscribe(client *transport.Client, ctx context.Context, taskID string, cursor int, headers http.Header) (<-chan types.TaskStreamEvent, error) {
	trimmed := strings.TrimSpace(taskID)
	if trimmed == "" {
		return nil, &shared.Error{Kind: shared.ErrGeneral, Message: "task_id is required"}
	}
	if headers == nil {
		headers = http.Header{}
	}
	headers.Set("Accept", acceptEventStream)

	path := PathTaskStream + url.PathEscape(trimmed) + "/stream"
	if cursor > 0 {
		path += "?cursor=" + strconv.Itoa(cursor)
	}

	resp, err := client.RequestStream(ctx, http.MethodGet, path, nil, headers)
	if err != nil {
		return nil, err
	}
	return readTaskStream(ctx, resp)
}

// readTaskStream parses the gateway's generation SSE stream into task events.
func readTaskStream(ctx context.Context, resp *http.Response) (<-chan types.TaskStreamEvent, error) {
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		payload, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, &shared.Error{
				Kind:    shared.ErrGeneral,
				Message: "failed to read stream error response: " + readErr.Error(),
			}
		}
		return nil, syncDeliveryError(resp.StatusCode, payload)
	}

	ch := make(chan types.TaskStreamEvent, 8)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		eventName := ""
		dataLines := make([]string, 0, 4)
		terminal := false

		emit := func() bool {
			if len(dataLines) == 0 && eventName == "" {
				return true
			}

			data := strings.Join(dataLines, "\n")
			name := eventName
			eventName = ""
			dataLines = dataLines[:0]

			if data == "" || data == "[DONE]" {
				terminal = true
				return sendTaskStreamEvent(ctx, ch, types.TaskStreamEvent{Event: "done", Done: true})
			}

			event := parseTaskStreamEvent(name, []byte(data))
			if event.Done {
				terminal = true
			}
			return sendTaskStreamEvent(ctx, ch, event)
		}

		// A stream that ends without a terminal event is a truncated delivery: the
		// caller must not treat the partial result as success.
		finish := func() {
			if !terminal {
				sendTaskStreamEvent(ctx, ch, types.TaskStreamEvent{
					Done: true,
					Err: &shared.Error{
						Kind:    shared.ErrNetwork,
						Message: "stream ended before a terminal event; resume with Subscribe(taskID, cursor)",
					},
				})
			}
		}

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					emit()
					finish()
					return
				}

				sendTaskStreamEvent(ctx, ch, types.TaskStreamEvent{
					Done: true,
					Err:  &shared.Error{Kind: shared.ErrNetwork, Message: "stream read failed: " + err.Error()},
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
				continue // keepalive comment sent while the task runs
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

func parseTaskStreamEvent(eventName string, data []byte) types.TaskStreamEvent {
	if eventName == "" {
		eventName = "output"
	}

	event := types.TaskStreamEvent{Event: eventName}
	switch eventName {
	case "done":
		var task types.TaskResponse
		if err := json.Unmarshal(data, &task); err != nil {
			return decodeFrameError(eventName, err)
		}
		event.Task = &task
		event.TaskID = task.ID
		event.Status = task.Status
		event.Done = true
	case "error":
		var frame struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Error  struct {
				Code         string `json:"code"`
				Message      string `json:"message"`
				ErrorMessage string `json:"error_message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &frame); err != nil {
			return decodeFrameError(eventName, err)
		}
		event.ErrorCode = frame.Error.Code
		// The gateway uses message on this endpoint, but error_message also appears
		// on gateway error payloads: keep whichever is present so the failure reason
		// is never dropped.
		event.ErrorMessage = frame.Error.Message
		if event.ErrorMessage == "" {
			event.ErrorMessage = frame.Error.ErrorMessage
		}
		event.TaskID = frame.ID
		event.Status = frame.Status
		event.Done = true
	default:
		var frame types.TaskStreamFrame
		if err := json.Unmarshal(data, &frame); err != nil {
			return decodeFrameError(eventName, err)
		}
		event.Cursor = frame.Cursor
		event.Chunks = frame.Output
		event.TaskID = frame.ID
		event.Status = frame.Status
	}
	return event
}

// decodeFrameError surfaces a malformed frame instead of turning it into an empty
// event that hides the failure reason.
func decodeFrameError(eventName string, err error) types.TaskStreamEvent {
	return types.TaskStreamEvent{
		Event: eventName,
		Err:   &shared.Error{Kind: shared.ErrGeneral, Message: "failed to decode stream frame: " + err.Error()},
	}
}

func sendTaskStreamEvent(ctx context.Context, ch chan<- types.TaskStreamEvent, event types.TaskStreamEvent) bool {
	select {
	case ch <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

// syncDeliveryError turns a failed synchronous delivery into an SDK error.
//
// It starts from the generic HTTP error, so the status keeps driving the kind
// (429 -> quota, 504 -> timeout), and then keeps what only this endpoint knows:
// the gateway error code and the task id that callers need to resume the task.
func syncDeliveryError(status int, payload []byte) error {
	base := httpError(status, payload)

	sdkErr, ok := base.(*shared.Error)
	if !ok {
		return base
	}

	var body struct {
		ID    string `json:"id"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(payload, &body)

	if body.Error.Code == syncTimeoutCode {
		sdkErr.Kind = shared.ErrTimeout
	}
	if body.ID != "" {
		sdkErr.TaskID = body.ID
	}
	return sdkErr
}

// taskFailedError mirrors the failure reported by WaitTask so both deliveries
// describe a failed task the same way.
func taskFailedError(task *types.TaskResponse, taskID string) error {
	message := "task failed"
	code := 0
	if task != nil && task.Error != nil {
		detail := task.Error.ErrorMessage
		if detail == "" {
			detail = task.Error.Message
		}
		if detail != "" {
			message = "task failed: " + detail
		}
		code = task.Error.Code
	}
	return &shared.Error{Kind: shared.ErrTaskFailed, Message: message, TaskID: taskID, Code: code}
}
