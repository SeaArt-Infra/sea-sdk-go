package sa_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	sa "github.com/SeaArt-Infra/sea-sdk-go"
)

// writeSSE streams the given frames as server-sent events.
func writeSSE(t *testing.T, w http.ResponseWriter, frames ...string) {
	t.Helper()

	flusher, ok := w.(http.Flusher)
	if !ok {
		t.Fatal("response writer does not support flushing")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	for _, frame := range frames {
		fmt.Fprint(w, frame)
		flusher.Flush()
	}
}

func TestCreateSync_ReturnsFinalTask(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/generation/sync" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("unexpected Accept header: %q", got)
		}
		if got := r.Header.Get("X-Model"); got != "microsoft_gpt_image_2_5_flare" {
			t.Fatalf("unexpected X-Model header: %q", got)
		}

		writeJSON(w, 200, map[string]any{
			"id": "task_sync_1", "status": "completed", "model": "microsoft_gpt_image_2_5_flare",
			"output":   []any{map[string]any{"content": []any{map[string]any{"type": "image", "url": "https://cdn.example.com/out.webp"}}}},
			"usage":    map[string]any{"cost": "0.0065", "discount": 1},
			"metadata": map[string]any{"completed_at": 4.2},
		})
	})

	task, err := client.Modal.CreateSync(context.Background(), sa.JSONMap{
		"model": "microsoft_gpt_image_2_5_flare",
		"input": []any{map[string]any{"params": map[string]any{"prompt": "a dog is running"}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != "task_sync_1" || task.Status != "completed" {
		t.Fatalf("unexpected task: %#v", task)
	}
	if len(task.Output) != 1 || task.Output[0].Content[0].URL != "https://cdn.example.com/out.webp" {
		t.Fatalf("unexpected output: %#v", task.Output)
	}
	if task.Usage == nil || task.Usage.Cost.String() != "0.0065" {
		t.Fatalf("unexpected usage: %#v", task.Usage)
	}
}

func TestCreateSync_FailedTaskReportsTaskFailed(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"id": "task_sync_failed", "status": "failed", "model": "m",
			"error": map[string]any{"code": 110001, "message": "vendor rejected"},
		})
	})

	_, err := client.Modal.CreateSync(context.Background(), sa.JSONMap{"model": "m"})
	var sdkErr *sa.Error
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected *sa.Error, got %T (%v)", err, err)
	}
	if sdkErr.Kind != sa.ErrTaskFailed || sdkErr.TaskID != "task_sync_failed" {
		t.Fatalf("unexpected error: %#v", sdkErr)
	}
	if !strings.Contains(sdkErr.Message, "vendor rejected") {
		t.Fatalf("unexpected message: %s", sdkErr.Message)
	}
}

func TestCreateSync_TimeoutKeepsTaskID(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 504, map[string]any{
			"id": "task_slow", "status": "in_progress",
			"error": map[string]any{"code": "SYNC_TIMEOUT", "message": "waited 15m0s, still running"},
		})
	})

	_, err := client.Modal.CreateSync(context.Background(), sa.JSONMap{"model": "m"})
	var sdkErr *sa.Error
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected *sa.Error, got %T (%v)", err, err)
	}
	if sdkErr.Kind != sa.ErrTimeout {
		t.Fatalf("unexpected kind: %s", sdkErr.Kind)
	}
	if sdkErr.TaskID != "task_slow" {
		t.Fatalf("timeout must keep the task id, got %q", sdkErr.TaskID)
	}
}

