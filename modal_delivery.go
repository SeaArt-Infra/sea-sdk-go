package sa

import (
	"context"

	mmservice "github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/service"
	mmtypes "github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/types"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/transport"
)

// TaskStreamEvent is one event of a streamed generation delivery.
//
// Event is one of:
//   - "output": newly produced chunks; Chunks holds them (one frame may carry
//     several, because the gateway batches them) and Cursor is the consumption
//     cursor to pass when resuming.
//   - "done": terminal event; Task holds the complete result, identical to what
//     Get returns once the task is finished.
//   - "error": the delivery failed or timed out after streaming had started; see
//     ErrorCode and ErrorMessage.
//
// Judge the end of the stream by Done (true for "done" and "error"), never by the
// status of a chunk frame: chunk frames always report "in_progress", so stopping
// on status == "completed" would drop the "done" event and lose the result.
type TaskStreamEvent struct {
	Event  string
	TaskID string
	// Status is the task status reported by the frame: always "in_progress" on
	// chunk frames, terminal on the "done" and "error" events.
	Status       string
	Cursor       int
	Chunks       []Output
	Task         *Task
	ErrorCode    string
	ErrorMessage string
	Done         bool
	Err          error
}

// CreateSync submits body and blocks until the task reaches a terminal state,
// returning the final result (usage included).
//
// One call instead of Create + Wait. The route and the response representation
// stay inside the SDK: the caller passes the same body as Create.
//
// Do not use this for tasks that may run longer than ~120 seconds: the wait is
// silent, so a proxy or load balancer can drop the connection at its idle
// timeout. Use Create + Wait for those tasks (each poll is a short request), or
// CreateStream when progress is wanted.
//
// A failed task returns an *Error with Kind ErrTaskFailed, like Wait. If the
// gateway gives up waiting, the returned *Error has Kind ErrTimeout and TaskID
// set, so the caller can keep polling Wait instead of resubmitting the work.
func (m *ModalService) CreateSync(ctx context.Context, body JSONMap, opts ...RequestOption) (*Task, error) {
	requestBody, headers, err := moveModelToHeader(body, buildRequestOptions(opts).headers)
	if err != nil {
		return nil, err
	}
	resp, err := mmservice.CreateSync(m.client, ctx, requestBody, headers)
	if err != nil {
		return nil, err
	}
	return newTaskFromResponse(m.client, resp), nil
}

// CreateStream submits body and streams its output as it is produced.
//
// The returned channel is closed when the stream ends; the terminal "done" event
// carries the complete result, and a read failure is reported as an event with
// Err set and Done true. Streaming holds one long-lived connection, which the
// gateway keeps alive with keepalive comments.
func (m *ModalService) CreateStream(ctx context.Context, body JSONMap, opts ...RequestOption) (<-chan TaskStreamEvent, error) {
	requestBody, headers, err := moveModelToHeader(body, buildRequestOptions(opts).headers)
	if err != nil {
		return nil, err
	}
	events, err := mmservice.CreateStream(m.client, ctx, requestBody, headers)
	if err != nil {
		return nil, err
	}
	return mapTaskStreamEvents(m.client, events), nil
}

// Subscribe streams an existing task's incremental output, starting after cursor.
//
// Pass the Cursor of the last event that was consumed to resume without
// duplicates: this is what makes a dropped stream resumable, and it also works
// for tasks created by someone else.
func (m *ModalService) Subscribe(ctx context.Context, taskID string, cursor int, opts ...RequestOption) (<-chan TaskStreamEvent, error) {
	events, err := mmservice.Subscribe(m.client, ctx, taskID, cursor, buildRequestOptions(opts).headers)
	if err != nil {
		return nil, err
	}
	return mapTaskStreamEvents(m.client, events), nil
}

// Stream subscribes to this task's incremental output starting at cursor.
func (t *Task) Stream(ctx context.Context, cursor int, opts ...RequestOption) (<-chan TaskStreamEvent, error) {
	if t == nil {
		return nil, &Error{Kind: ErrGeneral, Message: "task is nil"}
	}
	if t.client == nil {
		return nil, &Error{Kind: ErrGeneral, Message: "task is detached from client"}
	}
	events, err := mmservice.Subscribe(t.client, ctx, t.ID, cursor, buildRequestOptions(opts).headers)
	if err != nil {
		return nil, err
	}
	return mapTaskStreamEvents(t.client, events), nil
}

// mapTaskStreamEvents converts internal events into the public shape, binding the
// client to the final task so it can keep polling it.
func mapTaskStreamEvents(client *transport.Client, events <-chan mmtypes.TaskStreamEvent) <-chan TaskStreamEvent {
	out := make(chan TaskStreamEvent, cap(events))

	go func() {
		defer close(out)
		for event := range events {
			out <- TaskStreamEvent{
				Event:        event.Event,
				TaskID:       event.TaskID,
				Status:       event.Status,
				Cursor:       event.Cursor,
				Chunks:       event.Chunks,
				Task:         newTaskFromResponse(client, event.Task),
				ErrorCode:    event.ErrorCode,
				ErrorMessage: event.ErrorMessage,
				Done:         event.Done,
				Err:          event.Err,
			}
		}
	}()

	return out
}
