package domain

import "testing"

func TestValidAuditResult(t *testing.T) {
	t.Parallel()
	if !ValidAuditResult(AuditResultSuccess) || !ValidAuditResult(AuditResultFailure) || !ValidAuditResult(AuditResultError) {
		t.Fatal("expected known results to be valid")
	}
	if ValidAuditResult("nope") || ValidAuditResult("") {
		t.Fatal("expected unknown results to be invalid")
	}
}

func TestAuditResultFromBool(t *testing.T) {
	t.Parallel()
	if got := AuditResultFromBool(true); got != AuditResultSuccess {
		t.Fatalf("true: got %q", got)
	}
	if got := AuditResultFromBool(false); got != AuditResultFailure {
		t.Fatalf("false: got %q", got)
	}
}
