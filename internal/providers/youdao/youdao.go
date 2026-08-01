package youdao

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/config"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

type Provider struct {
	client                        *providers.Client
	key, secret, domain, endpoint string
}

func New(client *providers.Client) providers.Provider {
	return &Provider{client: client, key: config.Value("YOUDAO_APP_KEY"), secret: config.Value("YOUDAO_APP_SECRET"), domain: domain(), endpoint: "https://openapi.youdao.com/api"}
}
func domain() string {
	v := config.Value("YOUDAO_DOMAIN")
	if v == "" {
		return "general"
	}
	return v
}
func (p *Provider) ID() string            { return "youdao" }
func (p *Provider) Name() string          { return "Youdao" }
func (p *Provider) CacheIdentity() string { return p.endpoint + "|" + p.domain }
func (p *Provider) Configured() error {
	if p.key == "" {
		return errors.New("App Key is missing")
	}
	if p.secret == "" {
		return errors.New("App Secret is missing")
	}
	switch p.domain {
	case "general", "computers", "medicine", "finance", "game":
	default:
		return errors.New("unsupported translation domain")
	}
	return nil
}
func TruncateInput(q string) string {
	r := []rune(q)
	if len(r) <= 20 {
		return q
	}
	return string(r[:10]) + strconv.Itoa(len(r)) + string(r[len(r)-10:])
}
func Sign(appKey, q, salt, curtime, secret string) string {
	sum := sha256.Sum256([]byte(appKey + TruncateInput(q) + salt + curtime + secret))
	return hex.EncodeToString(sum[:])
}
func salt() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b)
}
func (p *Provider) Translate(ctx context.Context, req model.Request) (model.Result, error) {
	if !utf8.ValidString(req.Text) {
		return model.Result{}, errors.New("input is not valid UTF-8")
	}
	s := salt()
	now := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	to := "zh-CHS"
	if req.Target == model.TargetEN {
		to = "en"
	}
	v := url.Values{"q": {req.Text}, "from": {"auto"}, "to": {to}, "appKey": {p.key}, "salt": {s}, "sign": {Sign(p.key, req.Text, s, now, p.secret)}, "signType": {"v3"}, "curtime": {now}, "domain": {p.domain}}
	var out struct {
		ErrorCode   string   `json:"errorCode"`
		Translation []string `json:"translation"`
		Lang        string   `json:"l"`
	}
	if err := p.client.Form(ctx, p.endpoint, nil, v, &out); err != nil {
		return model.Result{}, err
	}
	if out.ErrorCode != "0" {
		return model.Result{}, errors.New(errorMessage(out.ErrorCode))
	}
	if len(out.Translation) == 0 {
		return model.Result{}, errors.New("Youdao returned no translation")
	}
	source := "AUTO"
	if i := strings.Index(out.Lang, "2"); i > 0 {
		source = strings.ToUpper(out.Lang[:i])
	}
	return model.Result{Provider: p.ID(), Text: strings.Join(out.Translation, "\n"), Source: source, Target: req.Target}, nil
}
func errorMessage(code string) string {
	m := map[string]string{"101": "invalid credentials", "102": "unsupported language", "103": "input is too long", "108": "invalid application ID", "110": "no related service instance", "111": "account has insufficient balance", "202": "signature verification failed", "401": "account is not valid", "411": "access frequency is limited"}
	if s := m[code]; s != "" {
		return "Youdao: " + s
	}
	return "Youdao returned error code " + code
}
