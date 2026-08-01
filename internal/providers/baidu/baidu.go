package baidu

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/config"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/model"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

type Provider struct {
	client               *providers.Client
	id, secret, endpoint string
}

func New(c *providers.Client) providers.Provider {
	return &Provider{c, config.Value("BAIDU_APP_ID"), config.Value("BAIDU_APP_SECRET"), "https://fanyi-api.baidu.com/api/trans/vip/translate"}
}
func (p *Provider) ID() string            { return "baidu" }
func (p *Provider) Name() string          { return "Baidu Translate" }
func (p *Provider) CacheIdentity() string { return p.endpoint }
func (p *Provider) Configured() error {
	if p.id == "" {
		return errors.New("App ID is missing")
	}
	if p.secret == "" {
		return errors.New("App Secret is missing")
	}
	return nil
}
func Sign(appID, text, salt, secret string) string {
	s := md5.Sum([]byte(appID + text + salt + secret))
	return hex.EncodeToString(s[:])
}
func randomSalt() string {
	b := make([]byte, 12)
	if _, e := rand.Read(b); e != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b)
}
func (p *Provider) Translate(ctx context.Context, r model.Request) (model.Result, error) {
	s := randomSalt()
	to := "zh"
	if r.Target == model.TargetEN {
		to = "en"
	}
	v := url.Values{"q": {r.Text}, "from": {"auto"}, "to": {to}, "appid": {p.id}, "salt": {s}, "sign": {Sign(p.id, r.Text, s, p.secret)}}
	var out struct {
		From      string `json:"from"`
		ErrorCode string `json:"error_code"`
		Trans     []struct {
			Dst string `json:"dst"`
		} `json:"trans_result"`
	}
	if e := p.client.Form(ctx, p.endpoint, nil, v, &out); e != nil {
		return model.Result{}, e
	}
	if out.ErrorCode != "" {
		return model.Result{}, errors.New(errorMessage(out.ErrorCode))
	}
	if len(out.Trans) == 0 {
		return model.Result{}, errors.New("Baidu returned no translation")
	}
	a := make([]string, 0, len(out.Trans))
	for _, x := range out.Trans {
		a = append(a, x.Dst)
	}
	return model.Result{Provider: p.ID(), Text: strings.Join(a, "\n"), Source: strings.ToUpper(out.From), Target: r.Target}, nil
}
func errorMessage(c string) string {
	m := map[string]string{"52001": "request timed out", "52002": "system error", "52003": "unauthorized user", "54000": "missing required parameter", "54001": "signature verification failed", "54003": "request frequency is limited", "54004": "account balance is insufficient", "54005": "long-query frequency is limited", "58000": "client IP is not authorized", "58001": "unsupported translation direction", "58002": "service is currently closed", "90107": "credential authentication failed"}
	if s := m[c]; s != "" {
		return "Baidu: " + s
	}
	return "Baidu returned error code " + c
}
