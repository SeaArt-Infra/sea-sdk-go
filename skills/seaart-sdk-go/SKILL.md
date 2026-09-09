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

For LLM APIs, keep the selected model in the payload's top-level `model` field. The SDK serializes it in the JSON body and does not use `X-Model`; do not pass `X-Model` with `sa.WithHeader(...)` for LLM requests. Multimodal task creation and precharge continue to route their body model through `X-Model`.

## Gateway Context Headers

Gateway API calls require `x-infra-project-id`, `x-infra-af-id`, `x-infra-session-id`, `x-infra-user-id`, and `x-request-id`. These values are request-specific, so derive them from the current caller and pass them with `sa.WithHeaders(...)`. `task.Wait(...)` does not require these headers.

```go
requestContext := sa.WithHeaders(http.Header{
    "x-infra-project-id": {projectID},
    "x-infra-af-id":      {afID},
    "x-infra-session-id": {sessionID},
    "x-infra-user-id":    {userID},
    "x-request-id":       {requestID},
})
raw, err := client.LLM.ChatCompletions(ctx, sa.JSONMap{
    "model": "model-id",
    "messages": []sa.JSONMap{{"role": "user", "content": "Hello"}},
}, requestContext)
if err != nil { log.Fatal(err) }
```

Pass `requestContext` as the final argument to every direct gateway API call, such as `client.Modal.Create(ctx, body, requestContext)`, `client.Billing.Query(ctx, query, requestContext)`, and `client.LLM.ChatCompletions(ctx, payload, requestContext)`. Reserve `ClientConfig.Headers` for headers that are genuinely fixed for the client's whole lifetime.

## Multimodal Tasks

Search before choosing a model, and retrieve its model skill when exact parameter names matter:

```go
models, err := client.Modal.ListModels(ctx, sa.ModalModelSearchParams{
    Query: "image",
    Limit: 10,
}, requestContext)
if err != nil {
    log.Fatal(err)
}
fmt.Println(models.Hits)

skill, err := client.Modal.GetModelSkill(ctx, "alibaba_wanx26_i2v_flash", requestContext)
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

task, err := client.Modal.Create(ctx, body, requestContext)
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

Use `client.Modal.Precharge(ctx, body, requestContext)` before a generation request when cost estimation is required. Do not assume every model uses the `input` and `parameters` nesting: follow the result from `GetModelSkill`.

## Billing Queries

Use `client.Billing.Query(ctx, sa.BillingQuery{...}, requestContext)` for the authenticated team's cost statement. The gateway derives the team from the Bearer token, so callers must not pass `team_alias`. The default environment scope is `develop` plus `release`; set `Environment` to one of those values to select a single environment. Use `Start`, `End`, `Provider`, `CredentialName`, `ModelGroup`, `Page`, and `PageSize` for supported filters.
Use RFC3339 or date-only values for `Start`/`End`; the range is `[start, end)`, and omitted values default to the previous seven days.

## ComfyUI Quick Apps

Use `ListComfyUITemplates(ctx, templateIDs)` to retrieve parameters for the supplied template IDs, then call `CreateComfyUITask` and poll with `task.Wait`.

```go
highMemory := true
task, err := client.Modal.CreateComfyUITask(ctx, "d32kq8le878c73876j5g", []sa.ComfyUIInput{
    {Field: "image", Value: "https://image.cdn2.seaart.me/upload/input.webp"},
    {Field: "select", Value: 1},
}, &highMemory, requestContext)
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
}, requestContext)
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
}, requestContext)
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

Use the dedicated scan methods for image/video, face, audio, sensitive-word, short-text, character-copy quality, or visual-and-structured-text checks. Image and face scans accept either `URI` or `ImgBase64`; video and audio scans require `URI`. Character quality scans use `client.Modal.ScanCharacterQuality(...)` with a flat production-line A or B field set.

```go
if _, err := client.Modal.ScanText(ctx, sa.TextScanRequest{Text: "Text to check"}, requestContext); err != nil {
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
