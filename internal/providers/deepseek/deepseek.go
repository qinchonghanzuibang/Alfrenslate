package deepseek

import (
	"net/url"
	"strings"

	"github.com/qinchonghanzuibang/Alfrenslate/internal/config"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

func New(client *providers.Client) (providers.Provider, error) {
	base := config.Value("DEEPSEEK_BASE_URL")
	if base == "" {
		base = "https://api.deepseek.com"
	}
	endpoint, err := normalize(base)
	if err != nil {
		return nil, err
	}
	model := config.Value("DEEPSEEK_MODEL")
	if model == "" {
		model = "deepseek-v4-flash"
	}
	return &providers.ChatProvider{ProviderID: "deepseek", ProviderName: "DeepSeek", Endpoint: endpoint, Key: config.Value("DEEPSEEK_API_KEY"), Model: model, Client: client, DisableThinking: true}, nil
}

func normalize(base string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err == nil && u.Scheme != "" && u.Host != "" && strings.Trim(u.Path, "/") == "" {
		u.Path = "/chat/completions"
		return u.String(), nil
	}
	return providers.NormalizeChatURL(base)
}
