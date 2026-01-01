package llmgatewayadmin

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/poly-workshop/llm-studio/internal/domain"
	infraConfig "github.com/poly-workshop/llm-studio/internal/infrastructure/config"

	adminv1 "github.com/poly-workshop/llm-gateway/gen/go/llmgateway/admin/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Client is the llm-gateway-admin-grpc client.
// It is used as an infrastructure adapter and should be wrapped by an application port.
//
// Note: llm-gateway admin API is intended to be called from trusted control-plane services.
// This BFF currently connects with insecure transport for local dev.
// In production, you should use TLS/mTLS.
type Client struct {
	pb adminv1.LLMGatewayAdminServiceClient
}

const mdAdminServiceToken = "x-service-token"

func serviceTokenUnaryInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, mdAdminServiceToken, token)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func serviceTokenStreamInterceptor(token string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, mdAdminServiceToken, token)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func MustNewClient(cfg infraConfig.Config) (*Client, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := cfg.LLMGateway.AdminGRPCAddr
	serviceToken := cfg.LLMGateway.ServiceToken
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(serviceTokenUnaryInterceptor(serviceToken)),
		grpc.WithStreamInterceptor(serviceTokenStreamInterceptor(serviceToken)),
		grpc.WithBlock(),
	)
	if err != nil {
		panic(err)
	}

	slog.Info("llm-gateway admin grpc connected",
		"addr", addr,
		"service_token_present", serviceToken != "",
	)
	return &Client{pb: adminv1.NewLLMGatewayAdminServiceClient(conn)}, func() { _ = conn.Close() }
}

func providerTypeToProto(p domain.LLMProviderType) adminv1.ProviderType {
	switch p {
	case domain.LLMProviderDashscope:
		return adminv1.ProviderType_PROVIDER_TYPE_DASHSCOPE
	case domain.LLMProviderOpenRouter:
		return adminv1.ProviderType_PROVIDER_TYPE_OPENROUTER
	default:
		return adminv1.ProviderType_PROVIDER_TYPE_UNSPECIFIED
	}
}

func providerTypeFromProto(p adminv1.ProviderType) domain.LLMProviderType {
	switch p {
	case adminv1.ProviderType_PROVIDER_TYPE_DASHSCOPE:
		return domain.LLMProviderDashscope
	case adminv1.ProviderType_PROVIDER_TYPE_OPENROUTER:
		return domain.LLMProviderOpenRouter
	default:
		return domain.LLMProviderUnspecified
	}
}

func capabilityToProto(c domain.LLMModelCapability) adminv1.ModelCapability {
	switch c {
	case domain.LLMCapabilityText:
		return adminv1.ModelCapability_MODEL_CAPABILITY_TEXT
	case domain.LLMCapabilityImages:
		return adminv1.ModelCapability_MODEL_CAPABILITY_IMAGES
	case domain.LLMCapabilityAudio:
		return adminv1.ModelCapability_MODEL_CAPABILITY_AUDIO
	case domain.LLMCapabilityVideo:
		return adminv1.ModelCapability_MODEL_CAPABILITY_VIDEO
	case domain.LLMCapabilityTools:
		return adminv1.ModelCapability_MODEL_CAPABILITY_TOOLS
	case domain.LLMCapabilityPromptCache:
		return adminv1.ModelCapability_MODEL_CAPABILITY_PROMPT_CACHE
	case domain.LLMCapabilityStreaming:
		return adminv1.ModelCapability_MODEL_CAPABILITY_STREAMING
	case domain.LLMCapabilityReasoning:
		return adminv1.ModelCapability_MODEL_CAPABILITY_REASONING
	default:
		return adminv1.ModelCapability_MODEL_CAPABILITY_UNSPECIFIED
	}
}

func capabilityFromProto(c adminv1.ModelCapability) domain.LLMModelCapability {
	switch c {
	case adminv1.ModelCapability_MODEL_CAPABILITY_TEXT:
		return domain.LLMCapabilityText
	case adminv1.ModelCapability_MODEL_CAPABILITY_IMAGES:
		return domain.LLMCapabilityImages
	case adminv1.ModelCapability_MODEL_CAPABILITY_AUDIO:
		return domain.LLMCapabilityAudio
	case adminv1.ModelCapability_MODEL_CAPABILITY_VIDEO:
		return domain.LLMCapabilityVideo
	case adminv1.ModelCapability_MODEL_CAPABILITY_TOOLS:
		return domain.LLMCapabilityTools
	case adminv1.ModelCapability_MODEL_CAPABILITY_PROMPT_CACHE:
		return domain.LLMCapabilityPromptCache
	case adminv1.ModelCapability_MODEL_CAPABILITY_STREAMING:
		return domain.LLMCapabilityStreaming
	case adminv1.ModelCapability_MODEL_CAPABILITY_REASONING:
		return domain.LLMCapabilityReasoning
	default:
		return ""
	}
}

