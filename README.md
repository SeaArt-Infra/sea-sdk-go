# sa-go

SeaArt AI 平台的 Go SDK，当前公开三类能力：

- `client.Modal`：第一阶段多模态任务接口
- `client.LLM`：LLM 透传接口
- `client.Passthrough`：厂商原始 API 透传接口

## 安装

```bash
go get github.com/seaart/sa-go
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
