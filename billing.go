package sa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/SeaArt-Infra/sea-sdk-go/internal/transport"
)

const billingPath = "/api/v1/cost/billing"

// BillingService provides team-scoped cost statement queries.
type BillingService struct {
	client *transport.Client
}

// Query returns the billing statement for the team identified by the gateway.
func (b *BillingService) Query(ctx context.Context, params BillingQuery, opts ...RequestOption) (*BillingResponse, error) {
	values := url.Values{}
	if params.Start != "" {
		values.Set("start", params.Start)
	}
	if params.End != "" {
		values.Set("end", params.End)
	}
	if params.Environment != "" {
		values.Set("environment", params.Environment)
	}
	if params.Provider != "" {
		values.Set("provider", params.Provider)
	}
	if params.CredentialName != "" {
		values.Set("credential_name", params.CredentialName)
	}
	if params.ModelGroup != "" {
		values.Set("model_group", params.ModelGroup)
	}
	if params.Page > 0 {
		values.Set("page", strconv.Itoa(params.Page))
	}
	if params.PageSize > 0 {
		values.Set("page_size", strconv.Itoa(params.PageSize))
	}
	path := billingPath
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	requestOptions := buildRequestOptions(opts)
	status, payload, err := b.client.Request(ctx, http.MethodGet, path, nil, requestOptions.headers)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, billingHTTPError(status, payload)
	}
	var envelope struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    BillingResponse `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, &Error{Kind: ErrGeneral, Message: "failed to decode billing response: " + err.Error()}
	}
	if envelope.Code != 0 {
		return nil, &Error{Kind: ErrGeneral, Status: status, Code: envelope.Code, Message: envelope.Message}
	}
	return &envelope.Data, nil
}

// Get is an alias for Query.
func (b *BillingService) Get(ctx context.Context, params BillingQuery, opts ...RequestOption) (*BillingResponse, error) {
	return b.Query(ctx, params, opts...)
}

func billingHTTPError(status int, payload []byte) *Error {
	var envelope struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(payload, &envelope)
	message := envelope.Message
	if message == "" {
		message = http.StatusText(status)
	}
	err := newHTTPError(status, message)
	err.Code = envelope.Code
	return err
}
