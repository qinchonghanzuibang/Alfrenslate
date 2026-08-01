package deepl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

func TestEndpointsAndTranslate(t *testing.T) {
	client := providers.NewClient(time.Second, 0)
	t.Setenv("DEEPL_API_TIER", "free")
	if got := New(client).(*Provider).endpoint; got != "https://api-free.deepl.com/v2/translate" {
		t.Fatalf("free endpoint=%s", got)
	}
	t.Setenv("DEEPL_API_TIER", "pro")
	if got := New(client).(*Provider).endpoint; got != "https://api.deepl.com/v2/translate" {
		t.Fatalf("pro endpoint=%s", got)
	}
	t.Setenv("DEEPL_API_TIER", "custom")
	t.Setenv("DEEPL_BASE_URL", "https://proxy.example/")
	if got := New(client).(*Provider).endpoint; got != "https://proxy.example/v2/translate" {
		t.Fatalf("custom endpoint=%s", got)
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "DeepL-Auth-Key key" {
			t.Errorf("auth=%q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"translations":[{"detected_source_language":"EN","text":"第一段"},{"detected_source_language":"EN","text":"第二段"}]}`))
	}))
	defer s.Close()
	p := &Provider{client: providers.NewClient(time.Second, 0), key: "key", tier: "custom", endpoint: s.URL}
	got, e := p.Translate(context.Background(), model.Request{Text: "one\ntwo", Target: model.TargetZH})
	if e != nil || got.Text != "第一段\n第二段" || got.Source != "EN" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
func TestTierMismatchMessage(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "forbidden", 403) }))
	defer s.Close()
	p := &Provider{client: providers.NewClient(time.Second, 0), key: "key", tier: "free", endpoint: s.URL}
	_, e := p.Translate(context.Background(), model.Request{Text: "hello", Target: model.TargetZH})
	if e == nil || e.Error() != "DeepL authentication or API tier endpoint mismatch; check Free/Pro selection" {
		t.Fatalf("err=%v", e)
	}
}
