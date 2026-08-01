package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
)

const ConfigVersion = "1"

type Store struct {
	dir     string
	ttl     time.Duration
	enabled bool
}
type entry struct {
	Created time.Time `json:"created"`
	Text    string    `json:"text"`
	Source  string    `json:"source"`
}

func New(base string, ttl time.Duration, enabled bool) *Store {
	if strings.TrimSpace(base) == "" {
		return &Store{ttl: ttl, enabled: false}
	}
	return &Store{filepath.Join(base, "translations"), ttl, enabled}
}
func Key(provider string, target model.Target, text, identity, promptVersion string) string {
	sum := sha256.Sum256([]byte(provider + "\x00" + string(target) + "\x00" + identity + "\x00" + ConfigVersion + "\x00" + promptVersion + "\x00" + text))
	return hex.EncodeToString(sum[:])
}
func (s *Store) Get(key string) (model.Result, bool) {
	if !s.enabled {
		return model.Result{}, false
	}
	b, e := os.ReadFile(filepath.Join(s.dir, key+".json"))
	if e != nil {
		return model.Result{}, false
	}
	var x entry
	if json.Unmarshal(b, &x) != nil || x.Text == "" || time.Since(x.Created) > s.ttl {
		return model.Result{}, false
	}
	return model.Result{Text: x.Text, Source: x.Source, Cached: true}, true
}
func (s *Store) Put(key string, r model.Result) error {
	if !s.enabled {
		return nil
	}
	if e := os.MkdirAll(s.dir, 0700); e != nil {
		return e
	}
	b, e := json.Marshal(entry{time.Now().UTC(), r.Text, r.Source})
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(s.dir, ".cache-*")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	if e = f.Chmod(0600); e == nil {
		_, e = f.Write(b)
	}
	if closeErr := f.Close(); e == nil {
		e = closeErr
	}
	if e != nil {
		return e
	}
	return os.Rename(name, filepath.Join(s.dir, key+".json"))
}
func (s *Store) Clear() error {
	if s.dir == "" || filepath.Base(filepath.Clean(s.dir)) != "translations" {
		return errors.New("refusing to clear an unsafe cache path")
	}
	entries, e := os.ReadDir(s.dir)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	for _, x := range entries {
		if e := os.RemoveAll(filepath.Join(s.dir, x.Name())); e != nil {
			return e
		}
	}
	return nil
}
func (s *Store) Dir() string { return s.dir }
