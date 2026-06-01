---
name: seaart-sdk-go
description: SeaArt Go SDK 使用助手 — 帮助用户用 sa-go 调用 SeaArt AI 平台 API，包括多模态任务（图像/视频生成）、厂商透传和 LLM（对话、流式、embeddings、rerank）
type: slash_command
tags:
  - go
  - seaart
  - sdk
  - llm
  - multimodal
---

当用户触发此技能时，提供 SeaArt Go SDK（`sa-go`）的调用指导。

**触发场景：** 用户需要用 Go 调用 SeaArt API、生成图像/视频、调用 LLM 接口，或遇到 SDK 使用问题时。

**处理逻辑：**

1. 根据用户需求判断使用 Modal API（统一多模态任务）、Passthrough API（厂商原始接口）还是 LLM API（文本生成）
2. 优先推荐 Builder 方式（`sa.NewTask(...).User(...).Param(...).Build()`）创建 Modal 任务
3. LLM 接口返回 `RawResponse`，提醒用户用 `sa.Decode[T](raw)` 反序列化
4. 流式接口推荐配合 `MessagesStreamTextAssembler` / `ResponsesStreamTextAssembler` 使用
5. 错误处理建议断言为 `*sa.Error` 并按 `Kind` 分类处理（ErrAuth/ErrQuota/ErrTimeout/ErrTaskFailed）

**输出格式：** 直接给出可运行的 Go 代码片段，附简短说明。代码使用标准导入 `sa "github.com/SeaVerseAI/sea-sdk-go"`。

---

# SeaArt Go SDK 完整参考

SeaArt Go SDK（`sa-go`）是 SeaArt AI 平台的官方 Go 客户端库，提供多模态任务（图像/视频生成）、厂商透传和 LLM 文本处理能力。

**要求：** Go 1.22+，无第三方依赖

## 安装

```bash
go get github.com/SeaVerseAI/sea-sdk-go
```

## 客户端配置

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey:             "sa-your-api-key",       // 必填：SeaArt API Key
    BaseURL:            "https://custom-url.com", // 可选：自定义基础地址
    ModelBaseURL:       "https://model-url.com",  // 可选：多模态端点
    LLMBaseURL:         "https://llm-url.com",    // 可选：LLM 端点
    PassthroughBaseURL: "https://model-url.com",  // 可选：厂商透传端点，默认同 ModelBaseURL
    Project:            "my-project",            // 可选：作为 X-Project 头发送
    HTTPClient:         &http.Client{},           // 可选：自定义 HTTP 客户端
    Timeout:            60 * time.Second,         // 可选：默认 5 分钟
})
```

**默认端点：** `https://gateway.example.com`
**认证方式：** `Authorization: Bearer {apiKey}`

---

## Modal API（多模态任务）

### 创建任务（Builder 方式，推荐）

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
```

### 创建任务（原始方式）

```go
task, err := client.Modal.Create(ctx, sa.JSONMap{
    "model": "vidu_q3_reference",
    "input": []map[string]any{
        {
            "type": "message",
            "role": "user",
            "content": []map[string]any{
                {"type": "text", "text": "cinematic shot"},
                {"type": "image_url", "url": "https://example.com/ref.webp"},
            },
        },
    },
    "parameters": map[string]any{"duration": 5},
})
```

### 内容类型构造器

| 函数 | 说明 |
|------|------|
| `sa.Text(text)` | 文本内容 |
| `sa.ImageURL(url)` | 图片 URL |
| `sa.VideoURL(url)` | 视频 URL |
| `sa.AudioURL(url)` | 音频 URL |
| `sa.FileID(id)` | 文件 ID |

### 等待任务完成

```go
task, err = task.Wait(ctx,
    sa.WithPollInterval(3*time.Second),
    sa.WithPollTimeout(5*time.Minute),
    sa.WithPollCallback(func(status string, progress float64) {
        fmt.Printf("状态: %s, 进度: %.1f%%\n", status, progress*100)
    }),
)
```

**轮询选项：** 默认间隔 3s，默认超时 5 分钟。

### 获取任务结果

```go
for _, output := range task.Output {
    for _, content := range output.Content {
        fmt.Printf("类型: %s, URL: %s\n", content.Type, content.URL)
    }
}
```

**Task 状态：** `"in_progress"` / `"completed"` / `"failed"`

### 图片/视频鉴黄

使用 `client.Modal.ScanImage` 调用 `ModelBaseURL + /v1/image/scan`。

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI: "https://example.com/image.jpg",
    RiskTypes: []sa.ImageScanRiskType{
        sa.ImageScanRiskTypePolity,
        sa.ImageScanRiskTypeErotic,
        sa.ImageScanRiskTypeViolent,
        sa.ImageScanRiskTypeChild,
    },
    IsVideo: 0,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(resp.OK, resp.NSFWLevel, resp.RiskTypes)
```

视频检测设置 `IsVideo: 1`，可传 `Duration`；响应中的 `FrameResults` 包含帧级检测结果。

