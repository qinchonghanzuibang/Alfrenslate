package microsoft

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/config"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

type Provider struct {
	client                *providers.Client
	key, region, endpoint string
}

func New(c *providers.Client) (providers.Provider, error) {
	raw := config.Value("MICROSOFT_ENDPOINT")
	if raw == "" {
		raw = "https://api.cognitive.microsofttranslator.com"
	}
	ep, e := normalize(raw)
	if e != nil {
		return nil, e
	}
	return &Provider{c, config.Value("MICROSOFT_API_KEY"), config.Value("MICROSOFT_REGION"), ep}, nil
}
func normalize(raw string) (string, error) {
	u, e := url.Parse(strings.TrimSpace(raw))
	if e != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("Microsoft endpoint must be an absolute URL")
	}
	p := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(p, "/translate") {
		p += "/translate"
	}
	u.Path = p
	q := u.Query()
	q.Set("api-version", "3.0")
	u.RawQuery = q.Encode()
	return u.String(), nil
}
func (p *Provider) ID() string            { return "microsoft" }
func (p *Provider) Name() string          { return "Microsoft Translator" }
func (p *Provider) CacheIdentity() string { return p.endpoint }
func (p *Provider) Configured() error {
	if p.key == "" {
		return errors.New("Subscription Key is missing")
	}
	return nil
}
func traceID() string { b := make([]byte, 16); _, _ = rand.Read(b); return hex.EncodeToString(b) }
func (p *Provider) Translate(ctx context.Context, r model.Request) (model.Result, error) {
	u, _ := url.Parse(p.endpoint)
	q := u.Query()
	if r.Target == model.TargetEN {
		q.Set("to", "en")
	} else {
		q.Set("to", "zh-Hans")
	}
	u.RawQuery = q.Encode()
	h := map[string]string{"Ocp-Apim-Subscription-Key": p.key, "Ocp-Apim-Subscription-Region": p.region, "X-ClientTraceId": traceID()}
	body := []map[string]string{{"Text": r.Text}}
	var out []struct {
		Detected struct {
			Language string `json:"language"`
		} `json:"detectedLanguage"`
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if e := p.client.JSON(ctx, "POST", u.String(), h, body, &out); e != nil {
		return model.Result{}, e
	}
	if len(out) == 0 || len(out[0].Translations) == 0 {
		return model.Result{}, errors.New("Microsoft returned no translation")
	}
	a := make([]string, 0, len(out))
	for _, x := range out {
		if len(x.Translations) > 0 {
			a = append(a, x.Translations[0].Text)
		}
	}
	return model.Result{Provider: p.ID(), Text: strings.Join(a, "\n"), Source: strings.ToUpper(out[0].Detected.Language), Target: r.Target}, nil
}
