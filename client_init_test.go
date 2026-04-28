package sa

import "testing"

func TestNew_DefaultBaseURLs(t *testing.T) {
	client, err := New(&ClientConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.baseURL != defaultBaseURL {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
	if client.modelBaseURL != defaultModelBaseURL {
		t.Fatalf("unexpected modelBaseURL: %s", client.modelBaseURL)
	}
	if client.llmBaseURL != defaultLLMBaseURL {
		t.Fatalf("unexpected llmBaseURL: %s", client.llmBaseURL)
	}
	if client.passthroughBaseURL != defaultPassthroughURL {
		t.Fatalf("unexpected passthroughBaseURL: %s", client.passthroughBaseURL)
	}
}

func TestNew_DerivesServiceBaseURLsFromBaseURL(t *testing.T) {
	client, err := New(&ClientConfig{
		APIKey:  "test-key",
		BaseURL: "https://gateway.example.com/",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.baseURL != "https://gateway.example.com" {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
	if client.modelBaseURL != "https://gateway.example.com/model" {
		t.Fatalf("unexpected modelBaseURL: %s", client.modelBaseURL)
	}
	if client.llmBaseURL != "https://gateway.example.com/llm" {
		t.Fatalf("unexpected llmBaseURL: %s", client.llmBaseURL)
	}
	if client.passthroughBaseURL != "https://gateway.example.com/model" {
		t.Fatalf("unexpected passthroughBaseURL: %s", client.passthroughBaseURL)
	}
}

func TestNew_ServiceBaseURLOverrides(t *testing.T) {
	client, err := New(&ClientConfig{
		APIKey:             "test-key",
		BaseURL:            "https://gateway.example.com",
		ModelBaseURL:       "https://model.example.com",
		LLMBaseURL:         "https://llm.example.com",
		PassthroughBaseURL: "https://passthrough.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.modelBaseURL != "https://model.example.com" {
		t.Fatalf("unexpected modelBaseURL: %s", client.modelBaseURL)
	}
	if client.llmBaseURL != "https://llm.example.com" {
		t.Fatalf("unexpected llmBaseURL: %s", client.llmBaseURL)
	}
	if client.passthroughBaseURL != "https://passthrough.example.com" {
		t.Fatalf("unexpected passthroughBaseURL: %s", client.passthroughBaseURL)
	}
}

func TestNew_InvalidBaseURL(t *testing.T) {
	_, err := New(&ClientConfig{
		APIKey:  "test-key",
		BaseURL: "://bad",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNew_NilConfigUsesDefaults(t *testing.T) {
	client, err := New(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.baseURL != defaultBaseURL {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
	if client.Modal == nil {
		t.Fatal("expected Modal service to be initialized")
	}
	if client.LLM == nil {
		t.Fatal("expected LLM service to be initialized")
	}
	if client.Passthrough == nil {
		t.Fatal("expected Passthrough service to be initialized")
	}
}
