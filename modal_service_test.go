package sa_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	sa "github.com/SeaVerseAI/sea-sdk-go"
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

func TestModalListModels_SearchesSkillModels(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/models/skill/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization: %s", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("unexpected accept: %s", got)
		}

		query := r.URL.Query()
		if query.Get("q") != "animate" {
			t.Fatalf("unexpected q: %s", query.Get("q"))
		}
		if query.Get("input") != "image" {
			t.Fatalf("unexpected input: %s", query.Get("input"))
		}
		if query.Get("output") != "video" {
			t.Fatalf("unexpected output: %s", query.Get("output"))
		}
		if query.Get("type") != "i2v" {
			t.Fatalf("unexpected type: %s", query.Get("type"))
		}
		if query.Get("provider") != "alibaba" {
			t.Fatalf("unexpected provider: %s", query.Get("provider"))
		}
		if query.Get("limit") != "2" {
			t.Fatalf("unexpected limit: %s", query.Get("limit"))
		}

		writeJSON(w, 200, map[string]any{
			"hits": []map[string]any{
				{
					"id":            "alibaba_animate_anyone_detect",
					"name":          "alibaba_animate_anyone_detect",
					"provider":      "alibaba",
					"input":         "image",
					"output":        "video",
					"media_type":    "video",
					"tags":          []string{"i2v"},
					"tags_abbr":     "i2v",
					"skill_content": "# alibaba_animate_anyone_detect",
				},
			},
			"query":              "animate",
			"limit":              2,
			"estimatedTotalHits": 1,
		})
	})

	resp, err := client.Modal.ListModels(context.Background(), sa.ModalModelSearchParams{
		Query:    "animate",
		Input:    "image",
		Output:   "video",
		Type:     "i2v",
		Provider: "alibaba",
		Limit:    2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Query != "animate" {
		t.Fatalf("unexpected query: %s", resp.Query)
	}
	if resp.Limit != 2 {
		t.Fatalf("unexpected limit: %d", resp.Limit)
	}
	if resp.EstimatedTotalHits != 1 {
		t.Fatalf("unexpected total hits: %d", resp.EstimatedTotalHits)
	}
	if len(resp.Hits) != 1 {
		t.Fatalf("unexpected hit count: %d", len(resp.Hits))
	}
	if resp.Hits[0]["name"] != "alibaba_animate_anyone_detect" {
		t.Fatalf("unexpected hit name: %v", resp.Hits[0]["name"])
	}
}

func TestModalSearchModelsAlias(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/models/skill/search" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "" {
			t.Fatalf("unexpected q: %s", got)
		}
		if got := r.URL.Query().Get("limit"); got != "2" {
			t.Fatalf("unexpected limit: %s", got)
		}

		writeJSON(w, 200, map[string]any{
			"hits":  []map[string]any{},
			"query": "",
			"limit": 2,
		})
	})

	resp, err := client.Modal.SearchModels(context.Background(), sa.ModalModelSearchParams{Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Limit != 2 {
		t.Fatalf("unexpected limit: %d", resp.Limit)
	}
}

func TestModalGetModelSkill_ReturnsMarkdown(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/models/skill/alibaba_animate_anyone_detect" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization: %s", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("unexpected accept: %s", got)
		}

		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# alibaba_animate_anyone_detect\n\nparameters"))
	})

	content, err := client.Modal.GetModelSkill(context.Background(), "alibaba_animate_anyone_detect")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "# alibaba_animate_anyone_detect\n\nparameters" {
		t.Fatalf("unexpected content: %q", content)
	}
}

