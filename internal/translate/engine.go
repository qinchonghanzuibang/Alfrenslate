package translate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/cache"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/prompt"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

type Engine struct {
	Providers []providers.Provider
	Cache     *cache.Store
	Timeout   time.Duration
}

func (e *Engine) Translate(ctx context.Context, req model.Request) []model.Result {
	results := make([]model.Result, len(e.Providers))
	var wg sync.WaitGroup
	for i, p := range e.Providers {
		i, p := i, p
		wg.Add(1)
		go func() { defer wg.Done(); results[i] = e.one(ctx, p, req) }()
	}
	wg.Wait()
	ok := make([]model.Result, 0, len(results))
	failed := make([]model.Result, 0, len(results))
	for _, r := range results {
		if r.Err == nil {
			ok = append(ok, r)
		} else {
			failed = append(failed, r)
		}
	}
	return append(ok, failed...)
}
func (e *Engine) one(parent context.Context, p providers.Provider, req model.Request) model.Result {
	r := model.Result{Provider: p.ID(), Target: req.Target}
	if err := p.Configured(); err != nil {
		r.Err = err
		return r
	}
	key := cache.Key(p.ID(), req.Target, req.Text, p.CacheIdentity(), prompt.Version)
	if e.Cache != nil {
		if x, ok := e.Cache.Get(key); ok {
			x.Provider = p.ID()
			x.Target = req.Target
			return x
		}
	}
	ctx := parent
	cancel := func() {}
	if e.Timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, e.Timeout)
	}
	defer cancel()
	start := time.Now()
	x, err := p.Translate(ctx, req)
	x.Elapsed = time.Since(start)
	x.Provider = p.ID()
	x.Target = req.Target
	if err != nil {
		x.Err = fmt.Errorf("%s", err)
		return x
	}
	if x.Text == "" {
		x.Err = fmt.Errorf("empty translation")
		return x
	}
	if e.Cache != nil {
		_ = e.Cache.Put(key, x)
	}
	return x
}
