package youdao

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

func TestSignatureVectors(t *testing.T) {
	if got := TruncateInput("1234567890123456789012345"); got != "1234567890256789012345" {
		t.Fatalf("truncate=%q", got)
	}
	if got := Sign("app", "hello", "salt", "1700000000", "secret"); got != "7663eff532d251908c402945c0075f8dc5cb21be0d6b13285f7c98acf9bbc6f3" {
		t.Fatalf("sign=%s", got)
	}
}
func TestTranslateRequestAndResponse(t *testing.T) {
	var form url.Values
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		form = r.Form
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errorCode":"0","translation":["你好"],"l":"en2zh-CHS"}`))
	}))
	defer s.Close()
	p := &Provider{client: providers.NewClient(time.Second, 0), key: "app", secret: "secret", domain: "general", endpoint: s.URL}
	got, e := p.Translate(context.Background(), model.Request{Text: "hello", Target: model.TargetZH})
	if e != nil || got.Text != "你好" || got.Source != "EN" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
	for _, k := range []string{"q", "from", "to", "appKey", "salt", "sign", "signType", "curtime", "domain"} {
		if form.Get(k) == "" {
			t.Errorf("missing %s", k)
		}
	}
	if form.Get("from") != "auto" || form.Get("to") != "zh-CHS" {
		t.Errorf("language mapping: %v", form)
	}
}
