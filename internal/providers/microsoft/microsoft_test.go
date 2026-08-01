package microsoft

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

func TestNormalize(t *testing.T) {
	got, e := normalize("https://example.com/")
	if e != nil || got != "https://example.com/translate?api-version=3.0" {
		t.Fatalf("got=%q err=%v", got, e)
	}
	got, e = normalize("https://example.com/translate?x=y")
	if e != nil || got != "https://example.com/translate?api-version=3.0&x=y" {
		t.Fatalf("got=%q err=%v", got, e)
	}
}
func TestTranslateHeadersAndDetection(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Ocp-Apim-Subscription-Key") != "key" || r.Header.Get("Ocp-Apim-Subscription-Region") != "eastasia" {
			t.Errorf("headers=%v", r.Header)
		}
		if r.URL.Query().Get("to") != "zh-Hans" || r.URL.Query().Get("api-version") != "3.0" {
			t.Errorf("query=%v", r.URL.Query())
		}
		var body []map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if len(body) != 1 || body[0]["Text"] != "hello" {
			t.Errorf("body=%v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"detectedLanguage":{"language":"en","score":1},"translations":[{"text":"你好","to":"zh-Hans"}]}]`))
	}))
	defer s.Close()
	p := &Provider{providers.NewClient(time.Second, 0), "key", "eastasia", s.URL + "/translate?api-version=3.0"}
	got, e := p.Translate(context.Background(), model.Request{Text: "hello", Target: model.TargetZH})
	if e != nil || got.Text != "你好" || got.Source != "EN" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
