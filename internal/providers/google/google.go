package google

import (
	"context"
	"errors"
	"html"
	"net/url"
	"strings"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/config"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

type Provider struct {
	client        *providers.Client
	key, endpoint string
}

func New(c *providers.Client) providers.Provider {
	return &Provider{c, config.Value("GOOGLE_API_KEY"), "https://translation.googleapis.com/language/translate/v2"}
}
func (p *Provider) ID() string            { return "google" }
func (p *Provider) Name() string          { return "Google Cloud Translation" }
func (p *Provider) CacheIdentity() string { return p.endpoint }
func (p *Provider) Configured() error {
	if p.key == "" {
		return errors.New("API key is missing")
	}
	return nil
}
func (p *Provider) Translate(ctx context.Context, r model.Request) (model.Result, error) {
	u, e := url.Parse(p.endpoint)
	if e != nil {
		return model.Result{}, e
	}
	q := u.Query()
	q.Set("key", p.key)
	u.RawQuery = q.Encode()
	target := "zh-CN"
	if r.Target == model.TargetEN {
		target = "en"
	}
	body := map[string]any{"q": []string{r.Text}, "target": target, "format": "text"}
	var out struct {
		Data struct {
			Translations []struct {
				Detected string `json:"detectedSourceLanguage"`
				Text     string `json:"translatedText"`
			} `json:"translations"`
		} `json:"data"`
	}
	if e = p.client.JSON(ctx, "POST", u.String(), nil, body, &out); e != nil {
		return model.Result{}, googleError(e)
	}
	if len(out.Data.Translations) == 0 {
		return model.Result{}, errors.New("Google returned no translation")
	}
	a := make([]string, 0, len(out.Data.Translations))
	for _, x := range out.Data.Translations {
		a = append(a, html.UnescapeString(x.Text))
	}
	return model.Result{Provider: p.ID(), Text: strings.Join(a, "\n"), Source: strings.ToUpper(out.Data.Translations[0].Detected), Target: r.Target}, nil
}
func googleError(e error) error {
	var h *providers.HTTPError
	if errors.As(e, &h) {
		switch h.Status {
		case 400:
			return errors.New("Google request failed; verify the API is enabled and the key is valid")
		case 403:
			return errors.New("Google access denied; check API enablement, key restrictions, and quota")
		case 429:
			return errors.New("Google translation quota exceeded")
		}
	}
	return e
}
