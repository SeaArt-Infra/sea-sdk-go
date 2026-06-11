package sa_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	sa "github.com/SeaArt-Infra/sea-sdk-go"
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
		if body["model"] != "alibaba_wanx26_i2v_flash" {
			t.Fatalf("unexpected model: %v", body["model"])
		}
		if body["moderation"] != true {
			t.Fatalf("unexpected moderation: %v", body["moderation"])
		}

		input := body["input"].([]any)
		params := input[0].(map[string]any)["params"].(map[string]any)
		modelInput := params["input"].(map[string]any)
		if modelInput["img_url"] != "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg" {
			t.Fatalf("unexpected img_url: %v", modelInput["img_url"])
		}
		parameters := params["parameters"].(map[string]any)
		if parameters["duration"] != float64(5) {
			t.Fatalf("unexpected duration: %v", parameters["duration"])
		}

		writeJSON(w, 200, map[string]any{
			"id":     "task_create",
			"status": "in_progress",
			"model":  "alibaba_wanx26_i2v_flash",
		})
	})

	task, err := client.Modal.Create(context.Background(), sa.JSONMap{
		"moderation": true,
		"model":      "alibaba_wanx26_i2v_flash",
		"input": []map[string]any{
			{
				"params": map[string]any{
					"input": map[string]any{
						"img_url": "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg",
						"prompt":  "小狗和女孩在秋天的公园里快乐地玩耍",
					},
					"parameters": map[string]any{
						"resolution":    "720P",
						"duration":      5,
						"prompt_extend": true,
						"watermark":     false,
					},
				},
			},
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
	if task.Model != "alibaba_wanx26_i2v_flash" {
		t.Fatalf("unexpected model: %s", task.Model)
	}
}

func TestModalPrecharge_ReturnsBillingPreview(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/generation/precharge" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		body := extractBody(t, r)
		if body["id"] != "d88pmute87128c73e9r0d0" {
			t.Fatalf("unexpected id: %v", body["id"])
		}
		if body["model"] != "volces_seedream_4_5" {
			t.Fatalf("unexpected model: %v", body["model"])
		}
		if body["moderation"] != false {
			t.Fatalf("unexpected moderation: %v", body["moderation"])
		}
		input := body["input"].([]any)
		params := input[0].(map[string]any)["params"].(map[string]any)
		if params["prompt"] != "A dog" {
			t.Fatalf("unexpected prompt: %v", params["prompt"])
		}

		writeJSON(w, 200, map[string]any{
			"data": map[string]any{
				"billing_model":  "volces_seedream_4_5",
				"cost":           "0.035714285714",
				"currency":       "USD",
				"discount":       0.7,
				"hash":           "v1:18a733f04d227d572950ed8f1f98a9ba4cd37c168c5c98c05a5e574984f58eaf",
				"model":          "volces_seedream_4_5",
				"original_model": "volces_seedream_4_5",
				"sample_count":   4,
				"updated_at":     1780633394064,
			},
			"status": "success",
		})
	})

	resp, err := client.Modal.Precharge(context.Background(), sa.JSONMap{
		"id":         "d88pmute87128c73e9r0d0",
		"model":      "volces_seedream_4_5",
		"input":      []map[string]any{{"params": map[string]any{"prompt": "A dog"}}},
		"moderation": false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	if resp.Data == nil {
		t.Fatal("expected precharge data")
	}
	if resp.Data.BillingModel != "volces_seedream_4_5" {
		t.Fatalf("unexpected billing model: %s", resp.Data.BillingModel)
	}
	if resp.Data.Cost == nil || resp.Data.Cost.String() != "0.035714285714" {
		t.Fatalf("unexpected cost: %v", resp.Data.Cost)
	}
	if resp.Data.Currency != "USD" {
		t.Fatalf("unexpected currency: %s", resp.Data.Currency)
	}
	if resp.Data.SampleCount != 4 {
		t.Fatalf("unexpected sample count: %d", resp.Data.SampleCount)
	}
}

func TestModalPrecharge_SupportsCacheMissResponse(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/generation/precharge" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeJSON(w, 200, map[string]any{
			"data": map[string]any{
				"cost":           nil,
				"hash":           "v1:02833b68895eeb61bf214d35fd669502ef788e4c8d58505893414ae9632ca8ab",
				"model":          "volces_seedream_4_5",
				"original_model": "volces_seedream_4_5",
				"reason":         "COST_CACHE_MISS",
			},
			"status": "failed",
		})
	})

	resp, err := client.Modal.Precharge(context.Background(), sa.JSONMap{
		"id":         "d88pmute87128c73e9r0d0",
		"model":      "volces_seedream_4_5",
		"input":      []map[string]any{{"params": map[string]any{"prompt": "A dog"}}},
		"moderation": false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "failed" {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	if resp.Data == nil {
		t.Fatal("expected precharge data")
	}
	if resp.Data.Cost != nil {
		t.Fatalf("expected nil cost, got %v", resp.Data.Cost)
	}
	if resp.Data.Reason != "COST_CACHE_MISS" {
		t.Fatalf("unexpected reason: %s", resp.Data.Reason)
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

func TestModalScanText_PostsTextScanRequest(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1/text/scan" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization: %s", got)
		}
		if got := r.Header.Get("X-Trace-Id"); got != "trace-text" {
			t.Fatalf("unexpected trace header: %s", got)
		}

		body := extractBody(t, r)
		if body["text"] != "a prompt to check" {
			t.Fatalf("unexpected text: %v", body["text"])
		}
		if body["scene"] != float64(1) {
			t.Fatalf("unexpected scene: %v", body["scene"])
		}
		areaTypes := body["area_types"].([]any)
		if len(areaTypes) != 1 || areaTypes[0] != float64(2) {
			t.Fatalf("unexpected area_types: %v", areaTypes)
		}
		if body["way"] != float64(0) {
			t.Fatalf("unexpected way: %v", body["way"])
		}

		writeJSON(w, 200, map[string]any{
			"data": map[string]any{
				"sensitive_words": []map[string]any{
					{
						"word":           "blocked",
						"start_index":    2,
						"end_index":      8,
						"risk_type_code": "political",
					},
				},
			},
			"status": map[string]any{
				"code":       10000,
				"msg":        "success",
				"request_id": "risk-req-1",
			},
			"usage": map[string]any{
				"cost": "0.003",
			},
		})
	})

	resp, err := client.Modal.ScanText(context.Background(), sa.TextScanRequest{
		Text:      "a prompt to check",
		Scene:     1,
		AreaTypes: []sa.TextScanAreaType{sa.TextScanAreaTypeForeign},
		Way:       sa.TextScanWayDictionary,
	}, sa.WithHeader("X-Trace-Id", "trace-text"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status == nil || resp.Status.Code != 10000 || resp.Status.RequestID != "risk-req-1" {
		t.Fatalf("unexpected status: %+v", resp.Status)
	}
	if resp.Data == nil || len(resp.Data.SensitiveWords) != 1 {
		t.Fatalf("unexpected data: %+v", resp.Data)
	}
	word := resp.Data.SensitiveWords[0]
	if word.Word != "blocked" || word.StartIndex != 2 || word.EndIndex != 8 || word.RiskTypeCode != "political" {
		t.Fatalf("unexpected sensitive word: %+v", word)
	}
	if resp.Usage == nil || resp.Usage.Cost.String() != "0.003" {
		t.Fatalf("unexpected usage: %+v", resp.Usage)
	}
	if len(resp.Extra) != 0 {
		t.Fatalf("unexpected extra fields: %+v", resp.Extra)
	}
}

func TestModalScanText_PreservesEmptyResultFields(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/text/scan" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		writeJSON(w, 200, map[string]any{
			"data": map[string]any{
				"sensitive_words": []map[string]any{},
				"combination":     nil,
				"is_sensitive":    false,
			},
			"status": map[string]any{
				"code":       10000,
				"msg":        "success",
				"request_id": "risk-empty",
			},
			"usage": map[string]any{
				"cost": "1",
			},
		})
	})

	resp, err := client.Modal.ScanText(context.Background(), sa.TextScanRequest{
		Text:      "clean prompt",
		Scene:     1,
		AreaTypes: []sa.TextScanAreaType{sa.TextScanAreaTypeDomestic, sa.TextScanAreaTypeForeign},
		Way:       sa.TextScanWayDictionary,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data == nil {
		t.Fatal("expected data, got nil")
	}
	if len(resp.Data.SensitiveWords) != 0 {
		t.Fatalf("unexpected sensitive words: %+v", resp.Data.SensitiveWords)
	}
	if resp.Data.IsSensitive {
		t.Fatal("expected is_sensitive=false")
	}

	encoded, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("unmarshal encoded response: %v", err)
	}
	data := payload["data"].(map[string]any)
	if words, ok := data["sensitive_words"].([]any); !ok || len(words) != 0 {
		t.Fatalf("expected empty sensitive_words array, got %#v", data["sensitive_words"])
	}
	if _, ok := data["combination"]; !ok {
		t.Fatalf("expected combination field, got %#v", data)
	}
	if sensitive, ok := data["is_sensitive"].(bool); !ok || sensitive {
		t.Fatalf("expected is_sensitive=false, got %#v", data["is_sensitive"])
	}
}

