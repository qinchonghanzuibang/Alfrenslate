package deepseek

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

func TestDeepSeekEndpointThinkingAndReasoningIgnored(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") || strings.Contains(r.URL.Path, "/v1/") {
			t.Errorf("path=%s", r.URL.Path)
		}
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		thinking, ok := b["thinking"].(map[string]any)
		if !ok || thinking["type"] != "disabled" {
			t.Errorf("thinking=%v", b["thinking"])
		}
		if b["model"] != "deepseek-v4-flash" {
			t.Errorf("model=%v", b["model"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"reasoning_content":"do not return me","content":"你好"}}]}`))
	}))
	defer s.Close()
	t.Setenv("DEEPSEEK_BASE_URL", s.URL)
	t.Setenv("DEEPSEEK_MODEL", "")
	t.Setenv("DEEPSEEK_API_KEY", "key")
	p, e := New(providers.NewClient(time.Second, 0))
	if e != nil {
		t.Fatal(e)
	}
	got, e := p.Translate(context.Background(), model.Request{Text: "hello", Target: model.TargetZH})
	if e != nil || got.Text != "你好" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
