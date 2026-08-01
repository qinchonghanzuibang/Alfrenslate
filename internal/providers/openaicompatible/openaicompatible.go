package openaicompatible

import (
	"github.com/qinchonghanzuibang/Alfrenslate/internal/config"
	"github.com/qinchonghanzuibang/Alfrenslate/internal/providers"
)

func New(client *providers.Client) (providers.Provider, error) {
	endpoint, err := providers.NormalizeChatURL(config.Value("OPENAI_COMPATIBLE_BASE_URL"))
	if err != nil {
		return nil, err
	}
	return &providers.ChatProvider{ProviderID: "openai-compatible", ProviderName: "OpenAI-Compatible", Endpoint: endpoint, Key: config.Value("OPENAI_COMPATIBLE_API_KEY"), Model: config.Value("OPENAI_COMPATIBLE_MODEL"), Client: client}, nil
}
