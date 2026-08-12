---
name: seaart-sdk-go
description: SeaArt Go SDK assistant — helps users call SeaArt AI platform APIs with sa-go, including multimodal tasks (image/video generation), vendor passthrough, and LLMs (chat, streaming, embeddings, rerank)
type: slash_command
tags:
  - go
  - seaart
  - sdk
  - llm
  - multimodal
---

When this skill is triggered, provide usage guidance for the SeaArt Go SDK (`sa-go`).

**Trigger scenarios:** Use when the user needs to call SeaArt APIs from Go, generate images/videos, call LLM APIs, or troubleshoot SDK usage.

**Workflow:**

1. Choose Modal API (unified multimodal tasks), Passthrough API (vendor-native APIs), or LLM API (text generation) based on the user request
2. Prefer the `input[*].params` structure; for typed construction, use `sa.NewTask(...).Params(...).Build()` to create Modal tasks
3. LLM APIs return `RawResponse`; remind users to deserialize with `sa.Decode[T](raw)`
4. For streaming APIs, recommend using `MessagesStreamTextAssembler` / `ResponsesStreamTextAssembler`
5. For error handling, recommend asserting to `*sa.Error` and branching by `Kind` (ErrAuth/ErrQuota/ErrTimeout/ErrTaskFailed)

**Output format:** Provide runnable Go snippets with brief explanations. Code should use the standard import `sa "github.com/SeaArt-Infra/sea-sdk-go"`.

---

# SeaArt Go SDK Complete Reference

SeaArt Go SDK (`sa-go`) is the official Go client for the SeaArt AI platform. It provides multimodal tasks (image/video generation), vendor passthrough, and LLM text processing capabilities.

**Requirements:** Go 1.22+, no third-party dependencies

## Installation

```bash
go get github.com/SeaArt-Infra/sea-sdk-go
```

## Client Configuration

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey:             "sa-your-api-key",       // Required: SeaArt API Key
    BaseURL:            "https://custom-url.com", // Optional: custom base URL
    ModelBaseURL:       "https://model-url.com",  // Optional: multimodal endpoint
    LLMBaseURL:         "https://llm-url.com",    // Optional: LLM endpoint
    PassthroughBaseURL: "https://model-url.com",  // Optional: vendor passthrough endpoint, defaults to ModelBaseURL
    Project:            "my-project",            // Optional: sent as the X-Project header
    HTTPClient:         &http.Client{},           // Optional: custom HTTP client
    Timeout:            60 * time.Second,         // Optional: default 5 minutes
})
```

**Default endpoint:** `https://gateway.example.com`
**Authentication:** `Authorization: Bearer {apiKey}`

Keep the selected model in the SDK payload's top-level `model` field. The SDK sends it as the `X-Model` header and removes it from the serialized JSON body. Do not pass `X-Model` with `sa.WithHeader(...)` when the payload already contains `model`.

---

## Modal API (Multimodal Tasks)

### Create a Task (Builder Style, Recommended)

```go
body := sa.NewTask("alibaba_wanx26_i2v_flash").
    Moderation(true).
    Params(map[string]any{
        "input": map[string]any{
            "img_url": "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg",
            "prompt":  "A dog and a girl playing happily in an autumn park",
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

task, err := client.Modal.Create(ctx, body)
```

### Create a Task (Raw Style)

```go
task, err := client.Modal.Create(ctx, sa.JSONMap{
    "moderation": true,
    "model": "alibaba_wanx26_i2v_flash",
    "input": []map[string]any{
        {
            "params": map[string]any{
                "input": map[string]any{
                    "img_url": "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg",
                    "prompt":  "A dog and a girl playing happily in an autumn park",
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
})
```

### Precharge Estimate

The precharge route is `/model/v1/generation/precharge`, and its request parameters are the same as task creation.

