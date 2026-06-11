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

// PrechargeResponse is returned by POST /v1/generation/precharge.
type PrechargeResponse struct {
	Data   *PrechargeData `json:"data,omitempty"`
	Status string         `json:"status"`
}

// PrechargeData contains the billing preview returned by precharge.
type PrechargeData struct {
	BillingModel  string       `json:"billing_model,omitempty"`
	Cost          *json.Number `json:"cost,omitempty"`
	Currency      string       `json:"currency,omitempty"`
	Discount      float64      `json:"discount,omitempty"`
	Hash          string       `json:"hash,omitempty"`
	Model         string       `json:"model,omitempty"`
	OriginalModel string       `json:"original_model,omitempty"`
	SampleCount   int          `json:"sample_count,omitempty"`
	UpdatedAt     int64        `json:"updated_at,omitempty"`
	Reason        string       `json:"reason,omitempty"`
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

// ImageScanRiskType selects which safety categories the scan service should detect.
type ImageScanRiskType string

const (
	// ImageScanRiskTypePolity detects political or public-safety sensitive content.
	ImageScanRiskTypePolity ImageScanRiskType = "POLITY"
	// ImageScanRiskTypeErotic detects erotic, pornographic, nudity, or sexually suggestive content.
	ImageScanRiskTypeErotic ImageScanRiskType = "EROTIC"
	// ImageScanRiskTypeViolent detects violent, bloody, weapon, or gore-related content.
	ImageScanRiskTypeViolent ImageScanRiskType = "VIOLENT"
	// ImageScanRiskTypeChild detects child-safety risks, especially sexualized or unsafe child-related content.
	ImageScanRiskTypeChild ImageScanRiskType = "CHILD"
)

// ImageScanRequest is the request body for POST /v1/image/scan.
type ImageScanRequest struct {
	// URI is the image, GIF, or video URL to scan.
	URI string `json:"uri"`
	// RiskTypes limits detection to the requested safety categories.
	RiskTypes []ImageScanRiskType `json:"risk_types"`
	// DetectedAge enables age-group detection when set to 1; set to 0 to disable it.
	DetectedAge int `json:"detected_age"`
	// IsVideo marks the URI as video content when set to 1; images and GIFs use 0.
	IsVideo int `json:"is_video"`
	// Duration is the video duration in seconds and is used for video billing when known.
	Duration float64 `json:"duration,omitempty"`
}

// ImageScanResponse is the parsed response returned by POST /v1/image/scan.
type ImageScanResponse struct {
	// OK reports whether the scan service completed the business request successfully.
	OK bool `json:"ok"`
	// NSFWLevel is the highest risk level, usually 0-6. Higher values indicate higher risk.
	NSFWLevel int `json:"nsfw_level,omitempty"`
	// LabelItems contains detailed labels detected on image content or the highest-risk frame.
	LabelItems []ImageScanLabel `json:"label_items,omitempty"`
	// RiskTypes lists the risk categories actually detected in the media.
	RiskTypes []ImageScanRiskType `json:"risk_types,omitempty"`
	// AgeGroup contains provider-specific age-group output, typically age bucket and confidence.
	AgeGroup []any `json:"age_group,omitempty"`
	// Error contains the scan service business error when OK is false.
	Error string `json:"error,omitempty"`
	// VideoDuration is the duration detected or used by the upstream video scanner, in seconds.
	VideoDuration float64 `json:"video_duration,omitempty"`
	// MaxRiskFrame is the frame index with the highest risk in a video scan.
	MaxRiskFrame int `json:"max_risk_frame,omitempty"`
	// FrameCount is the total number of frames sampled or considered by the scanner.
	FrameCount int `json:"frame_count,omitempty"`
	// FramesChecked is the actual number of frames scanned before completion or early exit.
	FramesChecked int `json:"frames_checked,omitempty"`
	// EarlyExit indicates the scanner stopped early after finding a high-risk frame.
	EarlyExit bool `json:"early_exit,omitempty"`
	// FrameResults contains per-frame results for video scans.
	FrameResults []ImageScanFrameResult `json:"frame_results,omitempty"`
	// Usage contains gateway billing metadata injected by inference-gateway.
	Usage *Usage `json:"usage,omitempty"`
}

// ImageScanLabel describes one safety label detected by the scan service.
type ImageScanLabel struct {
	// Name is the provider label name, for example a scene/category/tag path.
	Name string `json:"name"`
	// Score is the label risk score or level, usually aligned to the 0-6 risk scale.
	Score int `json:"score"`
	// RiskType is the safety category this label belongs to.
	RiskType ImageScanRiskType `json:"risk_type"`
}

// ImageScanFrameResult describes one sampled frame in a video scan.
type ImageScanFrameResult struct {
	// FrameIndex is the sampled frame index in the video.
	FrameIndex int `json:"frame_index"`
	// NSFWLevel is the highest risk level detected on this frame.
	NSFWLevel int `json:"nsfw_level"`
	// LabelItems contains detailed labels detected on this frame.
	LabelItems []ImageScanLabel `json:"label_items,omitempty"`
	// RiskTypes lists the risk categories detected on this frame.
	RiskTypes []ImageScanRiskType `json:"risk_types,omitempty"`
}

// TextScanAreaType selects which regional sensitive-word rules are applied.
type TextScanAreaType int

const (
	// TextScanAreaTypeAll checks both domestic and foreign regional rule sets.
	TextScanAreaTypeAll TextScanAreaType = 0
	// TextScanAreaTypeDomestic checks the domestic regional rule set.
	TextScanAreaTypeDomestic TextScanAreaType = 1
	// TextScanAreaTypeForeign checks the foreign regional rule set.
	TextScanAreaTypeForeign TextScanAreaType = 2
)

// TextScanWay selects the sensitive-word checking strategy.
type TextScanWay int

const (
	// TextScanWayDictionary uses dictionary matching. This is the upstream default.
	TextScanWayDictionary TextScanWay = 0
	// TextScanWayModel uses the big-data model checker.
	TextScanWayModel TextScanWay = 1
	// TextScanWayMixed uses both dictionary and model checks.
	TextScanWayMixed TextScanWay = 2
	// TextScanWayCharacter uses the digital-human checker.
	TextScanWayCharacter TextScanWay = 3
)

// TextScanRequest is the request body for POST /v1/text/scan.
type TextScanRequest struct {
	// Text is the prompt or text content to scan for sensitive words.
	Text string `json:"text"`
	// Scene selects the upstream moderation scenario, for example 1 for search
	// sensitive words and 2 for prompt sensitive words.
	Scene int `json:"scene"`
	// AreaTypes limits detection to regional rule sets: 0 checks both domestic
	// and foreign rules, 1 checks domestic rules, and 2 checks foreign rules.
	AreaTypes []TextScanAreaType `json:"area_types,omitempty"`
	// Way selects the checking strategy. When omitted or set to 0, the upstream
	// service uses dictionary matching.
	Way TextScanWay `json:"way"`
	// Scenes is currently unused by the upstream service. When provided, it
	// overrides Scene.
	Scenes []string `json:"scenes,omitempty"`
}

// TextScanResponse is the parsed response returned by POST /v1/text/scan.
//
// Extra keeps any upstream response fields that are not modeled by the SDK yet.
type TextScanResponse struct {
	// Data contains the sensitive words detected by the upstream service.
	Data *TextScanData `json:"data,omitempty"`
	// Status contains the upstream business status. Code 10000 means success.
	Status *TextScanStatus `json:"status,omitempty"`
	// Usage contains gateway billing metadata injected by inference-gateway.
	Usage *Usage `json:"usage,omitempty"`
	// Extra contains upstream response fields that are not modeled by the SDK yet.
	Extra map[string]any `json:"-"`
}

// TextScanData contains text moderation results.
type TextScanData struct {
	// SensitiveWords contains every sensitive word matched by the service.
	SensitiveWords []TextScanSensitiveWord `json:"sensitive_words"`
	// Combination contains upstream combination-rule details when a match is produced.
	Combination any `json:"combination"`
	// IsSensitive reports whether the scanned text matched sensitive content.
	IsSensitive bool `json:"is_sensitive"`
}

// TextScanSensitiveWord describes one sensitive-word match.
type TextScanSensitiveWord struct {
	// Word is the matched sensitive word.
	Word string `json:"word"`
	// StartIndex is the rune-array start index of the matched word.
	StartIndex int `json:"start_index"`
	// EndIndex is the rune-array end index of the matched word.
	EndIndex int `json:"end_index"`
	// RiskTypeCode is the upstream risk category, for example political,
	// violence, or porn.
	RiskTypeCode string `json:"risk_type_code,omitempty"`
}

// TextScanStatus contains the upstream business status.
type TextScanStatus struct {
	// Code is the upstream status code. 10000 means success.
	Code int `json:"code,omitempty"`
	// Msg is the upstream status message.
	Msg string `json:"msg,omitempty"`
	// RequestID is the upstream request trace ID.
	RequestID string `json:"request_id,omitempty"`
}

func (r *TextScanResponse) UnmarshalJSON(data []byte) error {
	type alias TextScanResponse
	var typed alias
	if err := json.Unmarshal(data, &typed); err != nil {
		return err
	}

	var extra map[string]any
	if err := json.Unmarshal(data, &extra); err != nil {
		return err
	}
	delete(extra, "usage")
	delete(extra, "data")
	delete(extra, "status")

	*r = TextScanResponse(typed)
	r.Extra = extra
	return nil
}

// FaceScanRequest is the request body for POST /v1/face/scan.
type FaceScanRequest struct {
	// URI is the image or video URL to scan.
	URI string `json:"uri,omitempty"`
	// ImgBase64 is a base64-encoded image payload. URI or ImgBase64 is required.
	ImgBase64 string `json:"img_base64,omitempty"`
	// IsVideo marks the URI as video content when set to 1; images use 0.
	IsVideo int `json:"is_video"`
	// Canary is forwarded to the upstream face scan service for canary routing.
	Canary string `json:"canary,omitempty"`
	// Scene is forwarded to the upstream face scan service for scenario-specific detection.
	Scene string `json:"scene,omitempty"`
	// Duration is the video duration in seconds and is used for video billing when known.
	Duration float64 `json:"duration,omitempty"`
}

// FaceScanResponse is the parsed response returned by POST /v1/face/scan.
//
// The upstream face scan service owns most response fields, so Extra keeps
// provider-specific fields available while Usage is decoded for gateway billing.
type FaceScanResponse struct {
	// OK reports whether the scan service completed the business request successfully.
	OK bool `json:"ok,omitempty"`
	// Error contains the scan service business error when OK is false.
	Error string `json:"error,omitempty"`
	// Usage contains gateway billing metadata injected by inference-gateway.
	Usage *Usage `json:"usage,omitempty"`
	// Extra contains upstream response fields that are not modeled by the SDK yet.
	Extra map[string]any `json:"-"`
}

func (r *FaceScanResponse) UnmarshalJSON(data []byte) error {
	type alias FaceScanResponse
	var typed alias
	if err := json.Unmarshal(data, &typed); err != nil {
		return err
	}

	var extra map[string]any
	if err := json.Unmarshal(data, &extra); err != nil {
		return err
	}
	delete(extra, "ok")
	delete(extra, "error")
	delete(extra, "usage")

	*r = FaceScanResponse(typed)
	r.Extra = extra
	return nil
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

// ModelSearchParams configures GET /v1/models/skill/search.
type ModelSearchParams struct {
	Query    string
	Input    string
	Output   string
	Type     string
	Provider string
	Limit    int
}

// ModelSearchResponse is the Meilisearch-compatible response returned by
// GET /v1/models/skill/search.
type ModelSearchResponse struct {
	Hits               []ModelSearchHit `json:"hits"`
	Query              string           `json:"query,omitempty"`
	ProcessingTimeMS   int              `json:"processingTimeMs,omitempty"`
	Limit              int              `json:"limit,omitempty"`
	Offset             int              `json:"offset,omitempty"`
	EstimatedTotalHits int              `json:"estimatedTotalHits,omitempty"`
	TotalHits          int              `json:"totalHits,omitempty"`
	TotalPages         int              `json:"totalPages,omitempty"`
	Page               int              `json:"page,omitempty"`
	HitsPerPage        int              `json:"hitsPerPage,omitempty"`
}

// ModelSearchHit keeps model metadata flexible because search documents may
// add provider-specific fields over time.
type ModelSearchHit map[string]any

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
