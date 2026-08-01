package google

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

func TestTranslateUnescapesHTML(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "key" {
			t.Error("missing key")
		}
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		if b["target"] != "zh-CN" {
			t.Errorf("target=%v", b["target"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"translations":[{"detectedSourceLanguage":"en","translatedText":"Tom &amp; Jerry&#39;s"}]}}`))
	}))
	defer s.Close()
	p := &Provider{providers.NewClient(time.Second, 0), "key", s.URL}
	got, e := p.Translate(context.Background(), model.Request{Text: "Tom & Jerry's", Target: model.TargetZH})
	if e != nil || got.Text != "Tom & Jerry's" || got.Source != "EN" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
