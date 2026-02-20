package cost

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEstimateUSD(t *testing.T) {
	got := EstimateUSD(1000, 500)
	if got <= 0 {
		t.Fatalf("expected positive cost, got %f", got)
	}
}

func TestAppend(t *testing.T) {
	path := filepath.Join(t.TempDir(), DefaultMetricsPath)
	err := Append(path, Metrics{
		LLMCalls:     1,
		InputTokens:  100,
		OutputTokens: 20,
		TotalCostUSD: EstimateUSD(100, 20),
	})
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read metrics failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected metrics content")
	}
}
