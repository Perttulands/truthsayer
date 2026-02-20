package cost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultMetricsPath is the default file for cost metrics logs.
const DefaultMetricsPath = ".truthsayer-cost.jsonl"

const (
	haikuInputUSDPerMillionTokens  = 0.80
	haikuOutputUSDPerMillionTokens = 4.00
)

// EstimateUSD estimates LLM spend from token counts.
func EstimateUSD(inputTokens, outputTokens int) float64 {
	if inputTokens < 0 {
		inputTokens = 0
	}
	if outputTokens < 0 {
		outputTokens = 0
	}
	return (float64(inputTokens)/1_000_000.0)*haikuInputUSDPerMillionTokens +
		(float64(outputTokens)/1_000_000.0)*haikuOutputUSDPerMillionTokens
}

// Metrics captures one run-level cost record.
type Metrics struct {
	RecordedAt      string  `json:"recorded_at"`
	LLMCalls        int     `json:"llm_calls"`
	InputTokens     int     `json:"input_tokens"`
	OutputTokens    int     `json:"output_tokens"`
	TotalCostUSD    float64 `json:"total_cost_usd"`
	BudgetUSD       float64 `json:"budget_usd,omitempty"`
	BudgetExhausted bool    `json:"budget_exhausted,omitempty"`
}

// Append writes one JSONL metrics record.
func Append(path string, m Metrics) error {
	if m.RecordedAt == "" {
		m.RecordedAt = time.Now().UTC().Format(time.RFC3339)
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("cost: create dir %s: %w", dir, err)
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("cost: open metrics %s: %w", path, err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(m); err != nil {
		return fmt.Errorf("cost: append metrics %s: %w", path, err)
	}
	return nil
}