```go
resp, err := client.Modal.Precharge(ctx, sa.JSONMap{
    "id":         "d88pmute87128c73e9r0d0",
    "model":      "volces_seedream_4_5",
    "input":      []map[string]any{{"params": map[string]any{"prompt": "A dog"}}},
    "moderation": false,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Status)
fmt.Println(resp.Data.BillingModel, resp.Data.Cost, resp.Data.Currency)
```

Typed helper:

```go
body := sa.NewTask("volces_seedream_4_5").
    Moderation(false).
    Field("id", "d88pmute87128c73e9r0d0").
    Params(map[string]any{
        "prompt": "A dog",
    }).
    Build()

resp, err := client.Modal.Precharge(ctx, body)
```

Success response example:

```json
{
  "data": {
    "billing_model": "volces_seedream_4_5",
    "cost": "0.035714285714",
    "currency": "USD",
    "discount": 0.7,
    "hash": "v1:18a733f04d227d572950ed8f1f98a9ba4cd37c168c5c98c05a5e574984f58eaf",
    "model": "volces_seedream_4_5",
    "original_model": "volces_seedream_4_5",
    "sample_count": 4,
    "updated_at": 1780633394064
  },
  "status": "success"
}
```

When no precharge data is matched, the response may be:

```json
{
  "data": {
    "cost": null,
    "hash": "v1:02833b68895eeb61bf214d35fd669502ef788e4c8d58505893414ae9632ca8ab",
    "model": "volces_seedream_4_5",
    "original_model": "volces_seedream_4_5",
    "reason": "COST_CACHE_MISS"
  },
  "status": "failed"
}
```

### Wait for Task Completion

```go
task, err = task.Wait(ctx,
    sa.WithPollInterval(3*time.Second),
    sa.WithPollTimeout(5*time.Minute),
    sa.WithPollCallback(func(status string, progress float64) {
        fmt.Printf("Status: %s, Progress: %.1f%%\n", status, progress*100)
    }),
)
```

**Polling options:** default interval 3s, default timeout 5 minutes.

### Get Task Results

```go
for _, output := range task.Output {
    for _, content := range output.Content {
        fmt.Printf("Type: %s, URL: %s\n", content.Type, content.URL)
    }
}
```

**Task Status:** `"in_progress"` / `"completed"` / `"failed"`

### Image/Video Safety Scan

Use `client.Modal.ScanImage` to call `ModelBaseURL + /v1/image/scan`. Pass either `URI` or `ImgBase64`; videos must use `URI`. Pass `CallbackURL` to enable async processing.

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI: "https://example.com/image.jpg",
    RiskTypes: []sa.ImageScanRiskType{
        sa.ImageScanRiskTypePolity,
        sa.ImageScanRiskTypeErotic,
        sa.ImageScanRiskTypeViolent,
        sa.ImageScanRiskTypeChild,
    },
    IsVideo: false,
    Canary:  "B",
    Scene:   "avatar",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.OK, resp.NSFWLevel, resp.RiskTypes)
