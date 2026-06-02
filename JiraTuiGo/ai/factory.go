package ai

import (
	"fmt"

	"jiratui/config"
)

// NewAiClient builds the appropriate AiClient from the given AiConfig.
// unprotect is called to decrypt ApiKeyProtected before use.
func NewAiClient(cfg *config.AiConfig, unprotect func(string) string) (AiClient, error) {
	if cfg.Adapter == "" {
		return nil, fmt.Errorf("AI adapter not configured")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("AI model not configured")
	}

	apiKey := unprotect(cfg.ApiKeyProtected)

	switch cfg.Adapter {
	case "anthropic":
		return NewAnthropicClient(apiKey, cfg.Model), nil
	default:
		return NewOpenAIClient(cfg.BaseURL, apiKey, cfg.Model), nil
	}
}
