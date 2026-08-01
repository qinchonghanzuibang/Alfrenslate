package security

import (
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	secret := "super-secret-token"
	in := "Authorization: Bearer demo-token api_key=" + secret + " https://example.com/v1?key=" + secret + "&q=x"
	got := Redact(in, secret)
	if strings.Contains(got, secret) || strings.Contains(got, "demo-token") {
		t.Fatalf("credential leaked: %s", got)
	}
}
