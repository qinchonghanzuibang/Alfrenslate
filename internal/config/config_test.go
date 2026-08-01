package config

import (
	"testing"
	"time"
)

func TestDefaultsAndBounds(t *testing.T) {
	for _, k := range []string{"PROVIDER_ORDER", "REQUEST_TIMEOUT_SECONDS", "RETRY_COUNT", "CACHE_ENABLED", "CACHE_TTL_HOURS", "alfred_workflow_cache"} {
		t.Setenv(k, "")
	}
	c := Load()
	if len(c.Order) != 7 || c.Timeout != 8*time.Second || c.Retries != 1 || !c.CacheEnabled || c.CacheTTL != 24*time.Hour {
		t.Fatalf("defaults=%+v", c)
	}
	t.Setenv("REQUEST_TIMEOUT_SECONDS", "999")
	t.Setenv("RETRY_COUNT", "-1")
	if c := Load(); c.Timeout != 8*time.Second || c.Retries != 1 {
		t.Fatalf("invalid bounds accepted: %+v", c)
	}
}
func TestEnabledMapping(t *testing.T) {
	t.Setenv("OPENAI_COMPATIBLE_ENABLED", "true")
	if !Enabled("openai-compatible") {
		t.Fatal("hyphen mapping failed")
	}
}
