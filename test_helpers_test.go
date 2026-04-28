package sa_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sa "github.com/seaart/sa-go"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *sa.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := sa.New(
		&sa.ClientConfig{
			APIKey:       "test-key",
			ModelBaseURL: srv.URL,
			LLMBaseURL:   srv.URL,
			Timeout:      5 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return srv, client
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func extractParams(t *testing.T, r *http.Request) map[string]any {
	t.Helper()

	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	return body["input"].([]any)[0].(map[string]any)["params"].(map[string]any)
}

func extractBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	return body
}