func TestCreateStream_YieldsChunksThenDone(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/generation/sync" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Fatalf("unexpected Accept header: %q", got)
		}
		writeSSE(t, w,
			"event: output\ndata: {\"id\":\"task_s\",\"model\":\"m\",\"status\":\"in_progress\",\"output\":[{\"content\":[{\"type\":\"audio\",\"url\":\"https://cdn.example.com/0.wav\",\"chunk_index\":0}]}],\"cursor\":1}\n\n",
			": keepalive\n\n",
			"event: output\ndata: {\"id\":\"task_s\",\"model\":\"m\",\"status\":\"in_progress\",\"output\":[{\"content\":[{\"type\":\"audio\",\"url\":\"https://cdn.example.com/1.wav\",\"chunk_index\":1}]}],\"cursor\":3}\n\n",
			"event: done\ndata: {\"id\":\"task_s\",\"status\":\"completed\",\"model\":\"m\",\"output\":[{\"content\":[{\"type\":\"audio\",\"url\":\"https://cdn.example.com/full.wav\"}]}],\"usage\":{\"cost\":\"0.0017\"}}\n\n",
		)
	})

	events, err := client.Modal.CreateStream(context.Background(), sa.JSONMap{"model": "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var collected []sa.TaskStreamEvent
	for event := range events {
		collected = append(collected, event)
	}
	if len(collected) != 3 {
		t.Fatalf("expected 3 events (keepalive ignored), got %d: %#v", len(collected), collected)
	}

	first := collected[0]
	if first.Event != "output" || first.Cursor != 1 || first.TaskID != "task_s" {
		t.Fatalf("unexpected first event: %#v", first)
	}
	if got := first.Chunks[0].Content[0].URL; got != "https://cdn.example.com/0.wav" {
		t.Fatalf("unexpected chunk url: %s", got)
	}
	if idx := first.Chunks[0].Content[0].ChunkIndex; idx == nil || *idx != 0 {
		t.Fatalf("expected chunk_index 0, got %v", idx)
	}
	if first.Done {
		t.Fatal("an output frame must not be terminal")
	}

	last := collected[len(collected)-1]
	if !last.Done || last.Event != "done" {
		t.Fatalf("unexpected terminal event: %#v", last)
	}
	if last.Task == nil || last.Task.Status != "completed" {
		t.Fatalf("done event must carry the final task: %#v", last.Task)
	}
	if got := last.Task.Output[0].Content[0].URL; got != "https://cdn.example.com/full.wav" {
		t.Fatalf("unexpected final url: %s", got)
	}
	if last.Task.Usage == nil || last.Task.Usage.Cost.String() != "0.0017" {
		t.Fatalf("done event must carry usage: %#v", last.Task.Usage)
	}
}

func TestCreateStream_HTTPErrorBeforeStreaming(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 429, map[string]any{"error": map[string]any{"code": 1002, "message": "rate limited"}})
	})

	_, err := client.Modal.CreateStream(context.Background(), sa.JSONMap{"model": "m"})
	var sdkErr *sa.Error
	if !errors.As(err, &sdkErr) {
		t.Fatalf("expected *sa.Error, got %T (%v)", err, err)
	}
	if sdkErr.Kind != sa.ErrQuota || sdkErr.Status != 429 {
		t.Fatalf("unexpected error: %#v", sdkErr)
	}
}

func TestSubscribe_ResumesFromCursor(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/generation/task/task_resume/stream" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("cursor"); got != "3" {
			t.Fatalf("unexpected cursor: %q", got)
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Fatalf("unexpected Accept header: %q", got)
		}
		writeSSE(t, w,
			"event: output\ndata: {\"id\":\"task_resume\",\"status\":\"in_progress\",\"output\":[{\"content\":[{\"type\":\"audio\",\"url\":\"https://cdn.example.com/3.wav\",\"chunk_index\":3}]}],\"cursor\":4}\n\n",
			"event: done\ndata: {\"id\":\"task_resume\",\"status\":\"completed\",\"output\":[]}\n\n",
		)
	})

	events, err := client.Modal.Subscribe(context.Background(), "task_resume", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var collected []sa.TaskStreamEvent
	for event := range events {
		collected = append(collected, event)
	}
	if len(collected) != 2 {
		t.Fatalf("expected 2 events, got %d", len(collected))
	}
	if collected[0].Cursor != 4 || collected[0].Chunks[0].Content[0].URL != "https://cdn.example.com/3.wav" {
		t.Fatalf("unexpected resume event: %#v", collected[0])
	}
	if !collected[1].Done || collected[1].Task == nil || collected[1].Task.ID != "task_resume" {
		t.Fatalf("unexpected terminal event: %#v", collected[1])
	}
}

