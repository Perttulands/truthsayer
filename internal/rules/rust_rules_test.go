package rules

import "testing"

// --- RustUnwrapInProduction ---

func TestRustUnwrapInProduction_Triggers(t *testing.T) {
	checker := &RustUnwrapInProduction{}
	lines := []string{
		`fn main() {`,
		`    let x = some_result().unwrap();`,
		`}`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 2 {
		t.Errorf("expected line 2, got %d", findings[0].Line)
	}
}

func TestRustUnwrapInProduction_SkipsTestDir(t *testing.T) {
	checker := &RustUnwrapInProduction{}
	lines := []string{`let x = foo().unwrap();`}
	findings := checker.CheckLines("tests/integration.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for test dir, got %d", len(findings))
	}
}

func TestRustUnwrapInProduction_SkipsCfgTest(t *testing.T) {
	checker := &RustUnwrapInProduction{}
	lines := []string{
		`#[cfg(test)]`,
		`mod tests {`,
		`    let x = foo().unwrap();`,
		`}`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for #[cfg(test)], got %d", len(findings))
	}
}

func TestRustUnwrapInProduction_SkipsComments(t *testing.T) {
	checker := &RustUnwrapInProduction{}
	lines := []string{`// foo().unwrap() is fine in comments`}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for comment, got %d", len(findings))
	}
}

// --- RustIgnoredResult ---

func TestRustIgnoredResult_Triggers(t *testing.T) {
	checker := &RustIgnoredResult{}
	lines := []string{
		`fn main() {`,
		`    let _ = file.write_all(b"data");`,
		`}`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 2 {
		t.Errorf("expected line 2, got %d", findings[0].Line)
	}
}

func TestRustIgnoredResult_SkipsTestFile(t *testing.T) {
	checker := &RustIgnoredResult{}
	lines := []string{`    let _ = something();`}
	findings := checker.CheckLines("tests/foo.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for test file, got %d", len(findings))
	}
}

// --- RustPanicInLib ---

func TestRustPanicInLib_Triggers(t *testing.T) {
	checker := &RustPanicInLib{}
	lines := []string{
		`pub fn validate(x: &str) {`,
		`    panic!("invalid input");`,
		`}`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestRustPanicInLib_SkipsMainRs(t *testing.T) {
	checker := &RustPanicInLib{}
	lines := []string{`panic!("fatal");`}
	findings := checker.CheckLines("src/main.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for main.rs, got %d", len(findings))
	}
}

func TestRustPanicInLib_SkipsTestDir(t *testing.T) {
	checker := &RustPanicInLib{}
	lines := []string{`panic!("test");`}
	findings := checker.CheckLines("tests/foo.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for test dir, got %d", len(findings))
	}
}

// --- RustUnsafeNoComment ---

func TestRustUnsafeNoComment_Triggers(t *testing.T) {
	checker := &RustUnsafeNoComment{}
	lines := []string{
		`fn dangerous() {`,
		`    unsafe {`,
		`        ptr::read(p);`,
		`    }`,
		`}`,
	}
	findings := checker.CheckLines("src/ffi.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Line != 2 {
		t.Errorf("expected line 2, got %d", findings[0].Line)
	}
}

func TestRustUnsafeNoComment_WithSafetyComment(t *testing.T) {
	checker := &RustUnsafeNoComment{}
	lines := []string{
		`fn safe_wrapper() {`,
		`    // SAFETY: pointer is valid and aligned, guaranteed by caller`,
		`    unsafe {`,
		`        ptr::read(p);`,
		`    }`,
		`}`,
	}
	findings := checker.CheckLines("src/ffi.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings with SAFETY comment, got %d", len(findings))
	}
}

// --- RustTodoInProduction ---

func TestRustTodoInProduction_Triggers(t *testing.T) {
	checker := &RustTodoInProduction{}
	lines := []string{
		`fn compute() -> i32 {`,
		`    todo!("implement this")`,
		`}`,
	}
	findings := checker.CheckLines("src/math.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != "warning" {
		t.Errorf("expected warning severity, got %s", findings[0].Severity)
	}
}

func TestRustTodoInProduction_UnimplementedTriggers(t *testing.T) {
	checker := &RustTodoInProduction{}
	lines := []string{`    unimplemented!("not done")`}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestRustTodoInProduction_SkipsTestDir(t *testing.T) {
	checker := &RustTodoInProduction{}
	lines := []string{`todo!("test placeholder")`}
	findings := checker.CheckLines("tests/foo.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for test dir, got %d", len(findings))
	}
}

// --- RustExpectGeneric ---

func TestRustExpectGeneric_Triggers(t *testing.T) {
	checker := &RustExpectGeneric{}
	lines := []string{
		`let val = opt.expect("error");`,
		`let x = res.expect("failed");`,
		`let y = r.expect("none");`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
}

func TestRustExpectGeneric_DescriptiveOk(t *testing.T) {
	checker := &RustExpectGeneric{}
	lines := []string{
		`let val = opt.expect("failed to parse config file");`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for descriptive expect, got %d", len(findings))
	}
}

// --- RustBoxedError ---

func TestRustBoxedError_Triggers(t *testing.T) {
	checker := &RustBoxedError{}
	lines := []string{
		`fn run() -> Result<(), Box<dyn Error>> {`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestRustBoxedError_StdError(t *testing.T) {
	checker := &RustBoxedError{}
	lines := []string{
		`fn run() -> Result<(), Box<dyn std::error::Error>> {`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

// --- RustProcessExit ---

func TestRustProcessExit_Triggers(t *testing.T) {
	checker := &RustProcessExit{}
	lines := []string{
		`    std::process::exit(1);`,
		`    process::exit(0);`,
	}
	findings := checker.CheckLines("src/main.rs", lines)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

// --- RustAllowSuppress ---

func TestRustAllowSuppress_TriggersNoComment(t *testing.T) {
	checker := &RustAllowSuppress{}
	lines := []string{
		`#[allow(dead_code)]`,
		`fn old_function() {}`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestRustAllowSuppress_WithComment(t *testing.T) {
	checker := &RustAllowSuppress{}
	lines := []string{
		`#[allow(dead_code)]`,
		`// kept for backwards compatibility with v1 API`,
		`fn old_function() {}`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings with explanation comment, got %d", len(findings))
	}
}

// --- RustEprintlnInLib ---

func TestRustEprintlnInLib_TriggersInLibRs(t *testing.T) {
	checker := &RustEprintlnInLib{}
	lines := []string{
		`pub fn init() {`,
		`    eprintln!("initializing");`,
		`}`,
	}
	findings := checker.CheckLines("src/lib.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestRustEprintlnInLib_TriggersInLibDir(t *testing.T) {
	checker := &RustEprintlnInLib{}
	lines := []string{`    eprintln!("debug");`}
	findings := checker.CheckLines("crate/lib/utils.rs", lines)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for lib/ dir, got %d", len(findings))
	}
}

func TestRustEprintlnInLib_SkipsNonLib(t *testing.T) {
	checker := &RustEprintlnInLib{}
	lines := []string{`    eprintln!("error");`}
	findings := checker.CheckLines("src/main.rs", lines)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for non-lib file, got %d", len(findings))
	}
}

// --- Helper tests ---

func TestRsIsTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"tests/foo.rs", true},
		{"src/tests/bar.rs", true},
		{"test/integration.rs", true},
		{"src/lib.rs", false},
		{"src/main.rs", false},
		{"foo_test.rs", true},
	}
	for _, tt := range tests {
		got := rsIsTestFile(tt.path)
		if got != tt.want {
			t.Errorf("rsIsTestFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
