# sa-go

SeaArt AI 平台的 Go SDK，当前公开三类能力：

- `client.Modal`：第一阶段多模态任务接口
- `client.LLM`：LLM 透传接口
- `client.Passthrough`：厂商原始 API 透传接口

## 安装

```bash
go get github.com/SeaVerseAI/sea-sdk-go
```

要求：

- Go 1.22+
- 无第三方运行时依赖

## 初始化

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey: "sa-your-api-key",
})
if err != nil {
    log.Fatal(err)
}
```

默认网关配置：

- `baseURL`: `https://gateway.example.com`
- `modelBaseURL`: `https://gateway.example.com/model`
- `llmBaseURL`: `https://gateway.example.com/llm`
- `passthroughBaseURL`: `https://gateway.example.com/model`

如果显式传入 `BaseURL`，SDK 会默认派生：

- `modelBaseURL = baseURL + "/model"`
- `llmBaseURL = baseURL + "/llm"`
- `passthroughBaseURL = modelBaseURL`

也可以分别覆盖：

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey:             "sa-your-api-key",
    BaseURL:            "https://gateway.example.com",
    ModelBaseURL:       "https://mm-gateway.example.com",
    LLMBaseURL:         "https://llm-gateway.example.com",
    PassthroughBaseURL: "https://mm-gateway.example.com",
    Timeout:            60 * time.Second,
    Project:            "my-project",
})
if err != nil {
    log.Fatal(err)
}
```

## Modal API

第一阶段的多模态公开面只保留任务主链路：

- `client.Modal.Create(...)`
- `client.Modal.Get(...)`
- `client.Modal.Wait(...)`
- `client.Modal.ListModels(...)`
- `client.Modal.SearchModels(...)`
- `client.Modal.GetModelSkill(...)`
- `client.Modal.ScanImage(...)`
- `client.Modal.ScanFace(...)`
- `task.Wait(...)`

### 原始透传请求

```go
task, err := client.Modal.Create(ctx, sa.JSONMap{
    "model": "vidu_q3_reference",
    "input": []map[string]any{
        {
            "type": "message",
            "role": "user",
            "content": []map[string]any{
                {"type": "text", "text": "cinematic shot"},
                {"type": "image_url", "url": "https://example.com/ref1.webp"},
                {"type": "image_url", "url": "https://example.com/ref2.webp"},
            },
        },
    },
    "parameters": map[string]any{
        "duration": 5,
    },
}, sa.WithHeader("X-Trace-Id", "trace-123"))
if err != nil {
    log.Fatal(err)
}
fmt.Println(task.ID, task.Status)
```

### 等待任务完成

```go
task, err := client.Modal.Wait(ctx, "task_abc123",
    sa.WithPollInterval(3*time.Second),
    sa.WithPollTimeout(5*time.Minute),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(task.Status, task.Progress)
```

也可以在创建任务后继续等待：

```go
task, err := client.Modal.Create(ctx, sa.JSONMap{
    "model": "vidu_q3_reference",
})
if err != nil {
    log.Fatal(err)
}

task, err = task.Wait(ctx, sa.WithPollInterval(3*time.Second))
if err != nil {
    log.Fatal(err)
}
```

### Typed helper

SDK 也提供极薄的通用 helper，用于构造统一输入结构：

```go
body := sa.NewTask("vidu_q3_reference").
    User(
        sa.Text("cinematic shot"),
        sa.ImageURL("https://example.com/ref1.webp"),
        sa.ImageURL("https://example.com/ref2.webp"),
    ).
    Param("duration", 5).
    Metadata("trace_id", "trace-123").
    Build()

task, err := client.Modal.Create(ctx, body)
if err != nil {
    log.Fatal(err)
}
```

设计原则：

- Modal 核心层只做请求透传和任务生命周期管理
- 不在核心层维护 provider 参数枚举
- 不暴露 provider-specific builder

### 模型列表和参数详情

列表接口复用 `ModelBaseURL`，对应 `GET /v1/models/skill/search`：

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
```

可选筛选参数：

- `Query` -> `q`
- `Input` -> `input`
- `Output` -> `output`
- `Type` -> `type`
- `Provider` -> `provider`
- `Limit` -> `limit`

参数详情接口对应 `GET /v1/models/skill/{model}`，返回 markdown 文本：

```go
skill, err := client.Modal.GetModelSkill(ctx, "alibaba_animate_anyone_detect")
if err != nil {
    log.Fatal(err)
}
fmt.Println(skill)
```

### 图片/视频鉴黄

鉴黄接口复用 `ModelBaseURL`，对应 `POST /v1/image/scan`。请求会通过 openresty 转发到 inference-gateway。

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

视频检测时设置 `IsVideo: 1`，并可传 `Duration` 用于计费：

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI:       "https://example.com/video.mp4",
    RiskTypes: []sa.ImageScanRiskType{sa.ImageScanRiskTypeErotic, sa.ImageScanRiskTypeViolent},
    IsVideo:   1,
    Duration:  12.5,
})
```

常用响应字段包括 `OK`、`NSFWLevel`、`LabelItems`、`RiskTypes`、`FrameResults` 和 `Usage`。

风险类型说明：

| 常量 | 接口值 | 说明 |
|------|--------|------|
| `sa.ImageScanRiskTypePolity` | `POLITY` | 政治敏感、公共安全等风险内容 |
| `sa.ImageScanRiskTypeErotic` | `EROTIC` | 色情、裸露、性暗示等成人内容 |
| `sa.ImageScanRiskTypeViolent` | `VIOLENT` | 暴力、血腥、武器、伤害等内容 |
| `sa.ImageScanRiskTypeChild` | `CHILD` | 儿童安全风险，尤其是儿童相关不安全或性化内容 |

### 敏感词检测

敏感词检测接口复用 `ModelBaseURL`，对应 `POST /v1/text/scan`。

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
fmt.Println(resp.Data.SensitiveWords)
```

