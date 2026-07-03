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
| [Image/Video Safety Scan](#imagevideo-safety-scan) | `client.Modal.ScanImage(...)` | Detect content-safety risks in images, GIFs, or videos |
| [Sensitive-Word Scan](#sensitive-word-scan) | `client.Modal.ScanText(...)` | Detect sensitive words and combination-rule risks in text |
| [Text Content Safety Scan](#text-content-safety-scan) | `client.Modal.ScanTextContent(...)` | Review short text risk level and category label |
| [Face Scan](#face-scan) | `client.Modal.ScanFace(...)` | Detect face-related results in images or videos |
| [Audio Scan](#audio-scan) | `client.Modal.ScanAudio(...)` | Detect audio content risks |
| [LLM API](#llm-api) | `client.LLM` | OpenAI / Anthropic / Responses / Embeddings / Rerank compatible APIs |

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

Configure the unified gateway address through `BaseURL`. The SDK uses it to call multimodal, LLM, and passthrough capabilities.

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

The image/video safety scan endpoint is `POST /v1/image/scan`. It detects content-safety risks in images, GIFs, or videos. Provide the media URL and use `RiskTypes` to specify risk categories to detect.

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI: "https://example.com/image.jpg",
    RiskTypes: []sa.ImageScanRiskType{
        sa.ImageScanRiskTypePolity,
        sa.ImageScanRiskTypeErotic,
        sa.ImageScanRiskTypeViolent,
        sa.ImageScanRiskTypeChild,
    },
    DetectedAge: 0,
    IsVideo:     0,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.OK, resp.NSFWLevel, resp.RiskTypes)
```

Video scans are also supported:

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI:       "https://example.com/video.mp4",
    RiskTypes: []sa.ImageScanRiskType{sa.ImageScanRiskTypeErotic, sa.ImageScanRiskTypeViolent},
    IsVideo:   1,
    Duration:  12.5,
})
```

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
fmt.Println(resp.Reason)
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
| `Level` | `int` | Risk level from `0` to `6`; higher values indicate higher risk |
| `Label` | `string` | Category label in English |
| `Reason` | `string` | Judgment reason in English or error reason |
| `Usage` | `*Usage` | Gateway-injected billing metadata. `Usage.Cost` is the cost of this call |
| `Extra` | `map[string]any` | Upstream fields not modeled by the SDK |

**Pass response example**

```json
{
  "ok": true,
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
  "level": 5,
  "label": "pornography",
  "reason": "Explicit sexual description",
  "usage": {
    "cost": "0.001"
  }
}
```

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
