package sa

import (
	"context"
	"net/http"
	"strings"

	mmservice "github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/service"
	mmtypes "github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/types"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/transport"
)

func (m *ModalService) Create(ctx context.Context, body JSONMap, opts ...RequestOption) (*Task, error) {
	requestBody, headers, err := moveModelToHeader(body, buildRequestOptions(opts).headers)
	if err != nil {
		return nil, err
	}
	resp, err := mmservice.CreateTask(m.client, ctx, requestBody, headers)
	if err != nil {
		return nil, err
	}
	return &Task{
		ID:     resp.ID,
		Status: resp.Status,
		Model:  resp.Model,
		Error:  resp.Error,
		client: m.client,
	}, nil
}

// Precharge queries the billing preview via POST /v1/generation/precharge.
// The request body shape is the same as Create.
func (m *ModalService) Precharge(ctx context.Context, body JSONMap, opts ...RequestOption) (*PrechargeResponse, error) {
	requestBody, headers, err := moveModelToHeader(body, buildRequestOptions(opts).headers)
	if err != nil {
		return nil, err
	}
	return mmservice.Precharge(m.client, ctx, requestBody, headers)
}

// CreateComfyUITask creates a ComfyUI quick-app task with the gateway-required request shape.
func (m *ModalService) CreateComfyUITask(ctx context.Context, templateID string, inputs []ComfyUIInput, highMemory *bool, opts ...RequestOption) (*Task, error) {
	if strings.TrimSpace(templateID) == "" {
		return nil, &Error{Kind: ErrGeneral, Message: "template_id is required"}
	}
	if len(inputs) == 0 {
		return nil, &Error{Kind: ErrGeneral, Message: "inputs is required"}
	}
	for _, input := range inputs {
		if strings.TrimSpace(input.Field) == "" {
			return nil, &Error{Kind: ErrGeneral, Message: "each ComfyUI input requires field"}
		}
	}
	requestBody := JSONMap{"model": "comfyui", "input": []map[string]any{{
		"params": map[string]any{
			"template_id": templateID,
			"inputs":      inputs,
		},
	}}}
	if highMemory != nil {
		requestBody["input"].([]map[string]any)[0]["params"].(map[string]any)["high_memory"] = *highMemory
	}
	requestBody, headers, err := moveModelToHeader(requestBody, buildRequestOptions(opts).headers)
	if err != nil {
		return nil, err
	}
	resp, err := mmservice.CreateTask(m.client, ctx, requestBody, headers)
	if err != nil {
		return nil, err
	}
	return &Task{ID: resp.ID, Status: resp.Status, Model: resp.Model, Error: resp.Error, client: m.client}, nil
}

// ListComfyUITemplates returns parameter specifications for the supplied template IDs.
func (m *ModalService) ListComfyUITemplates(ctx context.Context, templateIDs []string, opts ...RequestOption) (*ComfyUITemplateSpecsResponse, error) {
	return mmservice.ListComfyUITemplates(m.client, ctx, templateIDs, buildRequestOptions(opts).headers)
}

// ListModels searches multimodal model skills via GET /v1/models/skill/search.
//
// Supported params:
//   - Query maps to q
//   - Input maps to input
//   - Output maps to output
//   - Type maps to type
//   - Provider maps to provider
//   - Limit maps to limit
func (m *ModalService) ListModels(ctx context.Context, params ModalModelSearchParams, opts ...RequestOption) (*ModalModelSearchResponse, error) {
	return mmservice.SearchModels(m.client, ctx, mmtypes.ModelSearchParams(params), buildRequestOptions(opts).headers)
}

// SearchModels searches multimodal model skills via GET /v1/models/skill/search.
//
// Supported params:
//   - Query maps to q
//   - Input maps to input
//   - Output maps to output
//   - Type maps to type
//   - Provider maps to provider
//   - Limit maps to limit
func (m *ModalService) SearchModels(ctx context.Context, params ModalModelSearchParams, opts ...RequestOption) (*ModalModelSearchResponse, error) {
	return m.ListModels(ctx, params, opts...)
}

func (m *ModalService) GetModelSkill(ctx context.Context, model string, opts ...RequestOption) (string, error) {
	return mmservice.GetModelSkill(m.client, ctx, model, buildRequestOptions(opts).headers)
}

// ScanImage scans an image, GIF, or video through ModelBaseURL + /v1/image/scan.
func (m *ModalService) ScanImage(ctx context.Context, req ImageScanRequest, opts ...RequestOption) (*ImageScanResponse, error) {
	return mmservice.ScanImage(m.client, ctx, mmtypes.ImageScanRequest(req), buildRequestOptions(opts).headers)
}

// ScanText scans prompt text through ModelBaseURL + /v1/text/scan.
func (m *ModalService) ScanText(ctx context.Context, req TextScanRequest, opts ...RequestOption) (*TextScanResponse, error) {
	return mmservice.ScanText(m.client, ctx, mmtypes.TextScanRequest(req), buildRequestOptions(opts).headers)
}

// ScanTextContent scans short text through ModelBaseURL + /v1/text/content/scan.
func (m *ModalService) ScanTextContent(ctx context.Context, req TextContentScanRequest, opts ...RequestOption) (*TextContentScanResponse, error) {
	return mmservice.ScanTextContent(m.client, ctx, mmtypes.TextContentScanRequest(req), buildRequestOptions(opts).headers)
}

// ScanVisualStructuredTextFusion scans a digital-human cover image and structured text together.
func (m *ModalService) ScanVisualStructuredTextFusion(ctx context.Context, req VisualStructuredTextFusionScanRequest, opts ...RequestOption) (*VisualStructuredTextFusionScanResponse, error) {
	return mmservice.ScanVisualStructuredTextFusion(m.client, ctx, mmtypes.VisualStructuredTextFusionScanRequest(req), buildRequestOptions(opts).headers)
}

// ScanAudio scans audio through ModelBaseURL + /v1/audio/scan.
func (m *ModalService) ScanAudio(ctx context.Context, req AudioScanRequest, opts ...RequestOption) (*AudioScanResponse, error) {
	return mmservice.ScanAudio(m.client, ctx, mmtypes.AudioScanRequest(req), buildRequestOptions(opts).headers)
}

// ScanFace scans an image or video through ModelBaseURL + /v1/face/scan.
func (m *ModalService) ScanFace(ctx context.Context, req FaceScanRequest, opts ...RequestOption) (*FaceScanResponse, error) {
	return mmservice.ScanFace(m.client, ctx, mmtypes.FaceScanRequest(req), buildRequestOptions(opts).headers)
}

func getTask(client *transport.Client, ctx context.Context, taskID string, headers http.Header) (*Task, error) {
	resp, err := mmservice.GetTask(client, ctx, taskID, headers)
	if err != nil {
		return nil, err
	}
	return newTaskFromResponse(client, resp), nil
}

func newTaskFromResponse(client *transport.Client, resp *mmtypes.TaskResponse) *Task {
	if resp == nil {
		return nil
	}
	return &Task{
		ID:       resp.ID,
		Status:   resp.Status,
		Model:    resp.Model,
		Progress: resp.Progress,
		Output:   resp.Output,
		Usage:    resp.Usage,
		Error:    resp.Error,
		client:   client,
	}
}