`AreaTypes` 可选 `TextScanAreaTypeAll`、`TextScanAreaTypeDomestic`、`TextScanAreaTypeForeign`。`Way` 可选 `TextScanWayDictionary`、`TextScanWayModel`、`TextScanWayMixed`、`TextScanWayCharacter`。敏感词索引 `StartIndex` / `EndIndex` 基于 rune 数组，网关注入的计费信息在 `resp.Usage`。

### 人脸检测

人脸检测接口复用 `ModelBaseURL`，对应 `POST /v1/face/scan`，由 openresty 转发到 inference-gateway，再转发到上游 `/cloud/face/scan`。

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
```

也可以传 `ImgBase64`。视频检测时设置 `IsVideo: 1`，并可传 `Duration` 用于计费。上游人脸检测返回结构会保留在 `resp.Extra`，网关注入的计费信息在 `resp.Usage`。

## Passthrough API

Passthrough 层保留厂商原始 API 形态。路径需要带厂商前缀，例如 `/kling/...`、`/vidu/...`、`/google/...`。

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

如果要完全透传原始 JSON 字节，使用 `RequestRaw`：

```go
resp, err := client.Passthrough.RequestRaw(
    ctx,
    http.MethodPost,
    "/google/v1beta/models/gemini-2.5-flash-image:generateContent",
    []byte(`{"contents":[{"parts":[{"text":"paint a cat"}]}]}`),
)
```

当前提供：

- `Request`
- `RequestRaw`
- `Get`
- `Post`
- `Put`
- `Delete`

## LLM API

LLM 层继续采用“请求透传 + 原始响应返回”的形式。

```go
raw, err := client.LLM.ChatCompletions(ctx, sa.JSONMap{
    "model": "gpt-4o-mini",
    "messages": []map[string]any{
        {"role": "user", "content": "hello"},
    },
    "max_tokens": 64,
}, sa.WithHeader("X-Trace-Id", "trace-123"))
if err != nil {
    log.Fatal(err)
}

resp, err := sa.Decode[sa.ChatCompletionResponse](raw)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.Choices[0].Message.Content)
```

当前支持：

- `ChatCompletions`
- `ChatCompletionsStream`
- `Messages`
- `MessagesStream`
- `Responses`
- `ResponsesStream`
- `Rerank`
- `Embeddings`
- `ListModels`

## 开发命令

```bash
make fmt
make test
make vet
make check

task fmt
task test
task vet
task check
```
