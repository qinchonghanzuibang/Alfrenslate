package baidu

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

func TestSignatureVector(t *testing.T) {
	if got := Sign("appid", "hello", "1435660288", "secret"); got != "1cf8da6c88f774a7ccd83ec44fe43b2f" {
		t.Fatalf("sign=%s", got)
	}
}
func TestTranslateMultiResult(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if e := r.ParseForm(); e != nil {
			t.Fatal(e)
		}
		if r.Form.Get("from") != "auto" || r.Form.Get("to") != "en" {
			t.Errorf("form=%v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"from":"zh","trans_result":[{"dst":"first"},{"dst":"second"}]}`))
	}))
	defer s.Close()
	p := &Provider{providers.NewClient(time.Second, 0), "appid", "secret", s.URL}
	got, e := p.Translate(context.Background(), model.Request{Text: "第一\n第二", Target: model.TargetEN})
	if e != nil || got.Text != "first\nsecond" || got.Source != "ZH" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