```

For video scans, set `IsVideo: 1` and optionally pass `Duration`; `FrameResults` in the response contains frame-level scan results.

Risk type descriptions:

| Constant | API Value | Description |
|------|--------|------|
| `sa.ImageScanRiskTypePolity` | `POLITY` | Political, public-safety, or related sensitive content |
| `sa.ImageScanRiskTypeErotic` | `EROTIC` | Erotic, nudity, sexually suggestive, or other adult content |
| `sa.ImageScanRiskTypeViolent` | `VIOLENT` | Violence, gore, weapons, harm, or related content |
| `sa.ImageScanRiskTypeChild` | `CHILD` | Child-safety risks, especially unsafe or sexualized child-related content |

### Sensitive-Word Scan

Use `client.Modal.ScanText` to call `ModelBaseURL + /v1/text/scan`.

```go
resp, err := client.Modal.ScanText(ctx, sa.TextScanRequest{
    Text:      "prompt to check",
    Scene:     1,
    AreaTypes: []sa.TextScanAreaType{sa.TextScanAreaTypeForeign},
    Way:       sa.TextScanWayDictionary,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Usage)
fmt.Println(resp.Status.Code, resp.Status.Msg)
fmt.Println(resp.Data.IsSensitive)
fmt.Println(resp.Data.SensitiveWords)
fmt.Println(resp.Data.Combination)
```

`AreaTypes` supports `TextScanAreaTypeAll`, `TextScanAreaTypeDomestic`, and `TextScanAreaTypeForeign`. `Way` supports `TextScanWayDictionary`, `TextScanWayModel`, `TextScanWayMixed`, and `TextScanWayCharacter`. Sensitive-word indexes `StartIndex` / `EndIndex` are based on rune arrays. `IsSensitive` indicates whether the whole text matched sensitive content. `Combination` keeps combination-rule match details. Unmodeled fields are preserved in `Extra`.


### Text Content Safety Scan

Use `client.Modal.ScanTextContent` to call `ModelBaseURL + /v1/text/content/scan`. This endpoint reviews short text for content safety and does not affect the legacy sensitive-word API `client.Modal.ScanText`.

```go
resp, err := client.Modal.ScanTextContent(ctx, sa.TextContentScanRequest{
    Text:   "hello world",
    Canary: "A",
    Scene:  "user_name",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.OK, resp.Level, resp.Label)
fmt.Println(resp.ReqID, resp.Reason, resp.Usage)
```

`TextContentScanRequest` contains required `Text` plus optional `Canary` and `Scene`. `TextContentScanResponse` contains `OK`, `ReqID`, `Level`, `Label`, `Reason`, `Usage`, and unmodeled fields in `Extra`. Log `ReqID` for downstream request tracing.

### Visual Structured Text Fusion Scan

Use `client.Modal.ScanVisualStructuredTextFusion` for `POST /v1/visual/structured/text/fusion/scan`. It requires `TextDict` plus `URI` or `ImgBase64` and returns image/text risk details, downstream `ReqID` for request tracing, and gateway usage.

### Face Scan

Use `client.Modal.ScanFace` to call `ModelBaseURL + /v1/face/scan`. The gateway forwards the request to upstream `/cloud/face/scan`.

```go
resp, err := client.Modal.ScanFace(ctx, sa.FaceScanRequest{
    URI:     "https://example.com/image.jpg",
    IsVideo: 0,
    Scene:   "avatar",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.OK, resp.Usage)
fmt.Println(resp.Extra["face_count"])
```

You can also pass `ImgBase64`. For video scans, set `IsVideo: 1` and optionally pass `Duration`; unmodeled upstream response fields are preserved in `Extra`.

---

## Passthrough API (Vendor Passthrough)

Paths must include a vendor prefix, such as `/kling/...`, `/vidu/...`, or `/google/...`.

```go
resp, err := client.Passthrough.Post(ctx, "/kling/v1/videos/text2video", sa.JSONMap{
    "model_name": "kling-v1",
    "prompt":     "cinematic shot",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.StatusCode, string(resp.Body))
```

Use `RequestRaw` when fully passing through raw JSON bytes:

```go
resp, err := client.Passthrough.RequestRaw(
    ctx,
    http.MethodPost,
    "/google/v1beta/models/gemini-2.5-flash-image:generateContent",
    []byte(`{"contents":[{"parts":[{"text":"paint a cat"}]}]}`),
)
```

---

## LLM API

### Chat Completions (OpenAI Compatible)

```go
// Non-streaming
raw, err := client.LLM.ChatCompletions(ctx, sa.JSONMap{
    "model":      "gpt-4o-mini",
    "messages":   []map[string]any{{"role": "user", "content": "hello"}},
    "max_tokens": 64,
})
resp, _ := sa.Decode[sa.ChatCompletionResponse](raw)
fmt.Println(resp.Choices[0].Message.Content)

// Streaming
ch, err := client.LLM.ChatCompletionsStream(ctx, sa.JSONMap{
    "model":    "gpt-4o-mini",
    "messages": []map[string]any{{"role": "user", "content": "hello"}},
})
for event := range ch {
    if event.Err != nil || event.Done { break }
    chunk, _ := sa.Decode[sa.ChatCompletionResponse](event.Data)
    fmt.Print(chunk.Choices[0].Delta.Content)
}
```

### Messages API (Anthropic Format)

```go
// Streaming + text assembler
ch, err := client.LLM.MessagesStream(ctx, sa.JSONMap{
    "model":      "claude-3-5-sonnet",
    "messages":   []sa.JSONMap{{"role": "user", "content": "hello"}},
    "max_tokens": 256,
})
var asm sa.MessagesStreamTextAssembler
for event := range ch {
    if event.Done { break }
    chunk, _ := sa.Decode[sa.MessagesStreamChunk](event.Data)
    asm.Add(chunk)
}
fmt.Println(asm.Text())
```

### Responses API

```go
ch, err := client.LLM.ResponsesStream(ctx, payload)
var asm sa.ResponsesStreamTextAssembler
for event := range ch {
    if event.Done { break }
    chunk, _ := sa.Decode[sa.ResponsesResponseStreamChunk](event.Data)
    asm.Add(chunk)
}
fmt.Println(asm.Text())
```

### Embeddings

```go
raw, err := client.LLM.Embeddings(ctx, sa.JSONMap{
    "model": "text-embedding-3-small",
    "input": "text to embed",
})
resp, _ := sa.Decode[sa.EmbeddingsResponse](raw)
```

### Reranking

```go
raw, err := client.LLM.Rerank(ctx, sa.JSONMap{
    "model":     "rerank-model",
    "query":     "search query",
    "documents": []string{"document 1", "document 2"},
})
resp, _ := sa.Decode[sa.RerankResponse](raw)
for _, r := range resp.Results {
    fmt.Printf("Index: %d, Score: %.4f\n", r.Index, r.RelevanceScore)
}
```

### List Available Models

```go
raw, err := client.LLM.ListModels(ctx)
resp, _ := sa.Decode[sa.LLMModelListResponse](raw)
for _, model := range resp.Data {
    fmt.Println(model.ID)
}
```

---

## Request Options

```go
client.LLM.ChatCompletions(ctx, payload,
    sa.WithHeader("X-Trace-Id", "abc-123"),
    sa.WithHeader("X-Tenant-Id", "tenant-a"),
)
```

---

## Error Handling

```go
if err != nil {
    if sdkErr, ok := err.(*sa.Error); ok {
        switch sdkErr.Kind {
        case sa.ErrAuth:       // 401/403 — invalid API key
        case sa.ErrQuota:      // 429 — rate limit exceeded
        case sa.ErrTimeout:    // 408/504 — timeout
        case sa.ErrNetwork:    // Network connection error
        case sa.ErrTaskFailed: // Task execution failed
        default:               // sa.ErrGeneral
        }
    }
}
```

---

## Complete Example: Video Generation

```go
package main

import (
    "context"
    "fmt"
    "log"

    sa "github.com/SeaArt-Infra/sea-sdk-go"
)

func main() {
    client, err := sa.New(&sa.ClientConfig{APIKey: "sa-your-api-key"})
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    task, err := client.Modal.Create(ctx,
        sa.NewTask("alibaba_wanx26_i2v_flash").
            Moderation(true).
            Params(map[string]any{
                "input": map[string]any{
                    "img_url": "https://dashscope.oss-cn-beijing.aliyuncs.com/images/dog_and_girl.jpeg",
                    "prompt":  "A dog and a girl playing happily in an autumn park",
                },
                "parameters": map[string]any{
                    "resolution":    "720P",
                    "duration":      5,
                    "prompt_extend": true,
                    "watermark":     false,
                },
            }).
            Build(),
    )
    if err != nil {
        log.Fatal(err)
    }

    task, err = task.Wait(ctx,
        sa.WithPollCallback(func(status string, progress float64) {
            fmt.Printf("\rProgress: %.0f%%", progress*100)
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    for _, output := range task.Output {
        for _, content := range output.Content {
            fmt.Printf("\nVideo URL: %s\n", content.URL)
        }
    }
}
```
