# SeaArt Go SDK 使用说明文档

SeaArt Go SDK（`sa-go`）是 SeaArt AI 平台的官方 Go 客户端库，提供多模态任务（图像/视频生成）、厂商透传和 LLM 文本处理能力。

**要求：** Go 1.22+，无第三方依赖

---

## 安装

```bash
go get github.com/SeaArt-Infra/sea-sdk-go
```

---

## 快速开始

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

**默认端点：**
- Base: `https://gateway.example.com`
- 认证方式: `Authorization: Bearer {apiKey}`

---

## Modal API（多模态任务）

用于图像生成、视频生成等多模态 AI 任务。

### 创建任务（原始方式）

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
// 方式一：直接在 task 对象上等待
task, err = task.Wait(ctx,
    sa.WithPollInterval(3*time.Second),
    sa.WithPollTimeout(5*time.Minute),
    sa.WithPollCallback(func(status string, progress float64) {
        fmt.Printf("状态: %s, 进度: %.1f%%\n", status, progress*100)
    }),
)

// 方式二：通过 client 等待
task, err = client.Modal.Wait(ctx, task.ID)
```

**轮询选项：**

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `sa.WithPollInterval(d)` | 轮询间隔 | 3s |
| `sa.WithPollTimeout(d)` | 最大等待时间 | 5 分钟 |
| `sa.WithPollCallback(fn)` | 进度回调 | - |

### 获取任务结果

```go
if task.Status == "completed" {
    for _, output := range task.Output {
        for _, content := range output.Content {
            fmt.Printf("类型: %s, URL: %s\n", content.Type, content.URL)
        }
    }
}
```

### 图片/视频鉴黄

鉴黄接口走 `ModelBaseURL`，对应 `POST /v1/image/scan`，用于图片、GIF 或视频风险检测。

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
for _, label := range resp.LabelItems {
    fmt.Println(label.Name, label.Score, label.RiskType)
}
```

视频检测设置 `IsVideo: 1`，可传 `Duration`：

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI:       "https://example.com/video.mp4",
    RiskTypes: []sa.ImageScanRiskType{sa.ImageScanRiskTypeErotic, sa.ImageScanRiskTypeViolent},
    IsVideo:   1,
    Duration:  12.5,
})
```

风险类型说明：

| 常量 | 接口值 | 说明 |
|------|--------|------|
| `sa.ImageScanRiskTypePolity` | `POLITY` | 政治敏感、公共安全等风险内容 |
| `sa.ImageScanRiskTypeErotic` | `EROTIC` | 色情、裸露、性暗示等成人内容 |
| `sa.ImageScanRiskTypeViolent` | `VIOLENT` | 暴力、血腥、武器、伤害等内容 |
| `sa.ImageScanRiskTypeChild` | `CHILD` | 儿童安全风险，尤其是儿童相关不安全或性化内容 |

### 敏感词检测

敏感词检测接口走 `ModelBaseURL`，对应 `POST /v1/text/scan`。

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

人脸检测接口走 `ModelBaseURL`，对应 `POST /v1/face/scan`，用于图片或视频人脸检测。网关会转发到上游 `/cloud/face/scan`。

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

也可以传 `ImgBase64`。视频检测设置 `IsVideo: 1`，可传 `Duration`。上游返回中的未建模字段会保留在 `Extra`。

**Task 结构体：**

```go
type Task struct {
    ID       string    // 任务 ID
    Status   string    // "in_progress" | "completed" | "failed"
    Model    string    // 使用的模型
    Progress float64   // 进度 0.0~1.0
    Output   []Output  // 生成结果
    Usage    *Usage    // 计费信息
    Error    *APIError // 错误详情（失败时）
}
```

---

## Passthrough API（厂商透传）

用于调用厂商原始 API 形态的接口，路径需要带厂商前缀，例如 `/kling/...`、`/vidu/...`、`/google/...`。

### JSON 请求

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

### 原始请求体透传

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

`PassthroughResponse` 会保留响应状态码、响应头和原始 body：

```go
type PassthroughResponse struct {
    StatusCode int
    Headers    http.Header
    Body       sa.RawResponse
}
```

---

## LLM API

### Chat Completions（OpenAI 兼容）

```go
raw, err := client.LLM.ChatCompletions(ctx, sa.JSONMap{
    "model": "gpt-4o-mini",
    "messages": []map[string]any{
        {"role": "user", "content": "你好"},
    },
    "max_tokens": 64,
})

