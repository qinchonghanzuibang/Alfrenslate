package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
)

type Provider interface {
	ID() string
	Name() string
	Configured() error
	CacheIdentity() string
	Translate(context.Context, model.Request) (model.Result, error)
}

type Client struct {
	HTTP      *http.Client
	Retries   int
	UserAgent string
	MaxBody   int64
	BaseDelay time.Duration
}

func NewClient(timeout time.Duration, retries int) *Client {
	return &Client{HTTP: &http.Client{Timeout: timeout, Transport: http.DefaultTransport}, Retries: retries, UserAgent: "Alfrenslate/1.0.0", MaxBody: 2 << 20, BaseDelay: 150 * time.Millisecond}
}

type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string { return e.Message }

func (c *Client) JSON(ctx context.Context, method, endpoint string, headers map[string]string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("invalid endpoint: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)
		for k, v := range headers {
			if v != "" {
				req.Header.Set(k, v)
			}
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			if attempt < c.Retries && retryableError(err) {
				if err := wait(ctx, c.delay(attempt, "")); err != nil {
					return err
				}
				continue
			}
			if errors.Is(err, context.Canceled) {
				return context.Canceled
			}
			if errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("request timed out: %w", context.DeadlineExceeded)
			}
			return errors.New("network request failed; check the endpoint and connection")
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, c.MaxBody+1))
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read response: %w", readErr)
		}
		if int64(len(data)) > c.MaxBody {
			return errors.New("provider response is too large")
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if attempt < c.Retries && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
				if err := wait(ctx, c.delay(attempt, resp.Header.Get("Retry-After"))); err != nil {
					return err
				}
				continue
			}
			return &HTTPError{Status: resp.StatusCode, Message: friendlyHTTP(resp.StatusCode, data)}
		}
		ct := strings.ToLower(resp.Header.Get("Content-Type"))
		if ct != "" && !strings.Contains(ct, "json") {
			return errors.New("provider returned a non-JSON response")
		}
		if len(bytes.TrimSpace(data)) == 0 {
			return errors.New("provider returned an empty response")
		}
		if err := json.Unmarshal(data, out); err != nil {
			return errors.New("provider returned malformed JSON")
		}
		return nil
	}
}

func (c *Client) Form(ctx context.Context, endpoint string, headers map[string]string, values url.Values, out any) error {
	body := values.Encode()
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body))
		if err != nil {
			return fmt.Errorf("invalid endpoint: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)
		for k, v := range headers {
			if v != "" {
				req.Header.Set(k, v)
			}
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			if attempt < c.Retries && retryableError(err) {
				if err := wait(ctx, c.delay(attempt, "")); err != nil {
					return err
				}
				continue
			}
			if errors.Is(err, context.Canceled) {
				return context.Canceled
			}
			if errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("request timed out: %w", context.DeadlineExceeded)
			}
			return errors.New("network request failed; check the endpoint and connection")
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, c.MaxBody+1))
		resp.Body.Close()
		if readErr != nil {
			return readErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if attempt < c.Retries && (resp.StatusCode == 429 || resp.StatusCode >= 500) {
				if err := wait(ctx, c.delay(attempt, resp.Header.Get("Retry-After"))); err != nil {
					return err
				}
				continue
			}
			return &HTTPError{Status: resp.StatusCode, Message: friendlyHTTP(resp.StatusCode, data)}
		}
		if len(bytes.TrimSpace(data)) == 0 {
			return errors.New("provider returned an empty response")
		}
		if err := json.Unmarshal(data, out); err != nil {
			return errors.New("provider returned malformed JSON")
		}
		return nil
	}
}

func (c *Client) delay(attempt int, retryAfter string) time.Duration {
	if n, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && n > 0 {
		d := time.Duration(n) * time.Second
		if d > 3*time.Second {
			return 3 * time.Second
		}
		return d
	}
	if when, err := http.ParseTime(strings.TrimSpace(retryAfter)); err == nil {
		d := time.Until(when)
		if d > 0 {
			if d > 3*time.Second {
				return 3 * time.Second
			}
			return d
		}
	}
	d := c.BaseDelay * time.Duration(1<<attempt)
	if d > 2*time.Second {
		return 2 * time.Second
	}
	return d
}
func wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
func retryableError(err error) bool {
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}
func friendlyHTTP(status int, data []byte) string {
	switch status {
	case 400:
		return "invalid request or provider configuration"
	case 401:
		return "authentication failed; check credentials"
	case 403:
		return "access denied; check credentials, endpoint, and quota"
	case 404:
		return "endpoint or model was not found"
	case 429:
		return "rate limit or quota exceeded"
	}
	if status >= 500 {
		return "provider is temporarily unavailable"
	}
	_ = data
	return fmt.Sprintf("provider returned HTTP %d", status)
}

func NormalizeChatURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("base URL is required")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("base URL must be an absolute HTTP(S) URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("base URL must use HTTP or HTTPS")
	}
	p := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(p, "/chat/completions") {
		if p == "" {
			p = "/v1"
		}
		p += "/chat/completions"
	}
	u.Path = p
	u.RawPath = ""
	return u.String(), nil
}
