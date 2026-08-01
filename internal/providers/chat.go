package providers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/prompt"
)

type ChatProvider struct {
	ProviderID, ProviderName, Endpoint, Key, Model string
	Client                                         *Client
	DisableThinking                                bool
}

func (p *ChatProvider) ID() string            { return p.ProviderID }
func (p *ChatProvider) Name() string          { return p.ProviderName }
func (p *ChatProvider) CacheIdentity() string { return p.Endpoint + "|" + p.Model }
func (p *ChatProvider) Configured() error {
	if p.Endpoint == "" {
		return errors.New("base URL is missing")
	}
	if p.Model == "" {
		return errors.New("model is missing")
	}
	if p.ProviderID == "deepseek" && p.Key == "" {
		return errors.New("API key is missing")
	}
	return nil
}

func (p *ChatProvider) Translate(ctx context.Context, req model.Request) (model.Result, error) {
	headers := map[string]string{}
	if p.Key != "" {
		headers["Authorization"] = "Bearer " + p.Key
	}
	body := map[string]any{"model": p.Model, "messages": prompt.Build(req.Text, req.Target), "stream": false, "temperature": 0}
	if p.DisableThinking {
		body["thinking"] = map[string]string{"type": "disabled"}
	}
	var response struct {
		Choices []struct {
			Message struct {
				Content   json.RawMessage `json:"content"`
				Reasoning json.RawMessage `json:"reasoning_content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := p.Client.JSON(ctx, "POST", p.Endpoint, headers, body, &response); err != nil {
		return model.Result{}, err
	}
	if len(response.Choices) == 0 {
		return model.Result{}, errors.New("provider returned no choices")
	}
	content, err := parseContent(response.Choices[0].Message.Content)
	if err != nil || strings.TrimSpace(content) == "" {
		return model.Result{}, errors.New("provider returned empty answer content")
	}
	source := "AUTO"
	if req.Target == model.TargetEN {
		source = "ZH"
	}
	return model.Result{Provider: p.ProviderID, Text: prompt.Clean(content), Source: source, Target: req.Target}, nil
}

func parseContent(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s, nil
	}
	var parts []struct {
		Type    string `json:"type"`
		Text    any    `json:"text"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(raw, &parts); err != nil {
		return "", err
	}
	var b strings.Builder
	for _, part := range parts {
		if part.Type != "" && part.Type != "text" && part.Type != "output_text" {
			continue
		}
		switch v := part.Text.(type) {
		case string:
			b.WriteString(v)
		case map[string]any:
			if x, ok := v["value"].(string); ok {
				b.WriteString(x)
			}
		}
		if part.Text == nil {
			b.WriteString(part.Content)
		}
	}
	return b.String(), nil
}
