package products

import "testing"

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{Field: "name", Message: "msg"}
	if err.Error() != "msg" {
		t.Fatalf("expected message %q, got %q", "msg", err.Error())
	}
}