func TestModalGetModelSkill_RequiresModel(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("request should not be sent: %s %s", r.Method, r.URL.Path)
	})

	_, err := client.Modal.GetModelSkill(context.Background(), " ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestModalScanImage_PostsImageScanRequest(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/image/scan" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization: %s", got)
		}
		if got := r.Header.Get("X-Trace-Id"); got != "trace-scan" {
			t.Fatalf("unexpected trace header: %s", got)
		}

		body := extractBody(t, r)
		if body["uri"] != "https://example.com/image.jpg" {
			t.Fatalf("unexpected uri: %v", body["uri"])
		}
		if body["detected_age"] != float64(1) {
			t.Fatalf("unexpected detected_age: %v", body["detected_age"])
		}
		if body["is_video"] != float64(0) {
			t.Fatalf("unexpected is_video: %v", body["is_video"])
		}
		risks := body["risk_types"].([]any)
		if len(risks) != 2 || risks[0] != "EROTIC" || risks[1] != "VIOLENT" {
			t.Fatalf("unexpected risk_types: %v", risks)
		}

		writeJSON(w, 200, map[string]any{
			"ok":         true,
			"nsfw_level": 2,
			"label_items": []map[string]any{
				{"name": "safe", "score": 2, "risk_type": "EROTIC"},
			},
			"risk_types": []string{"EROTIC"},
			"usage": map[string]any{
				"cost": "0.001",
			},
		})
	})

	resp, err := client.Modal.ScanImage(context.Background(), sa.ImageScanRequest{
		URI: "https://example.com/image.jpg",
		RiskTypes: []sa.ImageScanRiskType{
			sa.ImageScanRiskTypeErotic,
			sa.ImageScanRiskTypeViolent,
		},
		DetectedAge: 1,
	}, sa.WithHeader("X-Trace-Id", "trace-scan"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.OK {
		t.Fatal("expected ok response")
	}
	if resp.NSFWLevel != 2 {
		t.Fatalf("unexpected nsfw level: %d", resp.NSFWLevel)
	}
	if len(resp.LabelItems) != 1 || resp.LabelItems[0].RiskType != sa.ImageScanRiskTypeErotic {
		t.Fatalf("unexpected labels: %+v", resp.LabelItems)
	}
	if resp.Usage == nil || resp.Usage.Cost.String() != "0.001" {
		t.Fatalf("unexpected usage: %+v", resp.Usage)
	}
}

func TestModalScanImage_RequiresURI(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("request should not be sent: %s %s", r.Method, r.URL.Path)
	})

	_, err := client.Modal.ScanImage(context.Background(), sa.ImageScanRequest{URI: " "})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestModalScanFace_PostsFaceScanRequest(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/face/scan" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization: %s", got)
		}
		if got := r.Header.Get("X-Trace-Id"); got != "trace-face" {
			t.Fatalf("unexpected trace header: %s", got)
		}

		body := extractBody(t, r)
		if body["uri"] != "https://example.com/face.jpg" {
			t.Fatalf("unexpected uri: %v", body["uri"])
		}
		if body["is_video"] != float64(0) {
			t.Fatalf("unexpected is_video: %v", body["is_video"])
		}
		if body["scene"] != "avatar" {
			t.Fatalf("unexpected scene: %v", body["scene"])
		}
		if body["canary"] != "gray" {
			t.Fatalf("unexpected canary: %v", body["canary"])
		}

		writeJSON(w, 200, map[string]any{
			"ok":         true,
			"face_count": 1,
			"faces": []map[string]any{
				{"score": 0.99},
			},
			"usage": map[string]any{
				"cost": "0.002",
			},
		})
	})

	resp, err := client.Modal.ScanFace(context.Background(), sa.FaceScanRequest{
		URI:     "https://example.com/face.jpg",
		IsVideo: 0,
		Canary:  "gray",
		Scene:   "avatar",
	}, sa.WithHeader("X-Trace-Id", "trace-face"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.OK {
		t.Fatal("expected ok response")
	}
	if resp.Usage == nil || resp.Usage.Cost.String() != "0.002" {
		t.Fatalf("unexpected usage: %+v", resp.Usage)
	}
	if resp.Extra["face_count"] != float64(1) {
		t.Fatalf("unexpected extra fields: %+v", resp.Extra)
	}
}

func TestModalScanFace_AcceptsBase64AndVideoDuration(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/face/scan" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		body := extractBody(t, r)
		if body["img_base64"] != "abc123" {
			t.Fatalf("unexpected img_base64: %v", body["img_base64"])
		}
		if body["is_video"] != float64(1) {
			t.Fatalf("unexpected is_video: %v", body["is_video"])
		}
		if body["duration"] != 12.5 {
			t.Fatalf("unexpected duration: %v", body["duration"])
		}

		writeJSON(w, 200, map[string]any{
			"ok":             true,
			"video_duration": 12.5,
		})
	})

	resp, err := client.Modal.ScanFace(context.Background(), sa.FaceScanRequest{
		ImgBase64: "abc123",
		IsVideo:   1,
		Duration:  12.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Extra["video_duration"] != 12.5 {
		t.Fatalf("unexpected extra fields: %+v", resp.Extra)
	}
}

func TestModalScanFace_RequiresURIOrBase64(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("request should not be sent: %s %s", r.Method, r.URL.Path)
	})

	_, err := client.Modal.ScanFace(context.Background(), sa.FaceScanRequest{
		URI:       " ",
		ImgBase64: " ",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
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
