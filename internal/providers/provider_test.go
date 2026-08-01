package providers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
)

func TestNormalizeChatURL(t *testing.T) {
	tests := map[string]string{"https://api.example.com": "https://api.example.com/v1/chat/completions", "https://api.example.com/": "https://api.example.com/v1/chat/completions", "https://api.example.com/v1": "https://api.example.com/v1/chat/completions", "https://api.example.com/v1/": "https://api.example.com/v1/chat/completions", "https://api.example.com/v1/chat/completions": "https://api.example.com/v1/chat/completions"}
	for in, want := range tests {
		got, e := NormalizeChatURL(in)
		if e != nil || got != want {
			t.Errorf("NormalizeChatURL(%q)=%q,%v want %q", in, got, e, want)
		}
	}
	if _, e := NormalizeChatURL("relative"); e == nil {
		t.Fatal("relative URL accepted")
	}
}
func TestChatProviderStringAndArrayContent(t *testing.T) {
	responses := []string{`{"choices":[{"message":{"content":"你好","reasoning_content":"secret reasoning"}}]}`, `{"choices":[{"message":{"content":[{"type":"text","text":"你"},{"type":"output_text","text":{"value":"好"}},{"type":"reasoning","text":"ignore"}]}}]}`}
	for _, response := range responses {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("method=%s", r.Method)
			}
			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Errorf("authorization missing")
			}
			var body map[string]any
			if json.NewDecoder(r.Body).Decode(&body) != nil {
				t.Fatal("invalid request JSON")
			}
			if body["stream"] != false {
				t.Error("stream must be false")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(response))
		}))
		p := &ChatProvider{ProviderID: "test", ProviderName: "Test", Endpoint: s.URL, Key: "test-key", Model: "model", Client: NewClient(time.Second, 0)}
		r, e := p.Translate(context.Background(), model.Request{Text: "hello", Target: model.TargetZH})
		s.Close()
		if e != nil || r.Text != "你好" {
			t.Fatalf("result=%+v err=%v", r, e)
		}
	}
}
func TestClientRetryAfterAndCancellation(t *testing.T) {
	var n atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			http.Error(w, "busy", 429)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer s.Close()
	c := NewClient(time.Second, 1)
	c.BaseDelay = time.Millisecond
	var out map[string]bool
	if e := c.JSON(context.Background(), "POST", s.URL, nil, map[string]string{}, &out); e != nil || !out["ok"] || n.Load() != 2 {
		t.Fatalf("out=%v calls=%d err=%v", out, n.Load(), e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := c.JSON(ctx, "POST", s.URL, nil, nil, &out); e == nil {
		t.Fatal("canceled request succeeded")
	}
}
func TestClientDefensiveResponses(t *testing.T) {
	for name, handler := range map[string]http.HandlerFunc{"empty": func(w http.ResponseWriter, r *http.Request) { w.Header().Set("Content-Type", "application/json") }, "html": func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<h1>error</h1>"))
	}, "malformed": func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{"))
	}} {
		t.Run(name, func(t *testing.T) {
			s := httptest.NewServer(handler)
			defer s.Close()
			var out any
			if e := NewClient(time.Second, 0).JSON(context.Background(), "POST", s.URL, nil, map[string]string{}, &out); e == nil {
				t.Fatal("invalid response accepted")
			}
		})
	}
}

func TestClientStatusErrorsAndTimeout(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 429, 500, 502} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "<html>credential echo</html>", status) }))
			defer s.Close()
			var out any
			err := NewClient(time.Second, 0).JSON(context.Background(), "POST", s.URL, nil, map[string]string{}, &out)
			var h *HTTPError
			if !errors.As(err, &h) || h.Status != status || strings.Contains(err.Error(), "credential echo") {
				t.Fatalf("status=%d err=%v", status, err)
			}
		})
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(100 * time.Millisecond) }))
	defer s.Close()
	var out any
	err := NewClient(5*time.Millisecond, 0).JSON(context.Background(), "POST", s.URL, nil, map[string]string{}, &out)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout err=%v", err)
	}
}

func TestFormRetriesRecoverableStatus(t *testing.T) {
	var calls atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			http.Error(w, "temporary", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer s.Close()
	c := NewClient(time.Second, 1)
	c.BaseDelay = time.Millisecond
	var out map[string]bool
	if err := c.Form(context.Background(), s.URL, nil, url.Values{"q": {"hello"}}, &out); err != nil || !out["ok"] || calls.Load() != 2 {
		t.Fatalf("out=%v calls=%d err=%v", out, calls.Load(), err)
	}
}
