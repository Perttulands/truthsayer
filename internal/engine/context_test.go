package engine

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildSourceContext_WindowIsPlusMinusTen(t *testing.T) {
	lines := make([]string, 0, 30)
	for i := 1; i <= 30; i++ {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}

	got := buildSourceContext(lines, 15)
	if got == "" {
		t.Fatal("expected context, got empty string")
	}

	if !strings.Contains(got, ">> 15 | line 15") {
		t.Fatalf("expected focused line marker in context, got:\n%s", got)
	}
	if !strings.Contains(got, "| line 5") {
		t.Fatalf("expected start of context window at line 5, got:\n%s", got)
	}
	if !strings.Contains(got, "| line 25") {
		t.Fatalf("expected end of context window at line 25, got:\n%s", got)
	}
	if strings.Contains(got, "line 4") {
		t.Fatalf("expected line 4 to be outside context window, got:\n%s", got)
	}
	if strings.Contains(got, "line 26") {
		t.Fatalf("expected line 26 to be outside context window, got:\n%s", got)
	}
}

func TestBuildSourceContext_ClampsAtFileBounds(t *testing.T) {
	lines := []string{"a", "b", "c"}
	got := buildSourceContext(lines, 1)
	if !strings.Contains(got, ">> 1 | a") {
		t.Fatalf("expected line 1 marker, got:\n%s", got)
	}
	if !strings.Contains(got, "| c") {
		t.Fatalf("expected line 3 in context, got:\n%s", got)
	}
	if strings.Contains(got, "  4 |") {
		t.Fatalf("expected no lines outside file bounds, got:\n%s", got)
	}
}
