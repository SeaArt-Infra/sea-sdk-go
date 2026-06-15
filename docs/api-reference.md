# SeaArt Go SDK API 参考

本文档按公开 Go API 组织，帮助使用者快速确认方法、请求路径、参数形态和返回值。示例统一使用：

```go
import sa "github.com/SeaArt-Infra/sea-sdk-go"
```

## 客户端

使用 `sa.New` 创建客户端，建议在应用进程内复用同一个实例。

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey:             "sa-your-api-key",
    BaseURL:            "https://gateway.example.com",
    ModelBaseURL:       "https://gateway.example.com/model",
    LLMBaseURL:         "https://gateway.example.com/llm",
    PassthroughBaseURL: "https://gateway.example.com/model",
    Project:            "my-project",
    Timeout:            60 * time.Second,
    HTTPClient:         &http.Client{},
})
```

配置说明：

| 字段 | 必填 | 说明 |
|------|------|------|
| `APIKey` | 是 | 通过 `Authorization: Bearer {apiKey}` 发送。 |
| `BaseURL` | 否 | 根网关地址；设置后会派生默认 `ModelBaseURL` 和 `LLMBaseURL`。 |
| `ModelBaseURL` | 否 | 多模态任务、模型技能和内容安全接口地址。 |
| `LLMBaseURL` | 否 | LLM 接口地址。 |
| `PassthroughBaseURL` | 否 | 厂商透传接口地址；默认等于 `ModelBaseURL`。 |
| `Project` | 否 | 作为 `X-Project` 请求头发送。 |
| `Timeout` | 否 | HTTP 客户端超时；默认 5 分钟。 |
| `HTTPClient` | 否 | 自定义 HTTP 客户端；SDK 会复制一份并设置 `Timeout`。 |

默认端点：

| 配置项 | 默认值 |
|--------|--------|
| `BaseURL` | `https://gateway.example.com` |
| `ModelBaseURL` | `https://gateway.example.com/model` |
| `LLMBaseURL` | `https://gateway.example.com/llm` |
| `PassthroughBaseURL` | `https://gateway.example.com/model` |

## 请求选项

所有服务方法都支持在单次请求上追加 HTTP 头。

```go
resp, err := client.Modal.Create(ctx, body,
    sa.WithHeader("X-Trace-Id", "trace-123"),
    sa.WithHeaders(http.Header{
        "X-Tenant-Id": []string{"tenant-a"},
    }),
)
```

`WithHeader` 会覆盖同名头的值；`WithHeaders` 会逐个复制传入的 header。

## Modal API

Modal API 使用 `ModelBaseURL`。核心任务请求保持透传，SDK 只负责统一任务生命周期、常用响应结构和轮询。

| 方法 | HTTP 路径 | 说明 |
|------|-----------|------|
| `client.Modal.Create(ctx, body, opts...)` | `POST /v1/generation` | 创建多模态生成任务。 |
| `client.Modal.Precharge(ctx, body, opts...)` | `POST /v1/generation/precharge` | 查询创建任务前的预扣费信息。 |
| `client.Modal.Get(ctx, taskID, opts...)` | `GET /v1/generation/task/{taskID}` | 查询任务状态和结果。 |
| `client.Modal.Wait(ctx, taskID, opts...)` | 轮询 `GET /v1/generation/task/{taskID}` | 等待任务进入完成或失败状态。 |
| `task.Wait(ctx, opts...)` | 轮询 `GET /v1/generation/task/{taskID}` | 从 `Create` 返回的任务继续等待。 |
| `client.Modal.ListModels(ctx, params, opts...)` | `GET /v1/models/skill/search` | 搜索模型技能。 |
| `client.Modal.SearchModels(ctx, params, opts...)` | `GET /v1/models/skill/search` | `ListModels` 的同义方法。 |
| `client.Modal.GetModelSkill(ctx, model, opts...)` | `GET /v1/models/skill/{model}` | 获取模型参数说明 Markdown。 |
| `client.Modal.ScanImage(ctx, req, opts...)` | `POST /v1/image/scan` | 图片、GIF 或视频风险检测。 |
| `client.Modal.ScanText(ctx, req, opts...)` | `POST /v1/text/scan` | 文本敏感词检测。 |
| `client.Modal.ScanFace(ctx, req, opts...)` | `POST /v1/face/scan` | 图片或视频人脸检测。 |

### 创建任务

原始请求体使用 `sa.JSONMap`，字段结构与网关保持一致。

```go
task, err := client.Modal.Create(ctx, sa.JSONMap{
    "moderation": true,
    "model":      "alibaba_wanx26_i2v_flash",
    "input": []map[string]any{
        {
            "params": map[string]any{
                "input": map[string]any{
                    "img_url": "https://example.com/input.jpg",
                    "prompt":  "cinematic motion",
                },
                "parameters": map[string]any{
                    "resolution": "720P",
                    "duration":   5,
                },
            },
        },
    },
})
```

也可以使用 `NewTask` 构造通用任务体：

