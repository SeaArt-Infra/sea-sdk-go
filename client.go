package sa

import (
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/SeaArt-Infra/sea-sdk-go/internal/transport"
)

const (
	defaultBaseURL        = "https://gateway.example.com"
	defaultModelBaseURL   = defaultBaseURL + "/model"
	defaultLLMBaseURL     = defaultBaseURL + "/llm"
	defaultPassthroughURL = defaultModelBaseURL
	defaultTimeout        = 5 * time.Minute
	sdkVersion            = "0.1.0"
)

// Client is the SeaArt API client. Create one with New() and reuse it.
type Client struct {
	apiKey             string
	baseURL            string
	modelBaseURL       string
	llmBaseURL         string
	passthroughBaseURL string
	project            string
	httpClient         *http.Client

	Modal       *ModalService
	LLM         *LLMService
	Passthrough *PassthroughService
}

// ClientConfig configures a Client.
// BaseURL acts as the root gateway URL; if ModelBaseURL / LLMBaseURL are not
// explicitly set, they will be derived from it. PassthroughBaseURL defaults to
// ModelBaseURL unless explicitly set. If none of them are set, the SDK falls
// back to its built-in defaults.
type ClientConfig struct {
	APIKey             string
	BaseURL            string
	ModelBaseURL       string
	LLMBaseURL         string
	PassthroughBaseURL string
	Project            string
	HTTPClient         *http.Client
	Timeout            time.Duration
}

type resolvedEndpoints struct {
	root        string
	model       string
	llm         string
	passthrough string
}

// New creates a Client from an explicit ClientConfig.
func New(cfg *ClientConfig) (*Client, error) {
	if cfg == nil {
		cfg = &ClientConfig{}
	}

	endpoints, err := resolveEndpoints(*cfg)
	if err != nil {
		return nil, err
	}

	return newClient(*cfg, endpoints), nil
}

func resolveEndpoints(cfg ClientConfig) (resolvedEndpoints, error) {
	root, err := resolveRootURL(cfg.BaseURL)
	if err != nil {
		return resolvedEndpoints{}, err
	}

	model, err := resolveServiceURL(cfg.ModelBaseURL, cfg.BaseURL != "", root, "model", defaultModelBaseURL)
	if err != nil {
		return resolvedEndpoints{}, err
	}

	llm, err := resolveServiceURL(cfg.LLMBaseURL, cfg.BaseURL != "", root, "llm", defaultLLMBaseURL)
	if err != nil {
		return resolvedEndpoints{}, err
	}

	passthrough, err := resolvePassthroughURL(cfg.PassthroughBaseURL, model)
	if err != nil {
		return resolvedEndpoints{}, err
	}

	return resolvedEndpoints{
		root:        root,
		model:       model,
		llm:         llm,
		passthrough: passthrough,
	}, nil
}

func resolveRootURL(raw string) (string, error) {
	if raw == "" {
		raw = defaultBaseURL
	}

	return normalizeURL(raw)
}

func resolveServiceURL(raw string, deriveFromRoot bool, root, suffix, fallback string) (string, error) {
	switch {
	case raw != "":
		return normalizeURL(raw)
	case deriveFromRoot:
		return joinURL(root, suffix)
	default:
		return fallback, nil
	}
}

func resolvePassthroughURL(raw, model string) (string, error) {
	if raw != "" {
		return normalizeURL(raw)
	}
	return model, nil
}

func newClient(cfg ClientConfig, endpoints resolvedEndpoints) *Client {
	httpClient := buildHTTPClient(cfg)

	client := &Client{
		apiKey:             cfg.APIKey,
		baseURL:            endpoints.root,
		modelBaseURL:       endpoints.model,
		llmBaseURL:         endpoints.llm,
		passthroughBaseURL: endpoints.passthrough,
		project:            cfg.Project,
		httpClient:         httpClient,
	}

	client.Modal = &ModalService{
		client: &transport.Client{
			APIKey:     client.apiKey,
			BaseURL:    client.modelBaseURL,
			Project:    client.project,
			UserAgent:  "sa-go/" + sdkVersion,
			HTTPClient: httpClient,
		},
	}
	client.LLM = &LLMService{
		client: &transport.Client{
			APIKey:     client.apiKey,
			BaseURL:    client.llmBaseURL,
			Project:    client.project,
			UserAgent:  "sa-go/" + sdkVersion,
			HTTPClient: httpClient,
		},
	}
	client.Passthrough = &PassthroughService{
		client: &transport.Client{
			APIKey:     client.apiKey,
			BaseURL:    client.passthroughBaseURL,
			Project:    client.project,
			UserAgent:  "sa-go/" + sdkVersion,
			HTTPClient: httpClient,
		},
	}

	return client
}

func normalizeURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", &Error{Kind: ErrGeneral, Message: "invalid URL: " + err.Error()}
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", &Error{Kind: ErrGeneral, Message: "invalid URL: missing scheme or host"}
	}

	parsed.Path = path.Clean("/" + parsed.Path)
	if parsed.Path == "/" {
		parsed.Path = ""
	}

	return parsed.String(), nil
}

func joinURL(baseURL, suffix string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", &Error{Kind: ErrGeneral, Message: "invalid URL: " + err.Error()}
	}

	joined := *parsed
	joined.Path = path.Join(parsed.Path, suffix)
	return normalizeURL(joined.String())
}

func buildHTTPClient(cfg ClientConfig) *http.Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	if cfg.HTTPClient == nil {
		return &http.Client{Timeout: timeout}
	}

	cloned := *cfg.HTTPClient
	cloned.Timeout = timeout
	return &cloned
}
