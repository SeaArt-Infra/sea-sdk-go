package sa

import (
	mmtypes "github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/types"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/transport"
)

type TaskCreateRequest struct {
	Model         string         `json:"model"`
	Params        map[string]any `json:"-"`
	Parameters    map[string]any `json:"-"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	Options       map[string]any `json:"options,omitempty"`
	Moderation    *bool          `json:"moderation,omitempty"`
	ExtraTopLevel map[string]any `json:"-"`
}

type Task struct {
	ID       string
	Status   string
	Model    string
	Progress float64
	Output   []Output
	Usage    *Usage
	Error    *APIError

	client *transport.Client
}

type Output = mmtypes.OutputItem
type OutputContent = mmtypes.OutputContent
type Usage = mmtypes.Usage
type APIError = mmtypes.APIError

// ImageScanRiskType selects which safety categories POST /v1/image/scan should detect.
type ImageScanRiskType = mmtypes.ImageScanRiskType

// ImageScanRequest is the request body for POST /v1/image/scan.
type ImageScanRequest = mmtypes.ImageScanRequest

// ImageScanResponse is the parsed response returned by POST /v1/image/scan.
type ImageScanResponse = mmtypes.ImageScanResponse

// ImageScanLabel describes one safety label detected by the scan service.
type ImageScanLabel = mmtypes.ImageScanLabel

// ImageScanFrameResult describes one sampled frame in a video scan.
type ImageScanFrameResult = mmtypes.ImageScanFrameResult

// TextScanRequest is the request body for POST /v1/text/scan.
type TextScanRequest = mmtypes.TextScanRequest

// TextScanResponse is the parsed response returned by POST /v1/text/scan.
type TextScanResponse = mmtypes.TextScanResponse

// TextScanData contains text moderation results.
type TextScanData = mmtypes.TextScanData

// TextScanSensitiveWord describes one sensitive-word match.
type TextScanSensitiveWord = mmtypes.TextScanSensitiveWord

// TextScanStatus contains the upstream business status.
type TextScanStatus = mmtypes.TextScanStatus

// TextScanAreaType selects which regional sensitive-word rules are applied.
type TextScanAreaType = mmtypes.TextScanAreaType

// TextScanWay selects the sensitive-word checking strategy.
type TextScanWay = mmtypes.TextScanWay

// FaceScanRequest is the request body for POST /v1/face/scan.
type FaceScanRequest = mmtypes.FaceScanRequest

// FaceScanResponse is the parsed response returned by POST /v1/face/scan.
type FaceScanResponse = mmtypes.FaceScanResponse

type ModalModelSearchParams = mmtypes.ModelSearchParams
type ModalModelSearchResponse = mmtypes.ModelSearchResponse
type ModalModelSearchHit = mmtypes.ModelSearchHit
type PrechargeResponse = mmtypes.PrechargeResponse
type PrechargeData = mmtypes.PrechargeData

const (
	// ImageScanRiskTypePolity detects political or public-safety sensitive content.
	ImageScanRiskTypePolity = mmtypes.ImageScanRiskTypePolity
	// ImageScanRiskTypeErotic detects erotic, pornographic, nudity, or sexually suggestive content.
	ImageScanRiskTypeErotic = mmtypes.ImageScanRiskTypeErotic
	// ImageScanRiskTypeViolent detects violent, bloody, weapon, or gore-related content.
	ImageScanRiskTypeViolent = mmtypes.ImageScanRiskTypeViolent
	// ImageScanRiskTypeChild detects child-safety risks, especially sexualized or unsafe child-related content.
	ImageScanRiskTypeChild = mmtypes.ImageScanRiskTypeChild

	// TextScanAreaTypeAll checks both domestic and foreign regional rule sets.
	TextScanAreaTypeAll = mmtypes.TextScanAreaTypeAll
	// TextScanAreaTypeDomestic checks the domestic regional rule set.
	TextScanAreaTypeDomestic = mmtypes.TextScanAreaTypeDomestic
	// TextScanAreaTypeForeign checks the foreign regional rule set.
	TextScanAreaTypeForeign = mmtypes.TextScanAreaTypeForeign

	// TextScanWayDictionary uses dictionary matching. This is the upstream default.
	TextScanWayDictionary = mmtypes.TextScanWayDictionary
	// TextScanWayModel uses the big-data model checker.
	TextScanWayModel = mmtypes.TextScanWayModel
	// TextScanWayMixed uses both dictionary and model checks.
	TextScanWayMixed = mmtypes.TextScanWayMixed
	// TextScanWayCharacter uses the digital-human checker.
	TextScanWayCharacter = mmtypes.TextScanWayCharacter
)

func (r TaskCreateRequest) Raw() JSONMap {
	body := JSONMap{
		"model": r.Model,
	}
	if r.Moderation != nil {
		body["moderation"] = *r.Moderation
	}
	params := cloneMap(r.Params)
	if len(r.Parameters) > 0 {
		if existing, ok := params["parameters"].(map[string]any); ok {
			merged := cloneMap(existing)
			for key, value := range r.Parameters {
				merged[key] = value
			}
			params["parameters"] = merged
		} else {
			params["parameters"] = cloneMap(r.Parameters)
		}
	}
	if len(params) > 0 {
		body["input"] = []map[string]any{{"params": params}}
	}
	if len(r.Metadata) > 0 {
		body["metadata"] = r.Metadata
	}
	if len(r.Options) > 0 {
		body["options"] = r.Options
	}
	for key, value := range r.ExtraTopLevel {
		body[key] = value
	}
	return body
}

func cloneMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

type TaskBuilder struct {
	req TaskCreateRequest
}

func NewTask(model string) *TaskBuilder {
	return &TaskBuilder{
		req: TaskCreateRequest{
			Model:         model,
			Params:        map[string]any{},
			Parameters:    map[string]any{},
			Metadata:      map[string]any{},
			Options:       map[string]any{},
			ExtraTopLevel: map[string]any{},
		},
	}
}

func (b *TaskBuilder) Params(value map[string]any) *TaskBuilder {
	b.req.Params = cloneMap(value)
	return b
}

func (b *TaskBuilder) Param(key string, value any) *TaskBuilder {
	if b.req.Parameters == nil {
		b.req.Parameters = map[string]any{}
	}
	b.req.Parameters[key] = value
	return b
}

func (b *TaskBuilder) Metadata(key string, value any) *TaskBuilder {
	if b.req.Metadata == nil {
		b.req.Metadata = map[string]any{}
	}
	b.req.Metadata[key] = value
	return b
}

func (b *TaskBuilder) Option(key string, value any) *TaskBuilder {
	if b.req.Options == nil {
		b.req.Options = map[string]any{}
	}
	b.req.Options[key] = value
	return b
}

func (b *TaskBuilder) Moderation(value bool) *TaskBuilder {
	b.req.Moderation = &value
	return b
}

func (b *TaskBuilder) Field(key string, value any) *TaskBuilder {
	if b.req.ExtraTopLevel == nil {
		b.req.ExtraTopLevel = map[string]any{}
	}
	b.req.ExtraTopLevel[key] = value
	return b
}

func (b *TaskBuilder) Build() JSONMap {
	return b.req.Raw()
}
