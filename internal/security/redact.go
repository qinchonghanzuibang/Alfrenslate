package security

import (
	"net/url"
	"regexp"
	"strings"
)

var bearer = regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9._~+/=-]+`)
var secretPair = regexp.MustCompile(`(?i)(api[_ -]?key|app[_ -]?secret|subscription[_ -]?key|authorization)(\s*[:=]\s*)[^\s,&]+`)

func Redact(s string, secrets ...string) string {
	s = bearer.ReplaceAllString(s, "Bearer [REDACTED]")
	s = secretPair.ReplaceAllString(s, "$1$2[REDACTED]")
	for _, v := range secrets {
		if len(v) >= 4 {
			s = strings.ReplaceAll(s, v, "[REDACTED]")
		}
	}
	return redactURLKeys(s)
}
func redactURLKeys(s string) string {
	u, e := url.Parse(s)
	if e != nil || u.Host == "" {
		return s
	}
	q := u.Query()
	for k := range q {
		l := strings.ToLower(k)
		if strings.Contains(l, "key") || strings.Contains(l, "secret") || strings.Contains(l, "token") || l == "sign" {
			q.Set(k, "[REDACTED]")
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
