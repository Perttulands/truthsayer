package rules

import "testing"

func TestBareReturnOnError_BareReturn(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func load() (data string, err error) {
	if err != nil {
		return
	}
	return "ok", nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Rule != checker.Meta().ID {
		t.Fatalf("expected rule %s, got %s", checker.Meta().ID, findings[0].Rule)
	}
}

func TestBareReturnOnError_ExplicitZeroValues(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func load() (data string, err error) {
	if err != nil {
		return "", nil
	}
	return "ok", nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestBareReturnOnError_NamedReturnRequired(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func load() (string, error) {
	if err != nil {
		return "", nil
	}
	return "ok", nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestBareReturnOnError_NumericZeroValues(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func fetch() (count int, err error) {
	if err != nil {
		return 0, nil
	}
	return 42, nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for numeric zero return, got %d", len(findings))
	}
}

func TestBareReturnOnError_FalseZeroValue(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func check() (ok bool, err error) {
	if err != nil {
		return false, nil
	}
	return true, nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for false return, got %d", len(findings))
	}
}

func TestBareReturnOnError_RealErrorReturn_NoFinding(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
import "fmt"
func load() (data string, err error) {
	if err != nil {
		return "", fmt.Errorf("load failed: %w", err)
	}
	return "ok", nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when error is properly returned, got %d", len(findings))
	}
}

func TestBareReturnOnError_NoBody(t *testing.T) {
	checker := &BareReturnOnError{}
	// Interface method declaration — no body
	src := `package p
type Loader interface {
	Load() (string, error)
}`

	findings := runASTCheckerOnSource(t, checker, "iface.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for interface, got %d", len(findings))
	}
}

func TestBareReturnOnError_ReasonCommentSuppresses(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func load() (data string, err error) {
	if err != nil {
		// REASON: intentional zero value for cache miss
		return
	}
	return "ok", nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings when REASON comment present, got %d", len(findings))
	}
}

func TestBareReturnOnError_HexZeroLiteral(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func read() (n int, err error) {
	if err != nil {
		return 0x0, nil
	}
	return 10, nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for hex zero return, got %d", len(findings))
	}
}

func TestBareReturnOnError_OctalZeroLiteral(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func read() (n int, err error) {
	if err != nil {
		return 0o0, nil
	}
	return 10, nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for octal zero return, got %d", len(findings))
	}
}

func TestBareReturnOnError_BinaryZeroLiteral(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func read() (n int, err error) {
	if err != nil {
		return 0b0, nil
	}
	return 10, nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for binary zero return, got %d", len(findings))
	}
}

func TestBareReturnOnError_FloatZeroLiteral(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func measure() (val float64, err error) {
	if err != nil {
		return 0.0, nil
	}
	return 3.14, nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for float zero return, got %d", len(findings))
	}
}

func TestBareReturnOnError_NonZeroValues_NoFinding(t *testing.T) {
	checker := &BareReturnOnError{}
	src := `package p
func retry() (n int, err error) {
	if err != nil {
		return -1, nil
	}
	return 5, nil
}`

	findings := runASTCheckerOnSource(t, checker, "handler.go", src)
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings for non-zero return values, got %d", len(findings))
	}
}

func TestIsNumericZero(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"0", true},
		{"00", true},
		{"0x0", true},
		{"0x00", true},
		{"0b0", true},
		{"0b00", true},
		{"0o0", true},
		{"0o00", true},
		{"0_0", true},
		{"0.0", true},
		{"0.0e0", true},
		{"0e0", true},
		{"1", false},
		{"0x1", false},
		{"0b1", false},
		{"42", false},
		{"0.1", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isNumericZero(tt.input)
			if got != tt.want {
				t.Errorf("isNumericZero(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
