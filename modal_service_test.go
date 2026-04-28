package sa_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	sa "github.com/seaart/sa-go"
)

func TestMediaCreate_SubmitsRawBody(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/generation" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		body := extractBody(t, r)
		if body["model"] != "vidu_q3_reference" {
			t.Fatalf("unexpected model: %v", body["model"])
		}

		params := body["parameters"].(map[string]any)
		if params["duration"] != float64(5) {
			t.Fatalf("unexpected duration: %v", params["duration"])
		}

		writeJSON(w, 200, map[string]any{
			"id":     "task_create",
			"status": "in_progress",
			"model":  "vidu_q3_reference",
		})
	})

	task, err := client.Modal.Create(context.Background(), sa.JSONMap{
		"model": "vidu_q3_reference",
		"input": []map[string]any{
			{
				"type": "message",
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": "cinematic shot"},
					{"type": "image_url", "url": "https://example.com/ref1.webp"},
				},
			},
		},
		"parameters": map[string]any{
			"duration": 5,
		},
	}, sa.WithHeader("X-Trace-Id", "trace-123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != "task_create" {
		t.Fatalf("unexpected task id: %s", task.ID)
	}
	if task.Status != "in_progress" {
		t.Fatalf("unexpected status: %s", task.Status)
	}
	if task.Model != "vidu_q3_reference" {
		t.Fatalf("unexpected model: %s", task.Model)
	}
}

func TestMediaGet_ReturnsTask(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/generation/task/task_abc123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		writeJSON(w, 200, map[string]any{
			"id":       "task_abc123",
			"status":   "completed",
			"progress": 1.0,
			"model":    "vidu_q3_reference",
			"output": []map[string]any{
				{
					"content": []map[string]any{
						{"type": "video", "url": "https://example.com/out.mp4"},
					},
				},
			},
		})
	})

	task, err := client.Modal.Get(context.Background(), "task_abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != "task_abc123" {
		t.Fatalf("unexpected task id: %s", task.ID)
	}
	if task.Status != "completed" {
		t.Fatalf("unexpected status: %s", task.Status)
	}
	if task.Progress != 1.0 {
		t.Fatalf("unexpected progress: %v", task.Progress)
	}
	if len(task.Output) != 1 {
		t.Fatalf("unexpected output count: %d", len(task.Output))
	}
}

func TestMediaWait_Completes(t *testing.T) {
	var polls atomic.Int32

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/generation/task/task_wait":
			n := polls.Add(1)
			if n == 1 {
				writeJSON(w, 200, map[string]any{
					"id":       "task_wait",
					"status":   "in_progress",
					"progress": 0.4,
					"model":    "vidu_q3_reference",
				})
				return
			}
			writeJSON(w, 200, map[string]any{
				"id":       "task_wait",
				"status":   "completed",
				"progress": 1.0,
				"model":    "vidu_q3_reference",
				"output": []map[string]any{
					{
						"content": []map[string]any{
							{"type": "video", "url": "https://example.com/out.mp4"},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	task, err := client.Modal.Wait(context.Background(), "task_wait",
		sa.WithPollInterval(10*time.Millisecond),
		sa.WithPollTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status != "completed" {
		t.Fatalf("unexpected status: %s", task.Status)
	}
	if polls.Load() != 2 {
		t.Fatalf("unexpected poll count: %d", polls.Load())
	}
}

func TestTaskWait_UsesAttachedClient(t *testing.T) {
	var polls atomic.Int32

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/generation":
			writeJSON(w, 200, map[string]any{
				"id":     "task_attached",
				"status": "in_progress",
				"model":  "vidu_q3_reference",
			})
		case "/v1/generation/task/task_attached":
			polls.Add(1)
			writeJSON(w, 200, map[string]any{
				"id":       "task_attached",
				"status":   "completed",
				"progress": 1.0,
				"model":    "vidu_q3_reference",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	task, err := client.Modal.Create(context.Background(), sa.JSONMap{
		"model": "vidu_q3_reference",
	})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	task, err = task.Wait(context.Background(),
		sa.WithPollInterval(10*time.Millisecond),
		sa.WithPollTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("unexpected wait error: %v", err)
	}
	if task.Status != "completed" {
		t.Fatalf("unexpected status: %s", task.Status)
	}
	if polls.Load() != 1 {
		t.Fatalf("unexpected poll count: %d", polls.Load())
	}
}

func TestMediaWait_FailedTask(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"id":     "task_fail",
			"status": "failed",
			"error": map[string]any{
				"error_message": "provider rejected request",
			},
		})
	})

	_, err := client.Modal.Wait(context.Background(), "task_fail",
		sa.WithPollInterval(10*time.Millisecond),
		sa.WithPollTimeout(2*time.Second),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	sdkErr, ok := err.(*sa.Error)
	if !ok {
		t.Fatalf("expected *sa.Error, got %T", err)
	}
	if sdkErr.Kind != sa.ErrTaskFailed {
		t.Fatalf("unexpected error kind: %s", sdkErr.Kind)
	}
}

func TestTaskBuilderBuildsGenericRequest(t *testing.T) {
	body := sa.NewTask("vidu_q3_reference").
		User(
			sa.Text("cinematic shot"),
			sa.ImageURL("https://example.com/ref1.webp"),
		).
		Param("duration", 5).
		Metadata("trace_id", "trace-123").
		Build()

	if body["model"] != "vidu_q3_reference" {
		t.Fatalf("unexpected model: %v", body["model"])
	}
	if body["metadata"].(map[string]any)["trace_id"] != "trace-123" {
		t.Fatalf("unexpected metadata: %v", body["metadata"])
	}
}
