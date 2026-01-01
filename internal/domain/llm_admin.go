package domain

import (
	"strings"
)

type LLMProviderType string

const (
	LLMProviderUnspecified LLMProviderType = ""
	LLMProviderDashscope   LLMProviderType = "dashscope"
	LLMProviderOpenRouter  LLMProviderType = "openrouter"
)

func ParseLLMProviderType(s string) (LLMProviderType, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	// Accept proto enum names like "PROVIDER_TYPE_DASHSCOPE".
	s = strings.TrimPrefix(s, "provider_type_")
	// Accept common prefixes.
	s = strings.TrimPrefix(s, "provider-type-")
	s = strings.TrimPrefix(s, "provider_")
	s = strings.TrimPrefix(s, "provider-")

	// Normalize separators.
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")

	switch s {
	case "dashscope":
		return LLMProviderDashscope, true
	case "openrouter":
		return LLMProviderOpenRouter, true
	default:
		return LLMProviderUnspecified, false
	}
}

type LLMProviderConfig struct {
	Provider       LLMProviderType
	BaseURL        string
	APIKey         string
	TimeoutSeconds int64
}

type LLMProviderConfigView struct {
	Provider       LLMProviderType `json:"provider"`
	BaseURL        string          `json:"base_url"`
	TimeoutSeconds int64           `json:"timeout_seconds"`
	APIKeyPresent  bool            `json:"api_key_present"`
}

type LLMModelCapability string

const (
	LLMCapabilityText        LLMModelCapability = "text"
	LLMCapabilityImages      LLMModelCapability = "images"
	LLMCapabilityAudio       LLMModelCapability = "audio"
	LLMCapabilityVideo       LLMModelCapability = "video"
	LLMCapabilityTools       LLMModelCapability = "tools"
	LLMCapabilityPromptCache LLMModelCapability = "prompt_cache"
	LLMCapabilityStreaming   LLMModelCapability = "streaming"
	LLMCapabilityReasoning   LLMModelCapability = "reasoning"
)

func ParseLLMModelCapability(s string) (LLMModelCapability, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "model_capability_")
	s = strings.TrimPrefix(s, "model-capability-")
	s = strings.TrimPrefix(s, "capability_")
	s = strings.TrimPrefix(s, "capability-")

	s = strings.ReplaceAll(s, "-", "_")

	switch s {
	case "text":
		return LLMCapabilityText, true
	case "images", "image":
		return LLMCapabilityImages, true
	case "audio":
		return LLMCapabilityAudio, true
	case "video":
		return LLMCapabilityVideo, true
	case "tools", "tool":
		return LLMCapabilityTools, true
	case "prompt_cache", "promptcache":
		return LLMCapabilityPromptCache, true
	case "streaming", "stream":
		return LLMCapabilityStreaming, true
	case "reasoning", "reason":
		return LLMCapabilityReasoning, true
	default:
		return "", false
	}
}

type LLMModelConfig struct {
	Provider      LLMProviderType
	Capabilities  []LLMModelCapability
	UpstreamModel string
}

type LLMModelSpec struct {
	ID            string               `json:"id"`
	Provider      LLMProviderType      `json:"provider"`
	Capabilities  []LLMModelCapability `json:"capabilities"`
	UpstreamModel string               `json:"upstream_model"`
}