```go
body := sa.NewTask("alibaba_wanx26_i2v_flash").
    Moderation(true).
    Params(map[string]any{
        "input": map[string]any{
            "img_url": "https://example.com/input.jpg",
            "prompt":  "cinematic motion",
        },
        "parameters": map[string]any{
            "resolution": "720P",
            "duration":   5,
        },
    }).
    Metadata("trace_id", "trace-123").
    Option("priority", "normal").
    Build()

task, err := client.Modal.Create(ctx, body)
```

`TaskBuilder` 方法：

| 方法 | 说明 |
|------|------|
| `Params(map[string]any)` | 设置 `input[0].params`。 |
| `Param(key, value)` | 向 `params.parameters` 写入一个字段。 |
| `Metadata(key, value)` | 向顶层 `metadata` 写入一个字段。 |
| `Option(key, value)` | 向顶层 `options` 写入一个字段。 |
| `Moderation(bool)` | 设置顶层 `moderation`。 |
| `Field(key, value)` | 设置额外顶层字段，例如预扣费使用的 `id`。 |
| `Build()` | 返回可直接传给 `Create` 或 `Precharge` 的 `sa.JSONMap`。 |

`Create` 返回的 `Task` 包含：

```go
type Task struct {
    ID       string
    Status   string
    Model    string
    Progress float64
    Output   []sa.Output
    Usage    *sa.Usage
    Error    *sa.APIError
}
```

### 预扣费

`Precharge` 的请求体与 `Create` 相同。成功时常用字段在 `resp.Data` 中；当网关未匹配到费用缓存时，`Status` 可能为 `failed`，并在 `Data.Reason` 中给出原因。

```go
resp, err := client.Modal.Precharge(ctx,
    sa.NewTask("volces_seedream_4_5").
        Moderation(false).
        Field("id", "request-id").
        Params(map[string]any{"prompt": "A dog"}).
        Build(),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Status, resp.Data.Cost, resp.Data.Currency, resp.Data.Reason)
```

### 轮询

`Wait` 默认每 3 秒轮询一次，最多等待 5 分钟。

```go
task, err = task.Wait(ctx,
    sa.WithPollInterval(2*time.Second),
    sa.WithPollTimeout(3*time.Minute),
    sa.WithPollCallback(func(status string, progress float64) {
        fmt.Printf("status=%s progress=%.0f%%\n", status, progress*100)
    }),
)
```

### 模型技能

`ModalModelSearchParams` 支持下列字段：

| 字段 | 查询参数 |
|------|----------|
| `Query` | `q` |
| `Input` | `input` |
| `Output` | `output` |
| `Type` | `type` |
| `Provider` | `provider` |
| `Limit` | `limit` |

```go
models, err := client.Modal.SearchModels(ctx, sa.ModalModelSearchParams{
    Query:    "video",
    Provider: "alibaba",
    Limit:    10,
})
if err != nil {
    log.Fatal(err)
}
for _, hit := range models.Hits {
    fmt.Println(hit["model"], hit["name"])
}

skill, err := client.Modal.GetModelSkill(ctx, "alibaba_wanx26_i2v_flash")
fmt.Println(skill)
```

### 内容安全

图片、GIF 或视频风险检测：

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
```

文本敏感词检测：

```go
resp, err := client.Modal.ScanText(ctx, sa.TextScanRequest{
    Text:      "prompt to check",
    Scene:     1,
    AreaTypes: []sa.TextScanAreaType{sa.TextScanAreaTypeForeign},
    Way:       sa.TextScanWayDictionary,
})
```

人脸检测：

```go
resp, err := client.Modal.ScanFace(ctx, sa.FaceScanRequest{
    URI:     "https://example.com/image.jpg",
    IsVideo: 0,
    Scene:   "avatar",
})
```

`ScanTextResponse.Extra` 和 `ScanFaceResponse.Extra` 会保留 SDK 尚未建模的上游响应字段。

## LLM API

LLM API 使用 `LLMBaseURL`。非流式方法返回 `sa.RawResponse`，调用方可按目标响应类型解码。

| 方法 | HTTP 路径 | 返回 |
|------|-----------|------|
| `client.LLM.ChatCompletions(ctx, payload, opts...)` | `POST /chat/completions` | `sa.RawResponse` |
| `client.LLM.ChatCompletionsStream(ctx, payload, opts...)` | `POST /chat/completions` | `<-chan sa.LLMStreamEvent` |
| `client.LLM.Messages(ctx, payload, opts...)` | `POST /v1/messages` | `sa.RawResponse` |
| `client.LLM.MessagesStream(ctx, payload, opts...)` | `POST /v1/messages` | `<-chan sa.LLMStreamEvent` |
| `client.LLM.Responses(ctx, payload, opts...)` | `POST /responses` | `sa.RawResponse` |
| `client.LLM.ResponsesStream(ctx, payload, opts...)` | `POST /responses` | `<-chan sa.LLMStreamEvent` |
| `client.LLM.Rerank(ctx, payload, opts...)` | `POST /rerank` | `sa.RawResponse` |
| `client.LLM.Embeddings(ctx, payload, opts...)` | `POST /v1/embeddings` | `sa.RawResponse` |
| `client.LLM.ListModels(ctx, opts...)` | `GET /v1/models` | `sa.RawResponse` |

非流式示例：

```go
raw, err := client.LLM.ChatCompletions(ctx, sa.JSONMap{
    "model": "gpt-4o-mini",
    "messages": []map[string]any{
        {"role": "user", "content": "hello"},
    },
    "max_tokens": 64,
})
if err != nil {
    log.Fatal(err)
}