func TestModalScanText_RequiresText(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("request should not be sent: %s %s", r.Method, r.URL.Path)
	})

	_, err := client.Modal.ScanText(context.Background(), sa.TextScanRequest{Text: " "})
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
	body := sa.NewTask("alibaba_wanx26_i2v_flash").
		Moderation(true).
		Params(map[string]any{
			"input": map[string]any{
				"img_url": "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg",
				"prompt":  "小狗和女孩在秋天的公园里快乐地玩耍",
			},
			"parameters": map[string]any{
				"resolution":    "720P",
				"duration":      5,
				"prompt_extend": true,
				"watermark":     false,
			},
		}).
		Metadata("trace_id", "trace-123").
		Build()

	if body["model"] != "alibaba_wanx26_i2v_flash" {
		t.Fatalf("unexpected model: %v", body["model"])
	}
	if body["moderation"] != true {
		t.Fatalf("unexpected moderation: %v", body["moderation"])
	}
	input := body["input"].([]map[string]any)
	params := input[0]["params"].(map[string]any)
	modelInput := params["input"].(map[string]any)
	if modelInput["img_url"] != "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg" {
		t.Fatalf("unexpected img_url: %v", modelInput["img_url"])
	}
	parameters := params["parameters"].(map[string]any)
	if parameters["duration"] != 5 {
		t.Fatalf("unexpected duration: %v", parameters["duration"])
	}
	if body["metadata"].(map[string]any)["trace_id"] != "trace-123" {
		t.Fatalf("unexpected metadata: %v", body["metadata"])
	}
}

func TestTaskBuilderSupportsFlatParamsAndTopLevelFields(t *testing.T) {
	body := sa.NewTask("grok_imagine_image").
		Field("dash_scope", true).
		Moderation(true).
		Params(map[string]any{
			"aspect_ratio": "1:2",
			"prompt":       "Lego art version of Superman and Batman，Night scene",
			"n":            1,
			"resolution":   "1k",
		}).
		Build()

	if body["dash_scope"] != true {
		t.Fatalf("unexpected dash_scope: %v", body["dash_scope"])
	}
	if body["moderation"] != true {
		t.Fatalf("unexpected moderation: %v", body["moderation"])
	}
	input := body["input"].([]map[string]any)
	params := input[0]["params"].(map[string]any)
	if params["aspect_ratio"] != "1:2" {
		t.Fatalf("unexpected aspect ratio: %v", params["aspect_ratio"])
	}
	if params["prompt"] != "Lego art version of Superman and Batman，Night scene" {
		t.Fatalf("unexpected prompt: %v", params["prompt"])
	}
	if params["n"] != 1 {
		t.Fatalf("unexpected n: %v", params["n"])
	}
	if params["resolution"] != "1k" {
		t.Fatalf("unexpected resolution: %v", params["resolution"])
	}
}
