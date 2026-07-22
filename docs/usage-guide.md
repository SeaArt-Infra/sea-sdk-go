# SeaArt Go SDK Usage Guide

SeaArt Go SDK (`sa-go`) is the official Go client for the SeaArt AI platform. It provides multimodal tasks (image/video generation), vendor passthrough, and LLM text processing capabilities.

**Requirements:** Go 1.22+, no third-party dependencies

---

## Installation

```bash
go get github.com/SeaArt-Infra/sea-sdk-go
```

---

## Quick Start

```go
import sa "github.com/SeaArt-Infra/sea-sdk-go"

client, err := sa.New(&sa.ClientConfig{
    APIKey: "sa-your-api-key",
})
if err != nil {
    log.Fatal(err)
}
```

---

## Client Configuration

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey:     "sa-your-api-key",          // Required: SeaArt API Key
    BaseURL:    "https://gateway.example.com", // Optional: custom gateway address
    Project:    "my-project",               // Optional: sent as the X-Project header
    HTTPClient: &http.Client{},              // Optional: custom HTTP client
    Timeout:    60 * time.Second,            // Optional: default 5 minutes
})
```

**Default gateway address:** `https://gateway.example.com`
**Authentication:** `Authorization: Bearer {apiKey}`

Usually you only need to configure `BaseURL`; the SDK uses the same gateway address to call multimodal, LLM, and vendor passthrough capabilities.

---

## Multimodal API

Used for multimodal AI tasks such as image generation and video generation.

**Create a task**

```go
ctx := context.Background()

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

**Typed helper**

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

**Success response example**

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

**When no precharge data is matched, the response may be**

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
// Method 1: wait directly on the task object
task, err = task.Wait(ctx,
    sa.WithPollInterval(3*time.Second),
    sa.WithPollTimeout(5*time.Minute),
    sa.WithPollCallback(func(status string, progress float64) {
        fmt.Printf("Status: %s, Progress: %.1f%%\n", status, progress*100)
    }),
)

// Method 2: wait through the client
task, err = client.Modal.Wait(ctx, task.ID)
```

**Polling options:**

| Option | Description | Default |
|------|------|--------|
| `sa.WithPollInterval(d)` | Polling interval | 3s |
| `sa.WithPollTimeout(d)` | Maximum wait time | 5 minutes |
| `sa.WithPollCallback(fn)` | Progress callback | - |

### Get Task Results

```go
if task.Status == "completed" {
    for _, output := range task.Output {
        for _, content := range output.Content {
            fmt.Printf("Type: %s, URL: %s\n", content.Type, content.URL)
        }
    }
}
```

### Passthrough API (Vendor Passthrough)

Used to call vendor-native API endpoints. Paths must include a vendor prefix, such as `/kling/...`, `/vidu/...`, or `/google/...`.

#### JSON Request

```go
resp, err := client.Passthrough.Post(ctx, "/kling/v1/videos/text2video", sa.JSONMap{
    "model_name": "kling-v1",
    "prompt":     "cinematic shot",
}, sa.WithHeader("X-Trace-Id", "trace-123"))
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.StatusCode)
fmt.Println(string(resp.Body))
```

#### Raw Body Passthrough

```go
resp, err := client.Passthrough.RequestRaw(
    ctx,
    http.MethodPost,
    "/google/v1beta/models/gemini-2.5-flash-image:generateContent",
    []byte(`{"contents":[{"parts":[{"text":"paint a cat"}]}]}`),
)
if err != nil {
    log.Fatal(err)
}
```

`PassthroughResponse` preserves the response status code, response headers, and raw body:

```go
type PassthroughResponse struct {
    StatusCode int
    Headers    http.Header
    Body       sa.RawResponse
}
```

---

## Image/Video Safety Scan

