# Sea Go SDK

Sea AI 平台 Go SDK，用于通过统一网关调用多模态、LLM 和厂商透传能力。

特点：

- 纯标准库实现，无第三方运行时依赖
- 保留原始请求透传能力
- 支持 SSE 流式响应解析
- 支持任务轮询和通用 task builder

## 功能导航

| 服务 | Client 字段 | 功能 |
|------|-------------|------|
| [多模态 API](#多模态-api) | `client.Modal` | 模型列表、参数详情、生成任务、预扣费查询和厂商透传 |
| [图片/视频鉴黄](#图片视频鉴黄) | `client.Modal.ScanImage(...)` | 检测图片、GIF 或视频内容安全风险 |
| [敏感词检测](#敏感词检测) | `client.Modal.ScanText(...)` | 检测文本敏感词和组合词风险 |
| [文本内容安全审核](#文本内容安全审核) | `client.Modal.ScanTextContent(...)` | 审核短文本内容安全风险等级和分类标签 |
| [人脸检测](#人脸检测) | `client.Modal.ScanFace(...)` | 检测图片或视频中的人脸相关结果 |
| [音频检测](#音频检测) | `client.Modal.ScanAudio(...)` | 检测音频内容风险 |
| [大语言模型 API](#大语言模型-api) | `client.LLM` | OpenAI / Anthropic / Responses / Embeddings / Rerank 等兼容接口 |

## 安装

```bash
go get github.com/SeaArt-Infra/sea-sdk-go.git
```

要求：

- Go 1.22+

## 初始化

```go
client, err := sa.New(&sa.ClientConfig{
    APIKey: "sa-your-api-key",
})
if err != nil {
    log.Fatal(err)
}
```

通过 `BaseURL` 配置统一网关地址，SDK 会基于该地址调用多模态、LLM 和透传等能力。

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

## 多模态 API

### 模型列表和参数详情

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

`ListModels` / `SearchModels` 支持的查询参数：

- `Query` -> `q`
- `Input` -> `input`
- `Output` -> `output`
- `Type` -> `type`
- `Provider` -> `provider`
- `Limit` -> `limit`

### 生成任务

创建任务有两种常用方式：直接传入原始请求 `JSONMap`，或使用 `NewTask` typed helper 构造请求体。两种方式最终都会调用 `client.Modal.Create(...)`。

**方式一：直接传入原始请求 JSONMap**

```go
task, err := client.Modal.Create(ctx, sa.JSONMap{
    "moderation": true,
    "model":      "alibaba_wanx26_i2v_flash",
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
}, sa.WithHeader("X-Trace-Id", "trace-123"))
if err != nil {
    log.Fatal(err)
}
fmt.Println(task.ID, task.Status)
```

`moderation` 为布尔类型，非必传；`true` 表示开白，`false` 表示非开白。`params` 为模型参数，具体结构由模型定义决定。

**方式二：使用 Typed helper 构造请求体**

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
if err != nil {
    log.Fatal(err)
}
```

**轮询结果**

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

也可以在创建后继续等待：

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

### 预扣费查询

预扣费查询请求参数与创建任务相同，可用于提前预估费用。支持两种常用方式：直接传入原始请求 `JSONMap`，或使用 `NewTask` typed helper 构造请求体。

**方式一：直接传入原始请求 JSONMap**

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

**方式二：使用 Typed helper 构造请求体**

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

**响应示例**

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

### Passthrough API（厂商透传）

Passthrough 层保留厂商原始 API 形态。路径需要带厂商前缀，例如 `/kling/...`、`/vidu/...`、`/google/...`。

**方式一：JSON object 请求**

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

**方式二：原始字节透传**

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

当前提供：

- `Request`
- `RequestRaw`
- `Get`
- `Post`
- `Put`
- `Delete`

## 图片/视频鉴黄

图片/视频鉴黄接口对应 `POST /v1/image/scan`，用于对图片、GIF 或视频内容进行安全风险检测。调用时需要提供待检测媒体 URL，并通过 `RiskTypes` 指定需要检测的风险类型。

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

也支持视频检测：

```go
resp, err := client.Modal.ScanImage(ctx, sa.ImageScanRequest{
    URI:       "https://example.com/video.mp4",
    RiskTypes: []sa.ImageScanRiskType{sa.ImageScanRiskTypeErotic, sa.ImageScanRiskTypeViolent},
    IsVideo:   1,
    Duration:  12.5,
})
```

**审核通过响应示例**

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

**命中风险响应示例**

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

## 敏感词检测

敏感词检测接口对应 `POST /v1/text/scan`，用于检测输入文本中的敏感词、组合词和风险命中结果。

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

**审核通过响应示例**

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


## 文本内容安全审核

文本内容安全审核接口对应 `POST /v1/text/content/scan`，用于对短文本进行内容安全审核，返回风险等级、分类标签和判定理由。该接口不影响旧敏感词检测接口 `POST /v1/text/scan`。

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

**请求字段**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Text` | `string` | 是 | 待审核文本，对应 JSON 字段 `text` |
| `Canary` | `string` | 否 | 灰度分支，`A` 表示外部 LLM API 失败降级 vLLM，`B` 表示本地 vLLM |
| `Scene` | `string` | 否 | 业务场景标识，例如 `user_name`、`bio`、`comment`、`seasoul` |

**响应字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `OK` | `bool` | 是否审核成功 |
| `Level` | `int` | 风险等级，范围 `0-6`，数值越大风险越高 |
| `Label` | `string` | 分类标签，英文 |
| `Reason` | `string` | 判定理由，英文或错误原因 |
| `Usage` | `*Usage` | 网关注入的计费信息，`Usage.Cost` 为本次调用费用 |
| `Extra` | `map[string]any` | 上游返回的未建模字段 |

**审核通过响应示例**

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

**命中风险响应示例**

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

## 人脸检测

人脸检测接口对应 `POST /v1/face/scan`，用于检测图片或视频中的人脸相关结果。调用时可以传入媒体 URL，也可以传入图片 base64 内容。

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

**响应字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `OK` | `bool` | 检测请求是否成功完成 |
| `Error` | `string` | 上游业务错误信息；成功时通常为空 |
| `Usage` | `*Usage` | 网关注入的计费信息 |
| `Extra` | `map[string]any` | 上游返回的未建模字段，例如风险等级、标签、人脸数量等 |

**不含人脸图片响应示例（SDK 返回结构）**

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

**含人脸图片响应示例（SDK 返回结构）**

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

## 音频检测

音频检测接口对应 `POST /v1/audio/scan`，用于检测音频内容风险。调用时需要提供可访问的音频 URL，`Duration` 用于计费和统计。

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

**审核通过响应示例**

```json
{
  "code": 1100,
  "message": "成功",
  "requestId": "a63b89046c70435a4fb9a0d36439d0ee",
  "btId": "https://example.com/audio/sample.mp3",
  "detail": {
    "audioDetail": [],
    "audioTags": {},
    "audioText": "示例音频转写文本",
    "audioTime": 4,
    "code": 1100,
    "requestParams": {},
    "riskLevel": "PASS"
  }
}
```

## 大语言模型 API

LLM 方法均为同步调用，返回原始字节，使用 `sa.Decode[T](raw)` 反序列化。

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

当前支持的方法：

| 方法 | 说明 |
|------|------|
| `ChatCompletions` | 调用 OpenAI Chat Completions 兼容接口，返回原始响应字节 |
| `ChatCompletionsStream` | 调用 Chat Completions 流式接口，返回承载 SSE 流式事件的 channel |
| `Messages` | 调用 Anthropic Messages 兼容接口，返回原始响应字节 |
| `MessagesStream` | 调用 Messages 流式接口，返回承载 SSE 流式事件的 channel |
| `Responses` | 调用 OpenAI Responses 兼容接口，返回原始响应字节 |
| `ResponsesStream` | 调用 Responses 流式接口，返回承载 SSE 流式事件的 channel |
| `Rerank` | 调用文本重排接口 |
| `Embeddings` | 调用向量生成接口 |
| `ListModels` | 查询 LLM 模型列表 |

流式方法返回承载 SSE 流式事件的 channel：

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
