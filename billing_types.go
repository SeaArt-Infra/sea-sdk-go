package sa

// BillingQuery controls GET /monitor/api/v1/cost/billing.
// The team is derived from the authenticated gateway identity and must not be
// supplied by callers.
type BillingQuery struct {
	Start          string
	End            string
	Environment    string
	Provider       string
	CredentialName string
	ModelGroup     string
	Page           int
	PageSize       int
}

type BillingSummary struct {
	TotalRequests     int64  `json:"total_requests"`
	TotalCost         string `json:"total_cost"`
	DiscountTotalCost string `json:"discount_total_cost"`
	PromptTokens      int64  `json:"prompt_tokens"`
	CompletionTokens  int64  `json:"completion_tokens"`
	TotalTokens       int64  `json:"total_tokens"`
	Currency          string `json:"currency"`
}

type BillingItem struct {
	TeamAlias         string  `json:"team_alias"`
	Provider          string  `json:"provider"`
	ModelGroup        string  `json:"model_group"`
	TotalRequests     int64   `json:"total_requests"`
	TotalCost         string  `json:"total_cost"`
	DiscountTotalCost string  `json:"discount_total_cost"`
	DiscountRate      float64 `json:"discount_rate"`
	PromptTokens      int64   `json:"prompt_tokens"`
	CompletionTokens  int64   `json:"completion_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
}

type BillingPage struct {
	Items      []BillingItem `json:"items"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

type BillingResponse struct {
	Team         string         `json:"team"`
	Environments []string       `json:"environments"`
	Summary      BillingSummary `json:"summary"`
	Items        BillingPage    `json:"items"`
}
