package errors

import (
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	pkgerrs "github.com/pkg/errors"
)

func TestWrap_DoesNotMutateOriginalAppError(t *testing.T) {
	orig := ValidationError("E1", "original message")
	origMsg := orig.Message

	wrapped := Wrap(orig, "context")
	if wrapped == nil {
		t.Fatal("expected non-nil wrapped error")
	}

	if orig.Message != "original message" {
		t.Fatalf("original AppError was mutated: got %q want %q", orig.Message, origMsg)
	}

	out, ok := wrapped.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", wrapped)
	}
	if !strings.Contains(out.Message, "context") || !strings.Contains(out.Message, "original message") {
		t.Fatalf("unexpected wrapped message: %q", out.Message)
	}
	if !stderrors.Is(out, orig) {
		t.Fatal("expected errors.Is(wrapped, orig)")
	}
}

func TestWrap_NonAppError_UsesPkgErrors(t *testing.T) {
	base := fmt.Errorf("base")
	w := Wrap(base, "wrapped")
	if w == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(w.Error(), "wrapped") {
		t.Fatalf("expected context in error: %v", w)
	}
}

func TestCause_DelegatesToRootWithoutExtraWrapper(t *testing.T) {
	root := fmt.Errorf("root")
	chain := pkgerrs.Wrap(root, "layer")

	got := Cause(chain)
	if got == nil {
		t.Fatal("expected cause")
	}
	want := pkgerrs.Cause(chain)
	if got.Error() != want.Error() {
		t.Fatalf("Cause mismatch: got %v want %v", got, want)
	}
}

func TestCause_Nil(t *testing.T) {
	if Cause(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestIsValidationError_WrappedChain(t *testing.T) {
	inner := ValidationError("V", "bad")
	outer := fmt.Errorf("outer: %w", inner)
	if !IsValidationError(outer) {
		t.Fatal("expected validation error in chain")
	}
	if IsNotFoundError(outer) {
		t.Fatal("unexpected not found")
	}
}

func TestIsNotFoundError_AsChain(t *testing.T) {
	inner := NotFoundError("user", "1")
	outer := Wrap(inner, "handler")
	if !IsNotFoundError(outer) {
		t.Fatal("expected not found in chain")
	}
}

func TestIsDatabaseError(t *testing.T) {
	dbErr := fmt.Errorf("pq: syntax")
	app := DatabaseError("select", dbErr)
	wrapped := fmt.Errorf("repo: %w", app)
	if !IsDatabaseError(wrapped) {
		t.Fatal("expected database error in chain")
	}
}

func TestGetAppError(t *testing.T) {
	inner := ConflictError("user", "x@y.com")
	outer := fmt.Errorf("wrap: %w", inner)
	got, ok := GetAppError(outer)
	if !ok {
		t.Fatal("expected AppError in chain")
	}
	if got.Type != ErrorTypeConflict {
		t.Fatalf("type: got %v want %v", got.Type, ErrorTypeConflict)
	}
}

func TestIsServiceUnavailableError(t *testing.T) {
	e := ServiceUnavailableError("email", "timeout")
	if !IsServiceUnavailableError(fmt.Errorf("x: %w", e)) {
		t.Fatal("expected service unavailable in chain")
	}
}

func TestWithDetails_DoesNotMutateOriginal(t *testing.T) {
	orig := ValidationError("C", "msg")
	_ = orig.WithDetails("extra")
	if orig.Details != "" {
		t.Fatalf("original mutated: Details=%q", orig.Details)
	}
	out := orig.WithDetails("d1")
	if out.Details != "d1" {
		t.Fatalf("copy Details: %q", out.Details)
	}
}

func TestIsConflictError(t *testing.T) {
	e := ConflictError("user", "a@b.com")
	if !IsConflictError(fmt.Errorf("w: %w", e)) {
		t.Fatal("expected conflict in chain")
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	e := UnauthorizedError("no token")
	if !IsUnauthorizedError(fmt.Errorf("w: %w", e)) {
		t.Fatal("expected unauthorized in chain")
	}
}

func TestIsInternalError(t *testing.T) {
	e := InternalError("boom", fmt.Errorf("root"))
	if !IsInternalError(fmt.Errorf("w: %w", e)) {
		t.Fatal("expected internal in chain")
	}
}

func TestDatabaseError_NilUnderlying(t *testing.T) {
	if DatabaseError("op", nil) != nil {
		t.Fatal("expected nil when underlying err is nil")
	}
}
