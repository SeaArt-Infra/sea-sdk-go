# sa-go

SeaArt AI 平台的 Go SDK，当前公开三类能力：

- `client.Modal`：多模态任务接口
- `client.LLM`：大语言模型透传接口
- `client.Passthrough`：厂商原始 API 透传接口

## 安装

```bash
go get github.com/SeaArt-Infra/sea-sdk-go
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

- `baseURL`：`https://gateway.example.com`
- `modelBaseURL`：`https://gateway.example.com/model`
- `llmBaseURL`：`https://gateway.example.com/llm`
- `passthroughBaseURL`：`https://gateway.example.com/model`

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

## 多模态任务 API

第一阶段的多模态公开面只保留任务主链路：

- `client.Modal.Create(...)`
- `client.Modal.Precharge(...)`
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
}, sa.WithHeader("X-Trace-Id", "trace-123"))
if err != nil {
    log.Fatal(err)
}
fmt.Println(task.ID, task.Status)
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

类型化辅助构造：

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
    "model": "alibaba_wanx26_i2v_flash",
    "input": []map[string]any{{"params": map[string]any{"prompt": "A dog"}}},
})
if err != nil {
    log.Fatal(err)
}

task, err = task.Wait(ctx, sa.WithPollInterval(3*time.Second))
if err != nil {
    log.Fatal(err)
}
```

### 类型化辅助构造

SDK 也提供轻量的通用辅助构造器，用于构造统一输入结构：

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

不同模型的 `params` 结构可能不同。有些模型直接把模型字段平铺在 `params` 下：

```go
body := sa.NewTask("grok_imagine_image").
    Field("dash_scope", true).
    Moderation(true).
    Params(map[string]any{
        "aspect_ratio": "1:2",
        "prompt":       "Lego art version of Superman and Batman，Night scene",
        "n":            1,
        "resolution":   "1k",
    }).
    Build()

task, err := client.Modal.Create(ctx, body)
```

设计原则：

- 多模态核心层只做请求透传和任务生命周期管理
- 不在核心层维护厂商参数枚举
- 不暴露厂商专用构造器

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

可选筛选参数映射：

- `Query` 映射到查询参数 `q`
- `Input` 映射到查询参数 `input`
- `Output` 映射到查询参数 `output`
- `Type` 映射到查询参数 `type`
- `Provider` 映射到查询参数 `provider`
- `Limit` 映射到查询参数 `limit`

参数详情接口对应 `GET /v1/models/skill/{model}`，返回 Markdown 文本：

```go
skill, err := client.Modal.GetModelSkill(ctx, "alibaba_animate_anyone_detect")
if err != nil {
    log.Fatal(err)
}
fmt.Println(skill)
```

### 图片/视频鉴黄

鉴黄接口复用 `ModelBaseURL`，对应 `POST /v1/image/scan`。请求会通过网关转发到推理网关。

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
fmt.Println(resp.Data.IsSensitive)
fmt.Println(resp.Data.SensitiveWords)
fmt.Println(resp.Data.Combination)
```

`AreaTypes` 可选 `TextScanAreaTypeAll`、`TextScanAreaTypeDomestic`、`TextScanAreaTypeForeign`。`Way` 可选 `TextScanWayDictionary`、`TextScanWayModel`、`TextScanWayMixed`、`TextScanWayCharacter`。敏感词索引 `StartIndex` / `EndIndex` 基于 rune 数组；`IsSensitive` 表示整体是否命中敏感内容，`Combination` 保留组合规则命中详情，网关注入的计费信息在 `resp.Usage`。

### 人脸检测

人脸检测接口复用 `ModelBaseURL`，对应 `POST /v1/face/scan`，由网关转发到推理网关，再转发到上游 `/cloud/face/scan`。

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

## 厂商透传 API

厂商透传层保留厂商原始 API 形态。路径需要带厂商前缀，例如 `/kling/...`、`/vidu/...`、`/google/...`。

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

## 大语言模型 API

大语言模型层继续采用“请求透传 + 原始响应返回”的形式。

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