resp, err := sa.Decode[sa.ChatCompletionResponse](raw)
fmt.Println(resp.Choices[0].Message.Content)
```

### Chat Completions 流式

```go
ch, err := client.LLM.ChatCompletionsStream(ctx, sa.JSONMap{
    "model":    "gpt-4o-mini",
    "messages": []map[string]any{{"role": "user", "content": "你好"}},
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

### Messages API（Anthropic 格式）

```go
// 非流式
raw, err := client.LLM.Messages(ctx, sa.JSONMap{
    "model":      "claude-3-5-sonnet",
    "messages":   []sa.JSONMap{{"role": "user", "content": "你好"}},
    "max_tokens": 64,
})

// 流式 + 文本组装器
ch, err := client.LLM.MessagesStream(ctx, sa.JSONMap{
    "model":      "claude-3-5-sonnet",
    "messages":   []sa.JSONMap{{"role": "user", "content": "你好"}},
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
// 非流式
raw, err := client.LLM.Responses(ctx, payload)

// 流式 + 文本组装器
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
    "input": "需要向量化的文本",
})
resp, _ := sa.Decode[sa.EmbeddingsResponse](raw)
```

### Reranking

```go
raw, err := client.LLM.Rerank(ctx, sa.JSONMap{
    "model": "rerank-model",
    "query": "搜索查询",
    "documents": []string{"文档1", "文档2"},
})
resp, _ := sa.Decode[sa.RerankResponse](raw)
for _, result := range resp.Results {
    fmt.Printf("Index: %d, Score: %.4f\n", result.Index, result.RelevanceScore)
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

可对任意请求附加自定义 HTTP 头：

```go
client.LLM.ChatCompletions(ctx, payload,
    sa.WithHeader("X-Trace-Id", "trace-123"),
    sa.WithHeader("X-Tenant-Id", "tenant-a"),
)

// 批量设置
client.Modal.Create(ctx, body,
    sa.WithHeaders(http.Header{
        "X-Trace-Id": []string{"trace-123"},
    }),
)
```

---

## 错误处理

```go
_, err := client.LLM.ChatCompletions(ctx, payload)
if err != nil {
    sdkErr, ok := err.(*sa.Error)
    if ok {
        switch sdkErr.Kind {
        case sa.ErrAuth:
            log.Fatal("API Key 无效或无权限")
        case sa.ErrQuota:
            log.Fatal("请求频率超限，请稍后重试")
        case sa.ErrTimeout:
            log.Fatal("请求超时")
        case sa.ErrNetwork:
            log.Fatal("网络连接错误")
        case sa.ErrTaskFailed:
            log.Fatalf("任务执行失败: %s (TaskID: %s)", sdkErr.Message, sdkErr.TaskID)
        default:
            log.Fatalf("错误: %s", sdkErr.Message)
        }
    }
}
```

**错误类型常量：**

| 常量 | 触发场景 |
|------|----------|
| `sa.ErrAuth` | HTTP 401/403，认证失败 |
| `sa.ErrQuota` | HTTP 429，超出配额/频率限制 |
| `sa.ErrTimeout` | HTTP 408/504，轮询超时 |
| `sa.ErrNetwork` | 网络连接错误 |
| `sa.ErrTaskFailed` | 任务执行失败 |
| `sa.ErrGeneral` | 其他错误 |

---

## 完整示例

### 视频生成

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

    // 创建视频生成任务
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

    fmt.Printf("任务已创建: %s\n", task.ID)

    // 等待完成
    task, err = task.Wait(ctx,
        sa.WithPollCallback(func(status string, progress float64) {
            fmt.Printf("\r进度: %.0f%%", progress*100)
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    // 输出结果
    for _, output := range task.Output {
        for _, content := range output.Content {
            fmt.Printf("\n视频 URL: %s\n", content.URL)
        }
    }
}
```

### LLM 流式对话

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
            {"role": "user", "content": "用一句话介绍 Go 语言"},
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
