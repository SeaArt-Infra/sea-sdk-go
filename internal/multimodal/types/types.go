package types

import (
	"encoding/json"
	"time"
)

// GenerateRequest is the fully-built request object returned by model builders.
type GenerateRequest struct {
	Model      string         `json:"model"`
	DashScope  bool           `json:"dash_scope"`
	Moderation bool           `json:"moderation"`
	Input      []InputItem    `json:"input"`
	Metadata   map[string]any `json:"metadata"`
}

// InputItem represents one element of the input array.
type InputItem struct {
	Content []ContentItem  `json:"content,omitempty"`
	Params  map[string]any `json:"params"`
	SRInfo  *SRInfo        `json:"sr_info,omitempty"`
}

// ContentItem is a media reference passed as input.
type ContentItem struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// SRInfo enables Tencent super-resolution post-processing.
type SRInfo struct {
	Enable           bool   `json:"enable"`
	InputResolution  string `json:"input_resolution,omitempty"`
	OutputResolution string `json:"output_resolution,omitempty"`
}

// GenerationResponse is returned by POST /v1/generation.
type GenerationResponse struct {
	ID        string    `json:"id"`
	CreatedAt int64     `json:"created_at"`
	Status    string    `json:"status"`
	Model     string    `json:"model"`
	Error     *APIError `json:"error,omitempty"`
}

// TaskResponse is returned by GET /v1/generation/task/{id}.
type TaskResponse struct {
	ID        string        `json:"id"`
	Status    string        `json:"status"`
	Progress  float64       `json:"progress,omitempty"`
	CreatedAt int64         `json:"created_at"`
	Model     string        `json:"model"`
	Output    []OutputItem  `json:"output,omitempty"`
	Usage     *Usage        `json:"usage,omitempty"`
	Metadata  *TaskMetadata `json:"metadata,omitempty"`
	Error     *APIError     `json:"error,omitempty"`
}

type OutputItem struct {
	Content []OutputContent `json:"content,omitempty"`
}

type OutputContent struct {
	JobID string `json:"jobId,omitempty"`
	Type  string `json:"type,omitempty"`
	URL   string `json:"url,omitempty"`
}

type Usage struct {
	Cost           json.Number `json:"cost"`
	Discount       float64     `json:"discount"`
	Used           *int        `json:"used,omitempty"`
	ModelBatchUUID string      `json:"model_batch_uuid,omitempty"`
	TimePerUnit    float64     `json:"time_per_unit,omitempty"`
	InputTokens    *int        `json:"input_tokens,omitempty"`
	OutputTokens   *int        `json:"output_tokens,omitempty"`
	TotalTokens    *int        `json:"total_tokens,omitempty"`
}

func (u *Usage) CostFloat64() float64 {
	if u == nil {
		return 0
	}
	f, _ := u.Cost.Float64()
	return f
}

type TaskMetadata struct {
	CompletedAt float64 `json:"completed_at,omitempty"`
	InQueueAt   float64 `json:"in_queue_at,omitempty"`
	UploadAt    float64 `json:"upload_at,omitempty"`
}

type APIError struct {
	Code         int    `json:"code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

func (e *APIError) Error() string {
	if e.ErrorMessage != "" {
		return e.ErrorMessage
	}
	return "unknown API error"
}

func (t *TaskResponse) URLs() []string {
	var urls []string
	for _, out := range t.Output {
		for _, c := range out.Content {
			if c.URL != "" {
				urls = append(urls, c.URL)
			}
		}
	}
	return urls
}

type PriceRequest struct {
	Model string      `json:"model"`
	Input []InputItem `json:"input"`
}

type PriceResponse struct {
	ID        string  `json:"id"`
	Model     string  `json:"model"`
	Cost      float64 `json:"cost"`
	Discount  float64 `json:"discount"`
	CreatedAt int64   `json:"created_at"`
}

type ModerationRequest struct {
	URI     string `json:"uri"`
	IsVideo int    `json:"is_video"`
}

type ModerationResponse struct {
	OK         bool              `json:"ok"`
	NSFWLevel  int               `json:"nsfw_level"`
	LabelItems []ModerationLabel `json:"label_items"`
	RiskTypes  []string          `json:"risk_types"`
}

type ModerationLabel struct {
	Name     string `json:"name"`
	Score    int    `json:"score"`
	RiskType string `json:"risk_type"`
}

type PromptChoice struct {
	Index   int    `json:"index"`
	Text    string `json:"text,omitempty"`
	Message any    `json:"message,omitempty"`
}

type PromptResponse struct {
	ID      string         `json:"id"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []PromptChoice `json:"choices"`
	Usage   *Usage         `json:"usage,omitempty"`
}

type ModelPricingTier struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type ModelInfo struct {
	Model string             `json:"model"`
	Tiers []ModelPricingTier `json:"tiers,omitempty"`
}

type ModelPricesResponse struct {
	Success bool             `json:"success"`
	Data    *ModelPricesData `json:"data,omitempty"`
}

type ModelPricesData struct {
	Total  int         `json:"total"`
	Models []ModelInfo `json:"models"`
}

func NewGenerateRequest(model string) *GenerateRequest {
	return &GenerateRequest{
		Model:      model,
		DashScope:  true,
		Moderation: true,
		Input: []InputItem{
			{Params: map[string]any{}},
		},
		Metadata: map[string]any{},
	}
}

type TaskEvent struct {
	Status   string
	Progress float64
	Task     *TaskResponse
	Err      error
}

type PollOption func(*PollConfig)

type PollConfig struct {
	Interval time.Duration
	Timeout  time.Duration
	OnUpdate func(status string, progress float64)
}

func DefaultPollConfig() PollConfig {
	return PollConfig{
		Interval: 3 * time.Second,
		Timeout:  5 * time.Minute,
	}
}

func WithPollInterval(d time.Duration) PollOption {
	return func(p *PollConfig) { p.Interval = d }
}

func WithPollTimeout(d time.Duration) PollOption {
	return func(p *PollConfig) { p.Timeout = d }
}

func WithPollCallback(fn func(status string, progress float64)) PollOption {
	return func(p *PollConfig) { p.OnUpdate = fn }
}

func ApplyPollOptions(opts ...PollOption) PollConfig {
	cfg := DefaultPollConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