func TestSubscribe_RequiresTaskID(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})

	_, err := client.Modal.Subscribe(context.Background(), "   ", 0)
	var sdkErr *sa.Error
	if !errors.As(err, &sdkErr) || sdkErr.Kind != sa.ErrGeneral {
		t.Fatalf("expected a general error, got %v", err)
	}
}

func TestTask_StreamSubscribesToItsOwnTask(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/generation":
			writeJSON(w, 200, map[string]any{"id": "task_bound", "status": "in_progress", "model": "m"})
		case "/v1/generation/task/task_bound/stream":
			// cursor 0 is omitted: the stream starts from the task's first chunk.
			if got := r.URL.Query().Get("cursor"); got != "" {
				t.Fatalf("cursor 0 must be omitted, got %q", got)
			}
			writeSSE(t, w, "event: done\ndata: {\"id\":\"task_bound\",\"status\":\"completed\",\"output\":[]}\n\n")
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	task, err := client.Modal.Create(context.Background(), sa.JSONMap{"model": "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events, err := task.Stream(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var last sa.TaskStreamEvent
	for event := range events {
		last = event
	}
	if !last.Done || last.Task == nil || last.Task.ID != "task_bound" {
		t.Fatalf("unexpected terminal event: %#v", last)
	}
}

func TestCreateStream_FailsWhenTheStreamEndsWithoutATerminalEvent(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Chunks arrive, then the connection is cut: no done/error frame.
		writeSSE(t, w,
			"event: output\ndata: {\"id\":\"task_trunc\",\"status\":\"in_progress\",\"output\":[{\"content\":[{\"type\":\"audio\",\"url\":\"https://cdn.example.com/0.wav\"}]}],\"cursor\":1}\n\n",
		)
	})

	events, err := client.Modal.CreateStream(context.Background(), sa.JSONMap{"model": "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var collected []sa.TaskStreamEvent
	for event := range events {
		collected = append(collected, event)
	}
	if len(collected) != 2 {
		t.Fatalf("expected the chunk plus a truncation error, got %#v", collected)
	}
	if len(collected[0].Chunks) != 1 || collected[0].Status != "in_progress" {
		t.Fatalf("the chunks that did arrive must still be delivered: %#v", collected[0])
	}

	last := collected[1]
	if !last.Done || last.Err == nil {
		t.Fatalf("a truncated stream must end with an error event: %#v", last)
	}
	if !strings.Contains(last.Err.Error(), "terminal event") {
		t.Fatalf("unexpected error: %v", last.Err)
	}
}

func TestCreateStream_SurfacesMalformedFrames(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeSSE(t, w,
			"event: output\ndata: {not json\n\n",
			"event: done\ndata: {\"id\":\"task_bad\",\"status\":\"completed\",\"output\":[]}\n\n",
		)
	})

	events, err := client.Modal.CreateStream(context.Background(), sa.JSONMap{"model": "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var malformed error
	var terminal *sa.TaskStreamEvent
	for event := range events {
		if event.Err != nil {
			malformed = event.Err
		}
		if event.Done && event.Err == nil {
			copy := event
			terminal = &copy
		}
	}

	if malformed == nil || !strings.Contains(malformed.Error(), "decode stream frame") {
		t.Fatalf("a malformed frame must surface its decode error, got %v", malformed)
	}
	if terminal == nil || terminal.Task == nil || terminal.Task.Status != "completed" {
		t.Fatalf("the stream must continue after a malformed frame: %#v", terminal)
	}
}
