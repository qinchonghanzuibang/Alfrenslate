package alfred

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

func TestTranslationItems(t *testing.T) {
	all := []providers.Provider{stub{"deepseek", "DeepSeek"}, stub{"deepl", "DeepL"}}
	r := []model.Result{{Provider: "deepseek", Text: "你好", Source: "EN", Target: model.TargetZH, Elapsed: 428 * time.Millisecond}, {Provider: "deepl", Target: model.TargetZH, Err: errors.New("authentication failed")}}
	out := TranslationItems(r, all)
	b, e := json.Marshal(out)
	if e != nil || !json.Valid(b) {
		t.Fatalf("json=%s err=%v", b, e)
	}
	if len(out.Items) != 2 {
		t.Fatal("item count")
	}
	ok := out.Items[0]
	if ok.Title != "你好" || ok.Arg != "你好" || !ok.Valid || ok.Mods["cmd"].Arg != "你好" || ok.Mods["alt"].Arg != "你好" {
		t.Fatalf("success=%+v", ok)
	}
	if out.Items[1].Valid || out.Items[1].Arg != "" {
		t.Fatalf("failure=%+v", out.Items[1])
	}
}

type stub struct{ id, name string }

func (s stub) ID() string            { return s.id }
func (s stub) Name() string          { return s.name }
func (s stub) Configured() error     { return nil }
func (s stub) CacheIdentity() string { return "" }
func (s stub) Translate(_ context.Context, _ model.Request) (model.Result, error) {
	return model.Result{}, nil
}