The safety scan API maps to `POST /v1/image/scan` and is used for risk detection on images or videos. Provide either `URI` or `ImgBase64`; videos must use `URI`.

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI: "https://example.com/image.jpg",
    RiskTypes: []sa.ImageScanRiskType{
        sa.ImageScanRiskTypePolity,
        sa.ImageScanRiskTypeErotic,
        sa.ImageScanRiskTypeViolent,
        sa.ImageScanRiskTypeChild,
    },
    DetectedAge: true,
    IsVideo:     false,
    Canary:      "B",
    Scene:       "avatar",
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.OK, resp.NSFWLevel, resp.RiskTypes)
for _, label := range resp.LabelItems {
    fmt.Println(label.Name, label.Score, label.RiskType)
}
```

For video detection, set `IsVideo: true` and optionally pass `Duration`. Video scans must use `URI` and do not support `ImgBase64`:

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI:       "https://example.com/video.mp4",
    RiskTypes: []sa.ImageScanRiskType{sa.ImageScanRiskTypeErotic, sa.ImageScanRiskTypeViolent},
    IsVideo:   true,
    Duration:  12.5,
})
```

Base64 image content is also supported for image scans:

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{ImgBase64: "base64-image-content"})
```

Pass `CallbackURL` to enable async processing; `CallbackContext` is returned unchanged in the callback.

Risk type descriptions:

| Constant | API Value | Description |
|------|--------|------|
| `sa.ImageScanRiskTypePolity` | `POLITY` | Political, public-safety, or related sensitive content |
| `sa.ImageScanRiskTypeErotic` | `EROTIC` | Erotic, nudity, sexually suggestive, or other adult content |
| `sa.ImageScanRiskTypeViolent` | `VIOLENT` | Violence, gore, weapons, harm, or related content |
| `sa.ImageScanRiskTypeChild` | `CHILD` | Child-safety risks, especially unsafe or sexualized child-related content |

## Sensitive-Word Scan

The sensitive-word scan API maps to `POST /v1/text/scan`.

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


## Text Content Safety Scan

The text content safety scan endpoint is `POST /v1/text/content/scan`. It reviews short text and returns the risk level, category label, and judgment reason. This endpoint does not affect the legacy sensitive-word scan endpoint `POST /v1/text/scan`.

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
fmt.Println(resp.ReqID, resp.Reason)
fmt.Println(resp.Usage)
```

`TextContentScanRequest` contains required `Text` plus optional `Canary` and `Scene`. `TextContentScanResponse` contains `OK`, `ReqID`, `Level`, `Label`, `Reason`, `Usage`, and unmodeled fields in `Extra`. Log `ReqID` for downstream request tracing.

## Visual Structured Text Fusion Scan

The visual structured text fusion scan endpoint is `POST /v1/visual/structured/text/fusion/scan`. It evaluates a digital-human cover image together with structured copy. `TextDict` supports nested objects, and image URLs inside it are also scanned.

```go
resp, err := client.Modal.ScanVisualStructuredTextFusion(ctx, sa.VisualStructuredTextFusionScanRequest{
    URI: "https://example.com/cover.jpg",
    TextDict: map[string]any{
        "name":        "Xiaomei",
        "personality": "Gentle and considerate",
        "description": "Enjoys traveling",
        "greeting":    "Hello",
    },
    BusinessType: "v1",
    Canary:       "A",
    Mode:         "mixed",
    OCR:          1,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.OK, resp.NSFWLevel, resp.IssueSource, resp.RiskKeys)
fmt.Println(resp.ReqID, resp.Reason, resp.ImgReason, resp.TextReason)
fmt.Println(resp.Usage)
```

`TextDict` is required, and at least one of `URI` and `ImgBase64` must be provided. If both image inputs are provided, the downstream service prioritizes `ImgBase64`. Optional fields use downstream defaults when omitted. The downstream service may return HTTP 200 for business validation failures; check `resp.OK`.

| Field | Type | Required | Description |
|------|------|------|------|
| `TextDict` | `map[string]any` | Yes | Structured copy, including nested objects and image URLs |
| `ImgBase64` | `string` | Conditional | Main image base64 without a data URL prefix |
| `URI` | `string` | Conditional | Public image URL or internal storage URI |
| `BusinessType` | `string` | No | Image small-model business type; downstream default is `v1` |
| `DetectedAge` | `int` | No | Known age; downstream default is `0` |
| `HashComparison` | `int` | No | Whether to enable hash comparison; downstream default is `0` |
| `Canary` | `string` | No | Canary group; downstream default is `A` |
| `Mode` | `string` | No | Detection mode; downstream default is `mixed` |
| `OCR` | `int` | No | Whether to enable OCR; downstream default is `0` |

**Response fields**