func (c *Client) UpsertProviderConfig(ctx context.Context, cfg domain.LLMProviderConfig) error {
	_, err := c.pb.UpsertProviderConfig(ctx, &adminv1.UpsertProviderConfigRequest{
		Config: &adminv1.ProviderConfig{
			Provider:       providerTypeToProto(cfg.Provider),
			BaseUrl:        cfg.BaseURL,
			ApiKey:         cfg.APIKey,
			TimeoutSeconds: cfg.TimeoutSeconds,
		},
	})
	return err
}

func (c *Client) DeleteProviderConfig(ctx context.Context, provider domain.LLMProviderType) error {
	_, err := c.pb.DeleteProviderConfig(ctx, &adminv1.DeleteProviderConfigRequest{
		Provider: providerTypeToProto(provider),
	})
	return err
}

func (c *Client) ListProviderConfigs(ctx context.Context) ([]domain.LLMProviderConfigView, error) {
	resp, err := c.pb.ListProviderConfigs(ctx, &adminv1.ListProviderConfigsRequest{})
	if err != nil {
		return nil, err
	}

	out := make([]domain.LLMProviderConfigView, 0, len(resp.GetConfigs()))
	for _, pc := range resp.GetConfigs() {
		out = append(out, domain.LLMProviderConfigView{
			Provider:       providerTypeFromProto(pc.GetProvider()),
			BaseURL:        pc.GetBaseUrl(),
			TimeoutSeconds: pc.GetTimeoutSeconds(),
			APIKeyPresent:  pc.GetApiKeyPresent(),
		})
	}
	// Stable output.
	sort.Slice(out, func(i, j int) bool { return out[i].Provider < out[j].Provider })
	return out, nil
}

func (c *Client) UpsertModel(ctx context.Context, cfg domain.LLMModelConfig) (string, error) {
	capabilities := make([]adminv1.ModelCapability, 0, len(cfg.Capabilities))
	for _, cap := range cfg.Capabilities {
		pc := capabilityToProto(cap)
		if pc == adminv1.ModelCapability_MODEL_CAPABILITY_UNSPECIFIED {
			continue
		}
		capabilities = append(capabilities, pc)
	}

	resp, err := c.pb.UpsertModel(ctx, &adminv1.UpsertModelRequest{
		Model: &adminv1.ModelConfig{
			Provider:      providerTypeToProto(cfg.Provider),
			Capabilities:  capabilities,
			UpstreamModel: cfg.UpstreamModel,
		},
	})
	if err != nil {
		return "", err
	}
	return resp.GetId(), nil
}

func (c *Client) DeleteModel(ctx context.Context, id string) error {
	_, err := c.pb.DeleteModel(ctx, &adminv1.DeleteModelRequest{Id: id})
	return err
}

func (c *Client) ListModels(ctx context.Context) ([]domain.LLMModelSpec, error) {
	resp, err := c.pb.ListModels(ctx, &adminv1.ListModelsRequest{})
	if err != nil {
		return nil, err
	}

	out := make([]domain.LLMModelSpec, 0, len(resp.GetModels()))
	for _, m := range resp.GetModels() {
		caps := make([]domain.LLMModelCapability, 0, len(m.GetCapabilities()))
		for _, cap := range m.GetCapabilities() {
			if c := capabilityFromProto(cap); c != "" {
				caps = append(caps, c)
			}
		}
		out = append(out, domain.LLMModelSpec{
			ID:            m.GetId(),
			Provider:      providerTypeFromProto(m.GetProvider()),
			Capabilities:  caps,
			UpstreamModel: m.GetUpstreamModel(),
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (c *Client) IssueToken(ctx context.Context, subject string, ttlSeconds int64, allowedModelIDs []string) (string, int64, error) {
	resp, err := c.pb.IssueToken(ctx, &adminv1.IssueTokenRequest{
		Subject:         subject,
		TtlSeconds:      ttlSeconds,
		AllowedModelIds: allowedModelIDs,
	})
	if err != nil {
		return "", 0, err
	}
	return resp.GetAccessToken(), resp.GetExpiresAtUnix(), nil
}
