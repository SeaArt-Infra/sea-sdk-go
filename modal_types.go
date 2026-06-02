package sa

import (
	mmtypes "github.com/SeaVerseAI/sea-sdk-go/internal/multimodal/types"
	"github.com/SeaVerseAI/sea-sdk-go/internal/transport"
)

type TaskCreateRequest struct {
	Model      string         `json:"model"`
	Input      []InputItem    `json:"input,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Options    map[string]any `json:"options,omitempty"`
}

type InputItem struct {
	Type    string         `json:"type,omitempty"`
	Role    string         `json:"role,omitempty"`
	Text    string         `json:"text,omitempty"`
	URL     string         `json:"url,omitempty"`
	FileID  string         `json:"file_id,omitempty"`
	MIME    string         `json:"mime,omitempty"`
	Content []ContentPart  `json:"content,omitempty"`
	Extra   map[string]any `json:"extra,omitempty"`
}

type ContentPart struct {
	Type  string         `json:"type"`
	Text  string         `json:"text,omitempty"`
	URL   string         `json:"url,omitempty"`
	ID    string         `json:"file_id,omitempty"`
	MIME  string         `json:"mime,omitempty"`
	Extra map[string]any `json:"extra,omitempty"`
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
	if len(r.Input) > 0 {
		input := make([]map[string]any, 0, len(r.Input))
		for _, item := range r.Input {
			entry := make(map[string]any)
			if item.Type != "" {
				entry["type"] = item.Type
			}
			if item.Role != "" {
				entry["role"] = item.Role
			}
			if item.Text != "" {
				entry["text"] = item.Text
			}
			if item.URL != "" {
				entry["url"] = item.URL
			}
			if item.FileID != "" {
				entry["file_id"] = item.FileID
			}
			if item.MIME != "" {
				entry["mime"] = item.MIME
			}
			if len(item.Content) > 0 {
				content := make([]map[string]any, 0, len(item.Content))
				for _, part := range item.Content {
					partEntry := map[string]any{
						"type": part.Type,
					}
					if part.Text != "" {
						partEntry["text"] = part.Text
					}
					if part.URL != "" {
						partEntry["url"] = part.URL
					}
					if part.ID != "" {
						partEntry["file_id"] = part.ID
					}
					if part.MIME != "" {
						partEntry["mime"] = part.MIME
					}
					if len(part.Extra) > 0 {
						partEntry["extra"] = part.Extra
					}
					content = append(content, partEntry)
				}
				entry["content"] = content
			}
			if len(item.Extra) > 0 {
				entry["extra"] = item.Extra
			}
			input = append(input, entry)
		}
		body["input"] = input
	}
	if len(r.Parameters) > 0 {
		body["parameters"] = r.Parameters
	}
	if len(r.Metadata) > 0 {
		body["metadata"] = r.Metadata
	}
	if len(r.Options) > 0 {
		body["options"] = r.Options
	}
	return body
}

type TaskBuilder struct {
	req TaskCreateRequest
}

func NewTask(model string) *TaskBuilder {
	return &TaskBuilder{
		req: TaskCreateRequest{
			Model:      model,
			Parameters: map[string]any{},
			Metadata:   map[string]any{},
			Options:    map[string]any{},
		},
	}
}

func (b *TaskBuilder) Input(item InputItem) *TaskBuilder {
	b.req.Input = append(b.req.Input, item)
	return b
}

func (b *TaskBuilder) User(parts ...ContentPart) *TaskBuilder {
	b.req.Input = append(b.req.Input, User(parts...))
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

func (b *TaskBuilder) Build() JSONMap {
	return b.req.Raw()
}

func Text(text string) ContentPart {
	return ContentPart{Type: "text", Text: text}
}

func ImageURL(url string) ContentPart {
	return ContentPart{Type: "image_url", URL: url}
}

func VideoURL(url string) ContentPart {
	return ContentPart{Type: "video_url", URL: url}
}

func AudioURL(url string) ContentPart {
	return ContentPart{Type: "audio_url", URL: url}
}

func FileID(id string) ContentPart {
	return ContentPart{Type: "file_id", ID: id}
}

func User(parts ...ContentPart) InputItem {
	return InputItem{Type: "message", Role: "user", Content: parts}
}