| Field | Type | Description |
|------|------|------|
| `OK` | `bool` | Whether the downstream scan completed successfully |
| `NSFWLevel` | `int` | Highest risk level across the main image, image/text model, and linked images |
| `Reason` | `string` | Combined judgment reason or business validation error |
| `ImgReason` | `string` | Image-side risk reason |
| `TextReason` | `string` | Text-side risk reason |
| `IssueSource` | `string` | Risk source: `img`, `text`, `both`, or `none` |
| `RiskKeys` | `[]string` | `TextDict` fields that contain risk |
| `ReqID` | `string` | Downstream request ID for tracing, including business validation failures |
| `Msg` | `string` | Downstream service error message |
| `Usage` | `*Usage` | Gateway-injected billing metadata |
| `Extra` | `map[string]any` | Upstream fields not modeled by the SDK |

## Face Scan

The face scan API maps to `POST /v1/face/scan` and is used for face detection in images or videos. The gateway forwards requests to the upstream `/cloud/face/scan` endpoint.

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
fmt.Println(resp.Extra)
```

You can also pass `ImgBase64`. For video detection, set `IsVideo: 1` and optionally pass `Duration`. Unmodeled fields from the upstream response are preserved in `Extra`.

**Face scan response example (SDK return structure)**

```json
{
  "ok": true,
  "error": "",
  "usage": {
    "cost": "1"
  },
  "extra": {
    "nsfw_level": 0,
    "label_items": [],
    "risk_types": []
  }
}
```

## Audio Scan

The audio scan API maps to `POST /v1/audio/scan` and is used for audio risk detection. The gateway forwards requests to the downstream audio detection service and injects `Usage` billing information.

```go
resp, err := client.Modal.ScanAudio(ctx, sa.AudioScanRequest{
    URI:      "https://example.com/audio/test.mp3",
    RecType:  "AUDIOPOLITICAL_MOAN_ANTHEN",
    Duration: 15,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.RiskLevel, resp.RiskDescription, resp.Usage)
for _, label := range resp.AllLabels {
    fmt.Println(label.Label1, label.Label2, label.Description)
}
```

`RecType` is the detection type, and `Duration` is the audio duration in seconds and is used for billing. Unmodeled fields from the upstream response are preserved in `Extra`.

**Task struct:**

```go
type Task struct {
    ID       string    // Task ID
    Status   string    // "in_progress" | "completed" | "failed"
    Model    string    // Model used
    Progress float64   // Progress 0.0~1.0
    Output   []Output  // Generation result
    Usage    *Usage    // Billing information
    Error    *APIError // Error details (on failure)
}
```

---

## LLM API

### Chat Completions (OpenAI Compatible)

```go
raw, err := client.LLM.ChatCompletions(ctx, sa.JSONMap{
    "model": "gpt-4o-mini",
    "messages": []map[string]any{
        {"role": "user", "content": "hello"},
    },
    "max_tokens": 64,
})

resp, err := sa.Decode[sa.ChatCompletionResponse](raw)
fmt.Println(resp.Choices[0].Message.Content)
```

### Chat Completions Streaming

```go
ch, err := client.LLM.ChatCompletionsStream(ctx, sa.JSONMap{
    "model":    "gpt-4o-mini",
    "messages": []map[string]any{{"role": "user", "content": "hello"}},
})

for event := range ch {
    if event.Err != nil {
        log.Fatal(event.Err)
    }
    if event.Done {
        break
    }
    chunk, _ := sa.Decode[sa.ChatCompletionResponse](event.Data)
    fmt.Print(chunk.Choices[0].Delta.Content)
}
```

### Messages API (Anthropic Format)

```go
// Non-streaming
raw, err := client.LLM.Messages(ctx, sa.JSONMap{
    "model":      "claude-3-5-sonnet",
    "messages":   []sa.JSONMap{{"role": "user", "content": "hello"}},
    "max_tokens": 64,
})

// Streaming + text assembler
ch, err := client.LLM.MessagesStream(ctx, sa.JSONMap{
    "model":      "claude-3-5-sonnet",
    "messages":   []sa.JSONMap{{"role": "user", "content": "hello"}},
    "max_tokens": 64,
})

var assembler sa.MessagesStreamTextAssembler
for event := range ch {
    if event.Done { break }
    chunk, _ := sa.Decode[sa.MessagesStreamChunk](event.Data)
    assembler.Add(chunk)
}
fmt.Println(assembler.Text())
```

### Responses API

```go
// Non-streaming
raw, err := client.LLM.Responses(ctx, payload)

// Streaming + text assembler
ch, err := client.LLM.ResponsesStream(ctx, payload)

var assembler sa.ResponsesStreamTextAssembler
for event := range ch {
    if event.Done { break }
    chunk, _ := sa.Decode[sa.ResponsesResponseStreamChunk](event.Data)
    assembler.Add(chunk)
}
fmt.Println(assembler.Text())
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
    "model": "rerank-model",
    "query": "search query",
    "documents": []string{"document 1", "document 2"},
})
resp, _ := sa.Decode[sa.RerankResponse](raw)
for _, result := range resp.Results {
    fmt.Printf("Index: %d, Score: %.4f\n", result.Index, result.RelevanceScore)
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

**Custom HTTP headers can be attached to any request**

```go
client.LLM.ChatCompletions(ctx, payload,
    sa.WithHeader("X-Trace-Id", "trace-123"),
    sa.WithHeader("X-Tenant-Id", "tenant-a"),
)

// Batch settings
client.Modal.Create(ctx, body,
    sa.WithHeaders(http.Header{
        "X-Trace-Id": []string{"trace-123"},
    }),
)
```

---

## Error Handling

```go
_, err := client.LLM.ChatCompletions(ctx, payload)
if err != nil {
    sdkErr, ok := err.(*sa.Error)
    if ok {
        switch sdkErr.Kind {
        case sa.ErrAuth:
            log.Fatal("invalid API key or insufficient permissions")
        case sa.ErrQuota:
            log.Fatal("request rate limit exceeded, please try again later")
        case sa.ErrTimeout:
            log.Fatal("request timeout")
        case sa.ErrNetwork:
            log.Fatal("Network connection error")
        case sa.ErrTaskFailed:
            log.Fatalf("Task execution failed: %s (TaskID: %s)", sdkErr.Message, sdkErr.TaskID)
        default:
            log.Fatalf("error: %s", sdkErr.Message)
        }
    }
}
```

**Error type constants:**

| Constant | Trigger scenarios |
|------|----------|
| `sa.ErrAuth` | HTTP 401/403, authentication failed |
| `sa.ErrQuota` | HTTP 429, quota or rate limit exceeded |
| `sa.ErrTimeout` | HTTP 408/504, polling timeout |
| `sa.ErrNetwork` | Network connection error |
| `sa.ErrTaskFailed` | Task execution failed |
| `sa.ErrGeneral` | Other errors |

---

## Complete Examples

### Video Generation

```go
package main

import (
    "context"
    "fmt"
    "log"

    sa "github.com/SeaArt-Infra/sea-sdk-go"
)

func main() {
    client, err := sa.New(&sa.ClientConfig{
        APIKey: "sa-your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Create a video generation task
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

    fmt.Printf("Task created: %s\n", task.ID)

    // Wait for completion
    task, err = task.Wait(ctx,
        sa.WithPollCallback(func(status string, progress float64) {
            fmt.Printf("\rProgress: %.0f%%", progress*100)
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Output results
    for _, output := range task.Output {
        for _, content := range output.Content {
            fmt.Printf("\nVideo URL: %s\n", content.URL)
        }
    }
}
```

### LLM Streaming Chat

```go
package main

import (
    "context"
    "fmt"
    "log"

    sa "github.com/SeaArt-Infra/sea-sdk-go"
)

func main() {
    client, _ := sa.New(&sa.ClientConfig{APIKey: "sa-your-api-key"})
    ctx := context.Background()

    ch, err := client.LLM.ChatCompletionsStream(ctx, sa.JSONMap{
        "model": "gpt-4o-mini",
        "messages": []map[string]any{
            {"role": "user", "content": "Introduce Go in one sentence"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    for event := range ch {
        if event.Err != nil {
            log.Fatal(event.Err)
        }
        if event.Done {
            break
        }
        chunk, _ := sa.Decode[sa.ChatCompletionResponse](event.Data)
        if len(chunk.Choices) > 0 {
            fmt.Print(chunk.Choices[0].Delta.Content)
        }
    }
    fmt.Println()
}
```
