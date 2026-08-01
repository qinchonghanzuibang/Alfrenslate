package deepl

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/config"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

type Provider struct {
	client              *providers.Client
	key, tier, endpoint string
}

func New(client *providers.Client) providers.Provider {
	tier := strings.ToLower(config.Value("DEEPL_API_TIER"))
	if tier == "" {
		tier = "free"
	}
	base := "https://api-free.deepl.com"
	if tier == "pro" {
		base = "https://api.deepl.com"
	}
	if tier == "custom" {
		base = config.Value("DEEPL_BASE_URL")
	}
	base = strings.TrimSuffix(base, "/")
	endpoint := base + "/v2/translate"
	if strings.HasSuffix(base, "/v2") {
		endpoint = base + "/translate"
	} else if strings.HasSuffix(base, "/v2/translate") {
		endpoint = base
	}
	return &Provider{client: client, key: config.Value("DEEPL_API_KEY"), tier: tier, endpoint: endpoint}
}
func (p *Provider) ID() string            { return "deepl" }
func (p *Provider) Name() string          { return "DeepL" }
func (p *Provider) CacheIdentity() string { return p.endpoint }
func (p *Provider) Configured() error {
	if p.key == "" {
		return errors.New("API key is missing")
	}
	if p.tier != "free" && p.tier != "pro" && p.tier != "custom" {
		return errors.New("API tier must be free, pro, or custom")
	}
	if p.tier == "custom" && config.Value("DEEPL_BASE_URL") == "" {
		return errors.New("custom base URL is missing")
	}
	if u, err := url.Parse(p.endpoint); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("DeepL base URL must be an absolute HTTP(S) URL")
	}
	return nil
}
func (p *Provider) Translate(ctx context.Context, req model.Request) (model.Result, error) {
	target := "ZH-HANS"
	if req.Target == model.TargetEN {
		target = "EN-US"
	}
	body := map[string]any{"text": []string{req.Text}, "target_lang": target, "preserve_formatting": true}
	var out struct {
		Translations []struct {
			Detected string `json:"detected_source_language"`
			Text     string `json:"text"`
		} `json:"translations"`
	}
	err := p.client.JSON(ctx, "POST", p.endpoint, map[string]string{"Authorization": "DeepL-Auth-Key " + p.key}, body, &out)
	if err != nil {
		return model.Result{}, p.tierError(err)
	}
	if len(out.Translations) == 0 || strings.TrimSpace(out.Translations[0].Text) == "" {
		return model.Result{}, errors.New("DeepL returned no translation")
	}
	texts := make([]string, 0, len(out.Translations))
	for _, t := range out.Translations {
		texts = append(texts, t.Text)
	}
	return model.Result{Provider: p.ID(), Text: strings.Join(texts, "\n"), Source: strings.ToUpper(out.Translations[0].Detected), Target: req.Target}, nil
}
func (p *Provider) tierError(err error) error {
	var h *providers.HTTPError
	if errors.As(err, &h) && (h.Status == 403 || h.Status == 404) {
		return errors.New("DeepL authentication or API tier endpoint mismatch; check Free/Pro selection")
	}
	return err
}
