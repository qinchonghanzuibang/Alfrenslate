package translate

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/cache"
)

type fake struct {
	id          string
	delay       time.Duration
	text        string
	err         error
	active, max *atomic.Int32
	calls       *atomic.Int32
}

func (f *fake) ID() string            { return f.id }
func (f *fake) Name() string          { return f.id }
func (f *fake) Configured() error     { return nil }
func (f *fake) CacheIdentity() string { return f.id }
func (f *fake) Translate(ctx context.Context, r model.Request) (model.Result, error) {
	if f.calls != nil {
		f.calls.Add(1)
	}
	if err := ctx.Err(); err != nil {
		return model.Result{}, err
	}
	if f.active != nil {
		n := f.active.Add(1)
		for {
			m := f.max.Load()
			if n <= m || f.max.CompareAndSwap(m, n) {
				break
			}
		}
		defer f.active.Add(-1)
	}
	select {
	case <-ctx.Done():
		return model.Result{}, ctx.Err()
	case <-time.After(f.delay):
	}
	return model.Result{Text: f.text, Source: "EN"}, f.err
}

func TestCacheHitAvoidsProviderCall(t *testing.T) {
	var calls atomic.Int32
	p := &fake{id: "cached", text: "你好", calls: &calls}
	e := Engine{Providers: []providers.Provider{p}, Cache: cache.New(t.TempDir(), time.Hour, true), Timeout: time.Second}
	req := model.Request{Text: "hello", Target: model.TargetZH}
	first := e.Translate(context.Background(), req)
	second := e.Translate(context.Background(), req)
	if calls.Load() != 1 || first[0].Cached || !second[0].Cached || second[0].Text != "你好" {
		t.Fatalf("calls=%d first=%+v second=%+v", calls.Load(), first, second)
	}
}

var _ providers.Provider = (*fake)(nil)

func TestConcurrentStableSuccessFailureOrdering(t *testing.T) {
	var active, max atomic.Int32
	e := Engine{Providers: []providers.Provider{&fake{id: "slow", delay: 60 * time.Millisecond, text: "slow", active: &active, max: &max}, &fake{id: "fast", delay: 5 * time.Millisecond, text: "fast", active: &active, max: &max}, &fake{id: "fail", delay: time.Millisecond, err: errors.New("boom"), active: &active, max: &max}}, Timeout: time.Second}
	got := e.Translate(context.Background(), model.Request{Text: "hello", Target: model.TargetZH})
	if max.Load() < 2 {
		t.Fatalf("providers were not concurrent: %d", max.Load())
	}
	if len(got) != 3 || got[0].Provider != "slow" || got[1].Provider != "fast" || got[2].Provider != "fail" || got[2].Err == nil {
		t.Fatalf("order=%+v", got)
	}
}
func TestTimeoutAndCancellation(t *testing.T) {
	e := Engine{Providers: []providers.Provider{&fake{id: "ok", text: "ok"}, &fake{id: "timeout", delay: time.Second, text: "late"}}, Timeout: 20 * time.Millisecond}
	got := e.Translate(context.Background(), model.Request{Text: "x", Target: model.TargetZH})
	if len(got) != 2 || got[0].Provider != "ok" || got[1].Err == nil {
		t.Fatalf("got=%+v", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got = e.Translate(ctx, model.Request{Text: "x", Target: model.TargetZH})
	for _, r := range got {
		if r.Err == nil {
			t.Fatalf("canceled result succeeded: %+v", r)
		}
	}
}
