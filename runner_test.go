package pinky

import (
	"strings"
	"testing"
)

func TestRunSourceReturnsOutputAndDebugArtifacts(t *testing.T) {
	result := RunSource("println 1 + 2", true)

	if !result.Success {
		t.Fatalf("success = false: %+v", result)
	}
	if result.Output != "3\n" {
		t.Fatalf("output = %q", result.Output)
	}
	if len(result.Tokens) == 0 {
		t.Fatal("expected debug tokens")
	}
	if !strings.Contains(result.AST, "PrintStmt") {
		t.Fatalf("ast = %q", result.AST)
	}
}

func TestRunSourceReturnsStructuredErrors(t *testing.T) {
	result := RunSource("println missing", false)

	if result.Success {
		t.Fatalf("success = true: %+v", result)
	}
	if result.ErrorType != "runtime" {
		t.Fatalf("error type = %q", result.ErrorType)
	}
	if result.ErrorMessage != "Undeclared identifier 'missing'" {
		t.Fatalf("error message = %q", result.ErrorMessage)
	}
	if result.ErrorLine != 1 {
		t.Fatalf("error line = %d", result.ErrorLine)
	}
	if result.Output != "" {
		t.Fatalf("output = %q", result.Output)
	}
}