resp, err := sa.Decode[sa.ChatCompletionResponse](raw)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Choices[0].Message.Content)
```

如果 payload 中设置了 `stream: true`，请使用对应的 `*Stream` 方法；非流式方法会返回参数错误。

流式方法会自动把 `stream` 设置为 `true`，返回 Server-Sent Events：

```go
events, err := client.LLM.ResponsesStream(ctx, sa.JSONMap{
    "model": "gpt-4o-mini",
    "input": "write a haiku",
})
if err != nil {
    log.Fatal(err)
}

var text sa.ResponsesStreamTextAssembler
for event := range events {
    if event.Err != nil {
        log.Fatal(event.Err)
    }
    if event.Done {
        break
    }

    chunk, err := sa.Decode[sa.ResponsesResponseStreamChunk](event.Data)
    if err != nil {
        log.Fatal(err)
    }
    text.Add(chunk)
}
fmt.Println(text.Text())
```

常用解码类型：

| API | 解码类型 |
|-----|----------|
| Chat Completions | `sa.ChatCompletionResponse` |
| Messages | `sa.MessagesResponse` |
| Messages stream | `sa.MessagesStreamChunk` |
| Responses | `sa.ResponsesResponse` |
| Responses stream | `sa.ResponsesResponseStreamChunk` |
| Rerank | `sa.RerankResponse` |
| Embeddings | `sa.EmbeddingsResponse` |
| List models | `sa.LLMModelListResponse` |

## Passthrough API

Passthrough API 使用 `PassthroughBaseURL`，用于按厂商原始路径调用接口。路径必须是相对路径，可以带或不带开头的 `/`，不能传完整 URL。

| 方法 | 说明 |
|------|------|
| `client.Passthrough.Request(ctx, method, path, body, opts...)` | 将 `body` 编码为 JSON 后发送。 |
| `client.Passthrough.RequestRaw(ctx, method, path, body, opts...)` | 按原始字节发送请求体。 |
| `client.Passthrough.Get(ctx, path, opts...)` | GET 快捷方法。 |
| `client.Passthrough.Post(ctx, path, body, opts...)` | POST JSON 快捷方法。 |
| `client.Passthrough.Put(ctx, path, body, opts...)` | PUT JSON 快捷方法。 |
| `client.Passthrough.Delete(ctx, path, body, opts...)` | DELETE JSON 快捷方法。 |

```go
resp, err := client.Passthrough.Post(ctx, "/kling/v1/videos/text2video", sa.JSONMap{
    "model_name": "kling-v1",
    "prompt":     "cinematic shot",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.StatusCode, resp.Headers.Get("Content-Type"), string(resp.Body))
```

原始 JSON 字节透传：

```go
resp, err := client.Passthrough.RequestRaw(
    ctx,
    http.MethodPost,
    "/google/v1beta/models/gemini-2.5-flash-image:generateContent",
    []byte(`{"contents":[{"parts":[{"text":"paint a cat"}]}]}`),
)
```

返回值：

```go
type PassthroughResponse struct {
    StatusCode int
    Headers    http.Header
    Body       sa.RawResponse
}
```

## 错误处理

SDK 返回的错误通常可以断言为 `*sa.Error`。

```go
if err != nil {
    var sdkErr *sa.Error
    if errors.As(err, &sdkErr) {
        switch sdkErr.Kind {
        case sa.ErrAuth:
            log.Fatal("authentication failed")
        case sa.ErrQuota:
            log.Fatal("quota exceeded")
        case sa.ErrTimeout:
            log.Fatal("request timed out")
        case sa.ErrNetwork:
            log.Fatal("network error")
        case sa.ErrTaskFailed:
            log.Fatalf("task failed: %s", sdkErr.Message)
        default:
            log.Fatal(sdkErr.Message)
        }
    }
    log.Fatal(err)
}
```

错误类型：

| 常量 | 常见场景 |
|------|----------|
| `sa.ErrAuth` | HTTP 401 或 403。 |
| `sa.ErrQuota` | HTTP 429。 |
| `sa.ErrTimeout` | HTTP 408、504 或任务轮询超时。 |
| `sa.ErrNetwork` | 网络连接失败或流读取失败。 |
| `sa.ErrTaskFailed` | 多模态任务返回失败状态。 |
| `sa.ErrGeneral` | URL、参数、JSON 编解码等通用错误。 |

## 开发与验证

```bash
make fmt
make test
make vet
make check
```

仓库也提供等价的 Taskfile 命令：

```bash
task fmt
task test
task vet
task check
```
