package docs

import (
	"testing"

	"github.com/swaggo/swag"
)

func TestSwaggerRegistration(t *testing.T) {
	if SwaggerInfo == nil {
		t.Fatalf("SwaggerInfo is nil")
	}
	name := SwaggerInfo.InstanceName()
	if name == "" {
		t.Fatalf("expected non-empty instance name")
	}
	if doc, err := swag.ReadDoc(name); err != nil {
		t.Fatalf("expected swagger doc to be readable: %v", err)
	} else if doc == "" {
		t.Fatalf("expected non-empty swagger doc")
	}
}
