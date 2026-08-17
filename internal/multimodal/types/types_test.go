package types

import (
	"encoding/json"
	"testing"
)

func TestUsageUnmarshalJSONAcceptsStringNumberAndEmptyCost(t *testing.T) {
	tests := []struct {
		name string
		body string
		cost string
	}{
		{
			name: "numeric string",
			body: `{"cost":"0.01429","discount":0.5}`,
			cost: "0.01429",
		},
		{
			name: "number",
			body: `{"cost":0.01429,"discount":0.5}`,
			cost: "0.01429",
		},
		{
			name: "empty string",
			body: `{"cost":"","discount":0}`,
			cost: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var usage Usage
			if err := json.Unmarshal([]byte(tt.body), &usage); err != nil {
				t.Fatalf("unmarshal usage: %v", err)
			}
			if usage.Cost.String() != tt.cost {
				t.Fatalf("cost = %q, want %q", usage.Cost.String(), tt.cost)
			}
		})
	}
}

func TestUsageUnmarshalJSONRejectsNonNumericCost(t *testing.T) {
	var usage Usage
	if err := json.Unmarshal([]byte(`{"cost":"not-a-number"}`), &usage); err == nil {
		t.Fatal("expected a non-numeric cost to fail decoding")
	}
}
