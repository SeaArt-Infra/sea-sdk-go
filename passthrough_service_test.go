package sa_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sa "github.com/seaart/sa-go"
)

func newPassthroughTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *sa.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := sa.New(&sa.ClientConfig{
		APIKey:             "test-key",
		PassthroughBaseURL: srv.URL,
		Timeout:            5 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return srv, client
}

func TestPassthroughPost_UsesRootBaseURL(t *testing.T) {
	_, client := newPassthroughTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/kling/v1/videos/text2video" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %s", got)
		}
		if got := r.Header.Get("X-Trace-Id"); got != "trace-123" {
			t.Fatalf("unexpected trace header: %s", got)
		}

		body := extractBody(t, r)
		if body["model_name"] != "kling-v1" {
			t.Fatalf("unexpected body: %v", body)
		}

		w.Header().Set("X-Task-Route", "passthrough")
		writeJSON(w, http.StatusAccepted, map[string]any{
			"data": map[string]any{
				"task_id": "task_123",
			},
		})
	})

	resp, err := client.Passthrough.Post(
		context.Background(),
		"/kling/v1/videos/text2video",
		sa.JSONMap{"model_name": "kling-v1"},
		sa.WithHeader("X-Trace-Id", "trace-123"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if got := resp.Headers.Get("X-Task-Route"); got != "passthrough" {
		t.Fatalf("unexpected response header: %s", got)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["task_id"] != "task_123" {
		t.Fatalf("unexpected task id: %v", data["task_id"])
	}
}

func TestPassthroughRequestRaw_SendsBodyAsIs(t *testing.T) {
	rawBody := []byte(`{"contents":[{"parts":[{"text":"paint a cat"}]}]}`)

	_, client := newPassthroughTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/google/v1beta/models/gemini-2.5-flash-image:generateContent" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(r.Body)
		if !bytes.Equal(buf.Bytes(), rawBody) {
			t.Fatalf("unexpected raw body: %s", buf.String())
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	resp, err := client.Passthrough.RequestRaw(
		context.Background(),
		http.MethodPost,
		"google/v1beta/models/gemini-2.5-flash-image:generateContent",
		rawBody,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
}

func TestPassthroughReturnsErrorBodyForHTTPErrorStatus(t *testing.T) {
	_, client := newPassthroughTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{"message": "bad request"},
		})
	})

	resp, err := client.Passthrough.Get(context.Background(), "/vidu/v2/tasks/task_123/creations")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if !bytes.Contains(resp.Body, []byte("bad request")) {
		t.Fatalf("unexpected body: %s", string(resp.Body))
	}
}

func TestPassthroughRejectsAbsoluteURL(t *testing.T) {
	_, client := newPassthroughTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not reach server")
	})

	_, err := client.Passthrough.Get(context.Background(), "https://example.com/kling/v1/videos/text2video")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
