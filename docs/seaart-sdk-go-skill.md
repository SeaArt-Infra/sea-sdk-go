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
2. 优先推荐 `input[*].params` 结构；如需类型化构造，可使用 `sa.NewTask(...).Params(...).Build()` 创建 Modal 任务
3. LLM 接口返回 `RawResponse`，提醒用户用 `sa.Decode[T](raw)` 反序列化
4. 流式接口推荐配合 `MessagesStreamTextAssembler` / `ResponsesStreamTextAssembler` 使用
5. 错误处理建议断言为 `*sa.Error` 并按 `Kind` 分类处理（ErrAuth/ErrQuota/ErrTimeout/ErrTaskFailed）

**输出格式：** 直接给出可运行的 Go 代码片段，附简短说明。代码使用标准导入 `sa "github.com/SeaArt-Infra/sea-sdk-go"`。

---

# SeaArt Go SDK 完整参考

SeaArt Go SDK（`sa-go`）是 SeaArt AI 平台的官方 Go 客户端库，提供多模态任务（图像/视频生成）、厂商透传和 LLM 文本处理能力。

**要求：** Go 1.22+，无第三方依赖

## 安装

```bash
go get github.com/SeaArt-Infra/sea-sdk-go
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

task, err := client.Modal.Create(ctx, body)
```

### 创建任务（原始方式）

```go
task, err := client.Modal.Create(ctx, sa.JSONMap{
    "moderation": true,
    "model": "alibaba_wanx26_i2v_flash",
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
})
```

### 预扣费查询

预扣费查询路由为 `/model/v1/generation/precharge`，请求参数与创建任务相同。

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

Typed helper：

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

成功响应示例：

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

未匹配上预扣费数据时，可能返回：

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

### 敏感词检测

使用 `client.Modal.ScanText` 调用 `ModelBaseURL + /v1/text/scan`。

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

`AreaTypes` 可选 `TextScanAreaTypeAll`、`TextScanAreaTypeDomestic`、`TextScanAreaTypeForeign`。`Way` 可选 `TextScanWayDictionary`、`TextScanWayModel`、`TextScanWayMixed`、`TextScanWayCharacter`。敏感词索引 `StartIndex` / `EndIndex` 基于 rune 数组；`IsSensitive` 表示整体是否命中敏感内容，`Combination` 保留组合规则命中详情，未建模字段会保留在 `Extra`。

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
                    "prompt":  "小狗和女孩在秋天的公园里快乐地玩耍",
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