风险类型说明：

| 常量 | 接口值 | 说明 |
|------|--------|------|
| `sa.ImageScanRiskTypePolity` | `POLITY` | 政治敏感、公共安全等风险内容 |
| `sa.ImageScanRiskTypeErotic` | `EROTIC` | 色情、裸露、性暗示等成人内容 |
| `sa.ImageScanRiskTypeViolent` | `VIOLENT` | 暴力、血腥、武器、伤害等内容 |
| `sa.ImageScanRiskTypeChild` | `CHILD` | 儿童安全风险，尤其是儿童相关不安全或性化内容 |

### 人脸检测

使用 `client.Modal.ScanFace` 调用 `ModelBaseURL + /v1/face/scan`。网关会转发到上游 `/cloud/face/scan`。

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

也可以传 `ImgBase64`。视频检测设置 `IsVideo: 1`，可传 `Duration`；上游返回中的未建模字段会保留在 `Extra`。

---

## Passthrough API（厂商透传）

路径需要带厂商前缀，例如 `/kling/...`、`/vidu/...`、`/google/...`。

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

完全透传原始 JSON 字节时使用 `RequestRaw`：

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

### Chat Completions（OpenAI 兼容）

```go
// 非流式
raw, err := client.LLM.ChatCompletions(ctx, sa.JSONMap{
    "model":      "gpt-4o-mini",
    "messages":   []map[string]any{{"role": "user", "content": "你好"}},
    "max_tokens": 64,
})
resp, _ := sa.Decode[sa.ChatCompletionResponse](raw)
fmt.Println(resp.Choices[0].Message.Content)

// 流式
ch, err := client.LLM.ChatCompletionsStream(ctx, sa.JSONMap{
    "model":    "gpt-4o-mini",
    "messages": []map[string]any{{"role": "user", "content": "你好"}},
})
for event := range ch {
    if event.Err != nil || event.Done { break }
    chunk, _ := sa.Decode[sa.ChatCompletionResponse](event.Data)
    fmt.Print(chunk.Choices[0].Delta.Content)
}
```

### Messages API（Anthropic 格式）

```go
// 流式 + 文本组装器
ch, err := client.LLM.MessagesStream(ctx, sa.JSONMap{
    "model":      "claude-3-5-sonnet",
    "messages":   []sa.JSONMap{{"role": "user", "content": "你好"}},
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
    "input": "需要向量化的文本",
})
resp, _ := sa.Decode[sa.EmbeddingsResponse](raw)
```

### Reranking

```go
raw, err := client.LLM.Rerank(ctx, sa.JSONMap{
    "model":     "rerank-model",
    "query":     "搜索查询",
    "documents": []string{"文档1", "文档2"},
})
resp, _ := sa.Decode[sa.RerankResponse](raw)
for _, r := range resp.Results {
    fmt.Printf("Index: %d, Score: %.4f\n", r.Index, r.RelevanceScore)
}
```

### 列出可用模型

```go
raw, err := client.LLM.ListModels(ctx)
resp, _ := sa.Decode[sa.LLMModelListResponse](raw)
for _, model := range resp.Data {
    fmt.Println(model.ID)
}
```

---

## 请求选项

```go
client.LLM.ChatCompletions(ctx, payload,
    sa.WithHeader("X-Trace-Id", "abc-123"),
    sa.WithHeader("X-Tenant-Id", "tenant-a"),
)
```

---

## 错误处理

```go
if err != nil {
    if sdkErr, ok := err.(*sa.Error); ok {
        switch sdkErr.Kind {
        case sa.ErrAuth:       // 401/403 — API Key 无效
        case sa.ErrQuota:      // 429 — 超出频率限制
        case sa.ErrTimeout:    // 408/504 — 超时
        case sa.ErrNetwork:    // 网络连接错误
        case sa.ErrTaskFailed: // 任务执行失败
        default:               // sa.ErrGeneral
        }
    }
}
```

---

## 完整示例：视频生成

```go
package main

import (
    "context"
    "fmt"
    "log"

    sa "github.com/SeaVerseAI/sea-sdk-go"
)

func main() {
    client, err := sa.New(&sa.ClientConfig{APIKey: "sa-your-api-key"})
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    task, err := client.Modal.Create(ctx,
        sa.NewTask("vidu_q3_reference").
            User(
                sa.Text("一只猫在月光下奔跑，电影级画面"),
                sa.ImageURL("https://example.com/cat.jpg"),
            ).
            Param("duration", 5).
            Build(),
    )
    if err != nil {
        log.Fatal(err)
    }

    task, err = task.Wait(ctx,
        sa.WithPollCallback(func(status string, progress float64) {
            fmt.Printf("\r进度: %.0f%%", progress*100)
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    for _, output := range task.Output {
        for _, content := range output.Content {
            fmt.Printf("\n视频 URL: %s\n", content.URL)
        }
    }
}
```
