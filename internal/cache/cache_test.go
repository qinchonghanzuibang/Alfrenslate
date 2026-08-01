package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
)

func TestCacheLifecycleAndCredentialFreeKey(t *testing.T) {
	base := t.TempDir()
	s := New(base, time.Hour, true)
	key := Key("p", model.TargetZH, "hello", "model|url", "1")
	if strings.Contains(key, "hello") || strings.Contains(key, "secret") {
		t.Fatalf("unsafe key=%s", key)
	}
	r := model.Result{Text: "你好", Source: "EN"}
	if e := s.Put(key, r); e != nil {
		t.Fatal(e)
	}
	got, ok := s.Get(key)
	if !ok || got.Text != "你好" || !got.Cached {
		t.Fatalf("got=%+v ok=%v", got, ok)
	}
	if e := s.Clear(); e != nil {
		t.Fatal(e)
	}
	if _, ok := s.Get(key); ok {
		t.Fatal("cache survived clear")
	}
}
func TestExpiredDisabledCorruptAndSafeClear(t *testing.T) {
	base := t.TempDir()
	expired := New(base, -time.Second, true)
	if e := expired.Put("old", model.Result{Text: "x"}); e != nil {
		t.Fatal(e)
	}
	if _, ok := expired.Get("old"); ok {
		t.Fatal("expired hit")
	}
	disabled := New(base, time.Hour, false)
	if e := disabled.Put("disabled", model.Result{Text: "x"}); e != nil {
		t.Fatal(e)
	}
	if _, ok := disabled.Get("disabled"); ok {
		t.Fatal("disabled hit")
	}
	dir := filepath.Join(base, "translations")
	_ = os.MkdirAll(dir, 0700)
	_ = os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{"), 0600)
	if _, ok := New(base, time.Hour, true).Get("bad"); ok {
		t.Fatal("corrupt hit")
	}
	outside := filepath.Join(base, "outside.txt")
	_ = os.WriteFile(outside, []byte("keep"), 0600)
	if e := New(base, time.Hour, true).Clear(); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(outside); e != nil {
		t.Fatal("clear removed outside file")
	}
	unsafe := &Store{dir: base, enabled: true}
	if e := unsafe.Clear(); e == nil {
		t.Fatal("unsafe clear accepted")
	}
}
