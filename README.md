# Sea Go SDK

Sea AI Platform Go SDK for calling multimodal, LLM, and vendor passthrough capabilities through the unified gateway.

Features:

- Standard-library implementation with no third-party runtime dependencies
- Preserves raw request passthrough capabilities
- Supports SSE streaming response parsing
- Supports task polling and a general task builder

## Feature Navigation

| Service | Client Field | Capability |
|------|-------------|------|
| [Multimodal API](#multimodal-api) | `client.Modal` | Model listing, parameter details, generation tasks, precharge estimates, and vendor passthrough |
| [ComfyUI Quick Apps](#comfyui-quick-apps) | `client.Modal.CreateComfyUITask(...)` | Query template parameters, create ComfyUI quick-app tasks, and poll results |
| [Image/Video Safety Scan](#imagevideo-safety-scan) | `client.Modal.ScanImage(...)` | Detect content-safety risks in images, GIFs, or videos |
| [Sensitive-Word Scan](#sensitive-word-scan) | `client.Modal.ScanText(...)` | Detect sensitive words and combination-rule risks in text |
| [Text Content Safety Scan](#text-content-safety-scan) | `client.Modal.ScanTextContent(...)` | Review short text risk level and category label |
| [Visual Structured Text Fusion Scan](#visual-structured-text-fusion-scan) | `client.Modal.ScanVisualStructuredTextFusion(...)` | Scan digital-human cover images and structured copy together |
| [Face Scan](#face-scan) | `client.Modal.ScanFace(...)` | Detect face-related results in images or videos |
| [Audio Scan](#audio-scan) | `client.Modal.ScanAudio(...)` | Detect audio content risks |
| [LLM API](#llm-api) | `client.LLM` | OpenAI / Anthropic / Responses / Embeddings / Rerank compatible APIs |
| [Billing API](#billing-api) | `client.Billing` | Query the authenticated team's cost statement |

## Installation

```bash
go get github.com/SeaArt-Infra/sea-sdk-go.git
```

Requirements:

- Go 1.22+

## Initialization

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey: "sa-your-api-key",
})
if err != nil {
    log.Fatal(err)
}
```

Configure the unified gateway address through `BaseURL`. The SDK uses it to call multimodal, LLM, billing, and passthrough capabilities.

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey:  "sa-your-api-key",
    BaseURL: "https://gateway.example.com",
    Timeout: 60 * time.Second,
    Project: "my-project",
})
if err != nil {
    log.Fatal(err)
}
```

## Multimodal API

### Model List and Parameter Details

```go
models, err := client.Modal.ListModels(ctx, sa.ModalModelSearchParams{
    Query: "",
    Limit: 2,
})
if err != nil {
    log.Fatal(err)
}
for _, hit := range models.Hits {
    fmt.Println(hit["name"])
}

skill, err := client.Modal.GetModelSkill(ctx, "alibaba_animate_anyone_detect")
if err != nil {
    log.Fatal(err)
}
fmt.Println(skill)
```

`ListModels` / `SearchModels` supports these query parameters:

- `Query` -> `q`
- `Input` -> `input`
- `Output` -> `output`
- `Type` -> `type`
- `Provider` -> `provider`
- `Limit` -> `limit`

### Generation Tasks

There are two common ways to create a task: pass a raw `JSONMap`, or use the `NewTask` typed helper to build the request body. Both ultimately call `client.Modal.Create(...)`.

**Option 1: Pass a raw request JSONMap**

```go
task, err := client.Modal.Create(ctx, sa.JSONMap{
    "moderation": true,
    "model":      "alibaba_wanx26_i2v_flash",
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
}, sa.WithHeader("X-Trace-Id", "trace-123"))
if err != nil {
    log.Fatal(err)
}
fmt.Println(task.ID, task.Status)
```

`moderation` is a boolean and optional. `true` enables moderation allowlisting, while `false` disables it. `params` contains model parameters, whose structure is defined by the model.

**Option 2: Build the request body with the typed helper**

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
if err != nil {
    log.Fatal(err)
}
```

**Poll results**

```go
task, err := client.Modal.Wait(
    ctx,
    "task_abc123",
    sa.WithPollInterval(3*time.Second),
    sa.WithPollTimeout(5*time.Minute),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(task.Status, task.Progress, task.URLs())
```

You can also continue waiting after creation:

```go
task, err := client.Modal.Create(ctx, sa.JSONMap{"model": "alibaba_wanx26_i2v_flash"})
if err != nil {
    log.Fatal(err)
}

task, err = task.Wait(ctx, sa.WithPollInterval(5*time.Second))
if err != nil {
    log.Fatal(err)
}
```

When a task fails, `Wait` returns `*sa.Error` with `Kind`, `TaskID`, the gateway error `Code`, and the complete gateway error `Message`.

```go
var sdkErr *sa.Error
if errors.As(err, &sdkErr) {
    fmt.Println(sdkErr.Kind, sdkErr.Code, sdkErr.Message, sdkErr.TaskID)
}
```

### ComfyUI Quick Apps

Pass template IDs to `ListComfyUITemplates` to retrieve the corresponding quick-app parameters. `CreateComfyUITask` fixes the model to `comfyui`, routes it through `X-Model`, and builds the required request envelope.

```go
specs, err := client.Modal.ListComfyUITemplates(ctx, []string{"d32kq8le878c73876j5g"})
if err != nil { log.Fatal(err) }
highMemory := true
task, err := client.Modal.CreateComfyUITask(ctx, "d32kq8le878c73876j5g", []sa.ComfyUIInput{
    {Field: "image", Value: "https://image.cdn2.seaart.me/upload/input.webp"},
    {Field: "select", Value: 1},
}, &highMemory)
if err != nil { log.Fatal(err) }
task, err = task.Wait(ctx, sa.WithPollInterval(3*time.Second), sa.WithPollTimeout(5*time.Minute))
if err != nil { log.Fatal(err) }
fmt.Println(task.URLs())
```

### Precharge Estimate

The precharge request uses the same parameters as task creation and can estimate costs in advance. It supports two common request styles: pass a raw `JSONMap`, or build the request body with the `NewTask` typed helper.

**Option 1: Pass a raw request JSONMap**

```go
resp, err := client.Modal.Precharge(ctx, sa.JSONMap{
    "id":    "d88pmute87128c73e9r0d0",
    "model": "volces_seedream_4_5",
    "input": []map[string]any{
        {
            "params": map[string]any{
                "prompt": "A dog",
            },
        },
    },
    "moderation": false,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Status)
fmt.Println(resp.Data.BillingModel, resp.Data.Cost, resp.Data.Currency)
```

**Option 2: Build the request body with the typed helper**

```go
body := sa.NewTask("volces_seedream_4_5").
    Moderation(false).
    Field("id", "d88pmute87128c73e9r0d0").
    Params(map[string]any{
        "prompt": "A dog",
    }).
    Build()

resp, err := client.Modal.Precharge(ctx, body)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Status)
fmt.Println(resp.Data.BillingModel, resp.Data.Cost, resp.Data.Currency)
```

**Response example**

```json
{
  "status": "success",
  "data": {
    "model": "volces_seedream_4_5",
    "original_model": "volces_seedream_4_5",
    "billing_model": "volces_seedream_4_5",
    "sample_count": 1,
    "cost": "0.2",
    "currency": "credit",
    "discount": 1,
    "hash": "example-hash",
    "updated_at": 1710000000
  }
}
```

### Passthrough API (Vendor Passthrough)

The passthrough layer preserves vendor-native API shapes. Paths must include a vendor prefix, such as `/kling/...`, `/vidu/...`, or `/google/...`.

**Option 1: JSON object request**

```go
resp, err := client.Passthrough.Post(ctx, "/kling/v1/videos/text2video", sa.JSONMap{
    "model_name": "kling-v1",
    "prompt":     "cinematic shot",
}, sa.WithHeader("X-Trace-Id", "trace-123"))
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.StatusCode, string(resp.Body))
```

**Option 2: Raw byte passthrough**

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
fmt.Println(resp.StatusCode, string(resp.Body))
```

Currently available:

- `Request`
- `RequestRaw`
- `Get`
- `Post`
- `Put`
- `Delete`

## Image/Video Safety Scan

The image/video safety scan endpoint is `POST /v1/image/scan`. It detects content-safety risks in images or videos. Provide either a media URL or base64 image content, and use `RiskTypes` to specify risk categories to detect.

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
```

Video scans are also supported. Video scans must use `URI` and do not support `ImgBase64`:

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

To process asynchronously, pass `CallbackURL`:

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI:             "https://example.com/image.jpg",
    CallbackURL:     "https://example.com/callback",
    CallbackContext: map[string]any{"trace_id": "trace-123"},
})
```

**Request fields**

| Field | Type | Required | Description |
|------|------|------|------|
| `URI` | `string` | Conditionally required | Image or video URL to scan. Mutually exclusive with `ImgBase64`; videos must use `URI` |
| `ImgBase64` | `string` | Conditionally required | Base64-encoded image content. Mutually exclusive with `URI`; videos are not supported |
| `IsVideo` | `bool` or `0/1` | No | Whether the file is a video. Defaults to `false` |
| `CallbackURL` | `string` | Yes for async | Callback URL after detection completes. Only HTTP/HTTPS is supported. Passing this field enables async processing |
| `CallbackContext` | `map[string]any` | No | Caller passthrough fields. The server does not parse or modify them and returns them unchanged in the callback. Maximum 16KB |
| `RiskTypes` | `[]ImageScanRiskType` | No | Risk categories to detect. If omitted, all risk types are detected |
| `DetectedAge` | `bool` or `0/1` | No | Whether to perform age detection. Defaults to `false` |
| `Canary` | `string` | No | Canary parameter. Defaults to `B` |
| `Scene` | `string` | No | Scene identifier used for label-level config lookup and metrics |
| `Duration` | `float64` | No | Video duration in seconds. Recommended for video scans |

**Pass response example**

```json
{
  "label_items": [],
  "risk_types": [],
  "usage": {
    "cost": "0.1"
  },
  "ok": true,
  "nsfw_level": 0
}
```

**Risk-hit response example**

```json
{
  "nsfw_level": 5,
  "label_items": [
    {
      "name": "erotic_sexual_body",
      "score": 98,
      "risk_type": "EROTIC"
    }
  ],
  "risk_types": ["EROTIC"],
  "usage": {
    "cost": "0.1"
  },
  "ok": false
}
```

## Sensitive-Word Scan

The sensitive-word scan endpoint is `POST /v1/text/scan`. It detects sensitive words, combination rules, and risk hits in input text.

```go
resp, err := client.Modal.ScanText(ctx, sa.TextScanRequest{
    Text:      "a cute cat sitting on the sofa",
    Scene:     1,
    AreaTypes: []sa.TextScanAreaType{sa.TextScanAreaTypeForeign},
    Way:       0,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Data.IsSensitive)
fmt.Println(resp.Data.SensitiveWords)
fmt.Println(resp.Extra)
```

**Pass response example**

```json
{
  "usage": {
    "cost": "1"
  },
  "data": {
    "sensitive_words": [],
    "combination": null,
    "is_sensitive": false
  },
  "status": {
    "msg": "success",
    "request_id": "b5ebfb02a9d11adf98b05b397bd82e9e",
    "code": 10000
  }
}
```


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

**Request fields**

| Field | Type | Required | Description |
|------|------|------|------|
| `Text` | `string` | Yes | Text to review. Corresponds to JSON field `text` |
| `Canary` | `string` | No | Canary branch. `A` means external LLM API with vLLM fallback; `B` means local vLLM |
| `Scene` | `string` | No | Business scenario identifier, such as `user_name`, `bio`, `comment`, or `seasoul` |

**Response fields**

| Field | Type | Description |
|------|------|------|
| `OK` | `bool` | Whether the review succeeded |
| `ReqID` | `string` | Downstream request ID for tracing; returned for successful reviews and downstream business validation failures |
| `Level` | `int` | Risk level from `0` to `6`; higher values indicate higher risk |
| `Label` | `string` | Category label in English |
| `Reason` | `string` | Judgment reason in English or error reason |
| `Usage` | `*Usage` | Gateway-injected billing metadata. `Usage.Cost` is the cost of this call |
| `Extra` | `map[string]any` | Upstream fields not modeled by the SDK |

**Pass response example**

```json
{
  "ok": true,
  "req_id": "da49eb3d0b4b4d2cb8a64d2c92d70f81",
  "level": 0,
  "label": "normal",
  "reason": "Neutral greeting expression",
  "usage": {
    "cost": "0.001"
  }
}
```

**Risk-hit response example**

```json
{
  "ok": true,
  "req_id": "6d3597929be847589112510af59c5d2d",
  "level": 5,
  "label": "pornography",
  "reason": "Explicit sexual description",
  "usage": {
    "cost": "0.001"
  }
}
```

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

The face scan endpoint is `POST /v1/face/scan`. It detects face-related results in images or videos. You can pass either a media URL or base64 image content.

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

**Response fields**

| Field | Type | Description |
|------|------|------|
| `OK` | `bool` | Whether the scan request completed successfully |
| `Error` | `string` | Upstream business error message. Usually empty on success |
| `Usage` | `*Usage` | Gateway-injected billing metadata |
| `Extra` | `map[string]any` | Upstream fields not modeled by the SDK, such as risk level, labels, face count, and more |

**No-face image response example (SDK return structure)**

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

**Face image response example (SDK return structure)**

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

The audio scan endpoint is `POST /v1/audio/scan`. It detects risks in audio content. Provide an accessible audio URL. `Duration` is used for billing and statistics.

```go
resp, err := client.Modal.ScanAudio(ctx, sa.AudioScanRequest{
    URI:      "https://example.com/audio/test.mp3",
    RecType:  "AUDIOPOLITICAL_MOAN_ANTHEN",
    Duration: 15,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.RiskLevel, resp.AllLabels)
fmt.Println(resp.Extra)
```

**Pass response example**

```json
{
  "code": 1100,
  "message": "success",
  "requestId": "a63b89046c70435a4fb9a0d36439d0ee",
  "btId": "https://example.com/audio/sample.mp3",
  "detail": {
    "audioDetail": [],
    "audioTags": {},
    "audioText": "sample audio transcription text",
    "audioTime": 4,
    "code": 1100,
    "requestParams": {},
    "riskLevel": "PASS"
  }
}
```

## LLM API

LLM methods are synchronous and return raw bytes. Use `sa.Decode[T](raw)` to deserialize them.

```go
raw, err := client.LLM.ChatCompletions(ctx, sa.JSONMap{
    "model": "gpt-4o-mini",
    "messages": []map[string]string{
        {"role": "user", "content": "hello"},
    },
    "max_tokens": 64,
})
if err != nil {
    log.Fatal(err)
}

resp, err := sa.Decode[map[string]any](raw)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp)
```

Currently supported methods:

| Method | Description |
|------|------|
| `ChatCompletions` | Calls the OpenAI-compatible Chat Completions API and returns raw response bytes |
| `ChatCompletionsStream` | Calls the Chat Completions streaming API and returns a channel carrying SSE streaming events |
| `Messages` | Calls the Anthropic Messages-compatible API and returns raw response bytes |
| `MessagesStream` | Calls the Messages streaming API and returns a channel carrying SSE streaming events |
| `Responses` | Calls the OpenAI-compatible Responses API and returns raw response bytes |
| `ResponsesStream` | Calls the Responses streaming API and returns a channel carrying SSE streaming events |
| `Rerank` | Calls the text reranking API |
| `Embeddings` | Calls the embedding generation API |
| `ListModels` | Queries the LLM model list |

Streaming methods return a channel carrying SSE streaming events:

```go
events, err := client.LLM.ChatCompletionsStream(ctx, sa.JSONMap{
    "model": "gpt-4o-mini",
    "messages": []map[string]string{
        {"role": "user", "content": "hello"},
    },
})
if err != nil {
    log.Fatal(err)
}
for event := range events {
    if event.Done {
        break
    }
    fmt.Println(string(event.Data))
}
```

<script
  type="text/plain"
  data-doc-skill
  data-doc-skill-id="seaart-sdk-go"
  data-doc-skill-label="SeaArt Go SDK"
  data-doc-skill-filename="seaart-sdk-go-SKILL.md"
  data-doc-skill-version="1"
>
---
name: seaart-sdk-go
description: Build and troubleshoot SeaArt AI gateway integrations with the sea-sdk-go client. Use when generating images or videos, calling ComfyUI quick-app templates, searching model skills, estimating multimodal task cost, calling vendor-native passthrough APIs, running media or text safety scans, or using OpenAI- or Anthropic-compatible LLM, streaming, embedding, or rerank APIs from Go.
---

# SeaArt Go SDK

Use `github.com/SeaArt-Infra/sea-sdk-go` to call the SeaArt unified gateway from Go 1.22+. Import it as `sa`. The SDK is synchronous except for its streaming channels and uses only the standard library.

## Install

```bash
go get github.com/SeaArt-Infra/sea-sdk-go
```

## Workflow

1. Create one `sa.Client` with `sa.New` and reuse it across requests.
2. Select `client.Modal` for generation, model skills, precharge, or safety scans; `client.Billing` for team-scoped cost statements; `client.LLM` for LLM APIs; and `client.Passthrough` for vendor-native paths.
3. For a multimodal model, retrieve `client.Modal.GetModelSkill` before building model-specific parameters.
4. Poll generation tasks with `task.Wait`, checking both the returned task and error.
5. Decode successful LLM responses or stream event data with `sa.Decode[T]`; inspect `*sa.Error` at the request boundary.

## Initialize Client

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey:  "sa-your-api-key",
    BaseURL: "https://gateway.example.com", // optional
    Project: "my-project",                  // optional X-Project header
    Timeout: 60 * time.Second,
})
if err != nil {
    log.Fatal(err)
}
```

Passing `BaseURL` derives `/model` and `/llm` service URLs. Override `ModelBaseURL`, `LLMBaseURL`, or `PassthroughBaseURL` only when services use separate gateways. Do not expose API keys in source control or logs.

## Billing API

`client.Billing.Query` calls `GET /monitor/api/v1/cost/billing`. The gateway derives the team from the Bearer token and injects `X-User-ID`; callers do not pass a team identifier. By default the query covers `develop` and `release`. Set `Environment` to `develop` or `release` to select one environment.

`Start` and `End` define the time range. They accept RFC3339 timestamps such as `2026-08-19T00:00:00Z`, UTC date-times without a zone, date-only values such as `2026-08-19`, or Unix seconds. The range is `[start, end)`. When `End` is date-only, that whole day is included; without either value, the server defaults to the previous seven days.

```go
statement, err := client.Billing.Query(ctx, sa.BillingQuery{
    Start: "2026-08-19T00:00:00Z",
    End: "2026-08-20T00:00:00Z",
    Environment: "release",
    Page: 1,
    PageSize: 20,
})
if err != nil { log.Fatal(err) }
fmt.Println(statement.Team, statement.Summary.TotalCost)
for _, item := range statement.Items.Items {
    fmt.Println(item.Provider, item.ModelGroup, item.TotalCost)
}
```

Set `BillingBaseURL` only when the billing route is hosted separately; otherwise `BaseURL` derives it as `<BaseURL>/monitor`.

Keep the selected model in the SDK payload's top-level `model` field. The SDK sends it as the `X-Model` header and removes it from the serialized JSON body. Do not pass `X-Model` with `sa.WithHeader(...)` when the payload already contains `model`.

## Multimodal Tasks

Search before choosing a model, and retrieve its model skill when exact parameter names matter:

```go
models, err := client.Modal.ListModels(ctx, sa.ModalModelSearchParams{
    Query: "image",
    Limit: 10,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(models.Hits)

skill, err := client.Modal.GetModelSkill(ctx, "alibaba_wanx26_i2v_flash")
if err != nil {
    log.Fatal(err)
}
fmt.Println(skill)
```

Pass the documented model parameters in `input[*].params`, or build the same payload with `sa.NewTask(...)`:

```go
body := sa.NewTask("alibaba_wanx26_i2v_flash").
    Moderation(true).
    Params(map[string]any{
        "input": map[string]any{
            "img_url": "https://example.com/input.jpg",
            "prompt":  "A cinematic mountain sunrise",
        },
        "parameters": map[string]any{"resolution": "720P", "duration": 5},
    }).
    Build()

task, err := client.Modal.Create(ctx, body)
if err != nil {
    log.Fatal(err)
}
task, err = task.Wait(ctx, sa.WithPollInterval(3*time.Second), sa.WithPollTimeout(5*time.Minute))
if err != nil {
    log.Fatal(err)
}
for _, output := range task.Output {
    for _, content := range output.Content {
        fmt.Println(content.URL)
    }
}
```

Use `client.Modal.Precharge(ctx, body)` before a generation request when cost estimation is required. Do not assume every model uses the `input` and `parameters` nesting: follow the result from `GetModelSkill`.

## Billing Queries

Use `client.Billing.Query(ctx, sa.BillingQuery{...})` for the authenticated team's cost statement. The gateway derives the team from the Bearer token, so callers must not pass `team_alias`. The default environment scope is `develop` plus `release`; set `Environment` to one of those values to select a single environment. Use `Start`, `End`, `Provider`, `CredentialName`, `ModelGroup`, `Page`, and `PageSize` for supported filters.
Use RFC3339 or date-only values for `Start`/`End`; the range is `[start, end)`, and omitted values default to the previous seven days.

## ComfyUI Quick Apps

Use `ListComfyUITemplates(ctx, templateIDs)` to retrieve parameters for the supplied template IDs, then call `CreateComfyUITask` and poll with `task.Wait`.

```go
highMemory := true
task, err := client.Modal.CreateComfyUITask(ctx, "d32kq8le878c73876j5g", []sa.ComfyUIInput{
    {Field: "image", Value: "https://image.cdn2.seaart.me/upload/input.webp"},
    {Field: "select", Value: 1},
}, &highMemory)
if err != nil { log.Fatal(err) }
task, err = task.Wait(ctx, sa.WithPollInterval(3*time.Second), sa.WithPollTimeout(5*time.Minute))
if err != nil { log.Fatal(err) }
fmt.Println(task.URLs())
```

## LLM And Streaming APIs

Non-streaming LLM methods return `sa.RawResponse`. Deserialize them to the matching SDK type:

```go
raw, err := client.LLM.ChatCompletions(ctx, sa.JSONMap{
    "model": "gpt-4o-mini",
    "messages": []sa.JSONMap{{"role": "user", "content": "Hello"}},
})
if err != nil {
    log.Fatal(err)
}
response, err := sa.Decode[sa.ChatCompletionResponse](raw)
if err != nil {
    log.Fatal(err)
}
fmt.Println(response.Choices[0].Message.Content)
```

Use the dedicated streaming methods rather than setting `stream: true` on non-streaming methods. Stop on `Done` and surface an event error:

```go
events, err := client.LLM.ChatCompletionsStream(ctx, sa.JSONMap{
    "model": "gpt-4o-mini",
    "messages": []sa.JSONMap{{"role": "user", "content": "Hello"}},
})
if err != nil {
    log.Fatal(err)
}
for event := range events {
    if event.Err != nil {
        log.Fatal(event.Err)
    }
    if event.Done {
        break
    }
    chunk, err := sa.Decode[sa.ChatCompletionResponse](event.Data)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(chunk.Choices[0].Delta.Content)
}
```

Use `client.LLM.Messages` / `MessagesStream` for Anthropic Messages and `Responses` / `ResponsesStream` for OpenAI Responses. Accumulate their text with `sa.MessagesStreamTextAssembler` or `sa.ResponsesStreamTextAssembler`. Use `Embeddings`, `Rerank`, and `ListModels` for their corresponding LLM endpoints.

## Passthrough, Scans, And Errors

Use passthrough only for a vendor-native path such as `/kling/...`, `/vidu/...`, or `/google/...`; pass a relative path and preserve the returned status, headers, and raw body.

Use the dedicated scan methods for image/video, face, audio, sensitive-word, short-text, or visual-and-structured-text checks. Image and face scans accept either `URI` or `ImgBase64`; video and audio scans require `URI`.

```go
if _, err := client.Modal.ScanText(ctx, sa.TextScanRequest{Text: "Text to check"}); err != nil {
    var seaErr *sa.Error
    if errors.As(err, &seaErr) {
        switch seaErr.Kind {
        case sa.ErrAuth, sa.ErrQuota, sa.ErrTimeout:
            return err
        }
    }
    return err
}
```

Handle `sa.ErrAuth`, `sa.ErrQuota`, `sa.ErrTimeout`, `sa.ErrNetwork`, and `sa.ErrTaskFailed` explicitly where retries or user feedback differ. For failed multimodal tasks, inspect `sa.Error.TaskID` and the model response before retrying.
</script>
