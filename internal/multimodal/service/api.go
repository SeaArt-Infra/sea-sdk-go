package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	mmtypes "github.com/SeaVerseAI/sea-sdk-go/internal/multimodal/types"
	"github.com/SeaVerseAI/sea-sdk-go/internal/shared"
	"github.com/SeaVerseAI/sea-sdk-go/internal/transport"
)

const (
	PathGeneration       = "/v1/generation"
	PathTask             = "/v1/generation/task/"
	PathModelSkillSearch = "/v1/models/skill/search"
	PathModelSkill       = "/v1/models/skill/"
	// PathImageScan is the content-safety scan endpoint used for images, GIFs, and videos.
	PathImageScan = "/v1/image/scan"
	// PathFaceScan is the face-detection scan endpoint used for images and videos.
	PathFaceScan = "/v1/face/scan"
)

func CreateTask(client *transport.Client, ctx context.Context, body any, headers http.Header) (*mmtypes.GenerationResponse, error) {
	status, payload, err := client.Request(ctx, http.MethodPost, PathGeneration, body, headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, httpError(status, payload)
	}

	var resp mmtypes.GenerationResponse
	if err := decode(payload, &resp); err != nil {
		return nil, err
	}
	if resp.ID == "" {
		return nil, &shared.Error{Kind: shared.ErrGeneral, Message: "API returned no task ID"}
	}
	return &resp, nil
}

func GetTask(client *transport.Client, ctx context.Context, taskID string, headers http.Header) (*mmtypes.TaskResponse, error) {
	status, payload, err := client.Request(ctx, http.MethodGet, PathTask+taskID, nil, headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, httpError(status, payload)
	}

	var resp mmtypes.TaskResponse
	if err := decode(payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func SearchModels(client *transport.Client, ctx context.Context, params mmtypes.ModelSearchParams, headers http.Header) (*mmtypes.ModelSearchResponse, error) {
	status, payload, err := client.Request(ctx, http.MethodGet, PathModelSkillSearch+modelSearchQuery(params), nil, withDefaultHeader(headers, "Accept", "application/json"))
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, httpError(status, payload)
	}

	var resp mmtypes.ModelSearchResponse
	if err := decode(payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetModelSkill(client *transport.Client, ctx context.Context, model string, headers http.Header) (string, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return "", &shared.Error{Kind: shared.ErrGeneral, Message: "model is required"}
	}

	status, payload, err := client.Request(ctx, http.MethodGet, PathModelSkill+url.PathEscape(model), nil, withDefaultHeader(headers, "Accept", "application/json"))
	if err != nil {
		return "", err
	}
	if status >= 400 {
		return "", httpError(status, payload)
	}

	return string(payload), nil
}

// ScanImage sends an image, GIF, or video safety scan request to PathImageScan.
func ScanImage(client *transport.Client, ctx context.Context, req mmtypes.ImageScanRequest, headers http.Header) (*mmtypes.ImageScanResponse, error) {
	req.URI = strings.TrimSpace(req.URI)
	if req.URI == "" {
		return nil, &shared.Error{Kind: shared.ErrGeneral, Message: "uri is required"}
	}

	status, payload, err := client.Request(ctx, http.MethodPost, PathImageScan, req, headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, httpError(status, payload)
	}

	var resp mmtypes.ImageScanResponse
	if err := decode(payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ScanFace sends an image or video face-detection scan request to PathFaceScan.
func ScanFace(client *transport.Client, ctx context.Context, req mmtypes.FaceScanRequest, headers http.Header) (*mmtypes.FaceScanResponse, error) {
	req.URI = strings.TrimSpace(req.URI)
	req.ImgBase64 = strings.TrimSpace(req.ImgBase64)
	if req.URI == "" && req.ImgBase64 == "" {
		return nil, &shared.Error{Kind: shared.ErrGeneral, Message: "uri or img_base64 is required"}
	}

	status, payload, err := client.Request(ctx, http.MethodPost, PathFaceScan, req, headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, httpError(status, payload)
	}

	var resp mmtypes.FaceScanResponse
	if err := decode(payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func modelSearchQuery(params mmtypes.ModelSearchParams) string {
	values := url.Values{}
	values.Set("q", params.Query)
	if params.Input != "" {
		values.Set("input", params.Input)
	}
	if params.Output != "" {
		values.Set("output", params.Output)
	}
	if params.Type != "" {
		values.Set("type", params.Type)
	}
	if params.Provider != "" {
		values.Set("provider", params.Provider)
	}
	if params.Limit > 0 {
		values.Set("limit", strconv.Itoa(params.Limit))
	}
	return "?" + values.Encode()
}

func withDefaultHeader(headers http.Header, key, value string) http.Header {
	if headers.Get(key) != "" {
		return headers
	}

	cloned := make(http.Header, len(headers)+1)
	for name, values := range headers {
		for _, v := range values {
			cloned.Add(name, v)
		}
	}
	cloned.Set(key, value)
	return cloned
}

func httpError(status int, payload []byte) error {
	var apiErr struct {
		Error *mmtypes.APIError `json:"error"`
	}
	_ = json.Unmarshal(payload, &apiErr)

	message := "HTTP error"
	if apiErr.Error != nil && apiErr.Error.ErrorMessage != "" {
		message = apiErr.Error.ErrorMessage
	} else {
		message = http.StatusText(status)
		if message == "" {
			message = "HTTP error"
		}
	}
	return shared.NewHTTPError(status, message)
}

func decode(payload []byte, out any) error {
	if err := json.Unmarshal(payload, out); err != nil {
		return &shared.Error{
			Kind:    shared.ErrGeneral,
			Message: "failed to decode response: " + err.Error(),
		}
	}
	return nil
}
