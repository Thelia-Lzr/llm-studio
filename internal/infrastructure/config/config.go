package config

import (
	"log/slog"
	"strings"

	"github.com/poly-workshop/go-webmods/app"
)

type Config struct {
	HTTP struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"http"`

	Public struct {
		BaseURL string `mapstructure:"base_url"`
	} `mapstructure:"public"`

	Frontend struct {
		BaseURL string `mapstructure:"base_url"`
	} `mapstructure:"frontend"`

	Identra struct {
		GRPCAddr      string `mapstructure:"grpc_addr"`
		OAuthProvider string `mapstructure:"oauth_provider"`
	} `mapstructure:"identra"`

	LLMGateway struct {
		AdminGRPCAddr string `mapstructure:"admin_grpc_addr"`
		// Single admin token, passed via gRPC metadata/header: x-service-token
		ServiceToken string `mapstructure:"service_token"`
		// Data-plane token TTL for llm-gateway-http JWT (issued by admin service).
		TokenTTLSeconds int64 `mapstructure:"token_ttl_seconds"`
		// Browser-facing auth cookie storing the data-plane JWT (HttpOnly).
		AuthCookieName string `mapstructure:"auth_cookie_name"`
	} `mapstructure:"llm_gateway"`

	Database struct {
		Driver   string `mapstructure:"driver"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`

	Redis struct {
		Urls          []string `mapstructure:"urls"`
		Password      string   `mapstructure:"password"`
		SessionPrefix string   `mapstructure:"session_prefix"`
	} `mapstructure:"redis"`

	Auth struct {
		CookieDomain       string   `mapstructure:"cookie_domain"`
		CookieSecure       bool     `mapstructure:"cookie_secure"`
		SessionCookieName  string   `mapstructure:"session_cookie_name"`
		ReturnToCookieName string   `mapstructure:"return_to_cookie_name"`
		CookiePath         string   `mapstructure:"cookie_path"`
		CookieMaxAgeDays   int      `mapstructure:"cookie_max_age_days"`
		ReturnToMaxAgeMins int      `mapstructure:"return_to_max_age_mins"`
		SuperAdminEmails   []string `mapstructure:"super_admin_emails"`
	} `mapstructure:"auth"`
}

func Load() Config {
	v := app.Config()

	// Defaults (can be overridden by configs/*.toml or ENV with "__").
	v.SetDefault("http.addr", ":8080")
	v.SetDefault("public.base_url", "http://localhost:8080")
	v.SetDefault("frontend.base_url", "http://localhost:3000")
	v.SetDefault("identra.grpc_addr", "localhost:50051")
	v.SetDefault("identra.oauth_provider", "github")
	v.SetDefault("llm_gateway.admin_grpc_addr", "localhost:50052")
	// Match llm-gateway default config for local dev.
	v.SetDefault("llm_gateway.service_token", "changeme-admin-token")
	// Data-plane JWT TTL for browser chat usage (1h).
	v.SetDefault("llm_gateway.token_ttl_seconds", int64(3600))
	// Cookie used by the browser/proxy to authenticate to llm-gateway-http.
	v.SetDefault("llm_gateway.auth_cookie_name", "llmgw_access_token")

	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.host", "")
	v.SetDefault("database.port", 0)
	v.SetDefault("database.username", "")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "data/llm-studio.db")
	v.SetDefault("database.sslmode", "disable")

	v.SetDefault("redis.urls", []string{"localhost:6379"})
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.session_prefix", "llmstudio:sess:")

	v.SetDefault("auth.cookie_secure", false)
	v.SetDefault("auth.cookie_domain", "")
	v.SetDefault("auth.cookie_path", "/")
	v.SetDefault("auth.session_cookie_name", "llmstudio_session")
	v.SetDefault("auth.return_to_cookie_name", "llmstudio_return_to")
	v.SetDefault("auth.cookie_max_age_days", 7)
	v.SetDefault("auth.return_to_max_age_mins", 10)
	v.SetDefault("auth.super_admin_emails", []string{})

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	cfg.Public.BaseURL = strings.TrimRight(cfg.Public.BaseURL, "/")
	cfg.Frontend.BaseURL = strings.TrimRight(cfg.Frontend.BaseURL, "/")
	cfg.Identra.OAuthProvider = strings.TrimSpace(cfg.Identra.OAuthProvider)

	slog.Info("config loaded",
		"http.addr", cfg.HTTP.Addr,
		"public.base_url", cfg.Public.BaseURL,
		"frontend.base_url", cfg.Frontend.BaseURL,
		"identra.grpc_addr", cfg.Identra.GRPCAddr,
		"identra.oauth_provider", cfg.Identra.OAuthProvider,
		"llm_gateway.admin_grpc_addr", cfg.LLMGateway.AdminGRPCAddr,
		"llm_gateway.service_token_present", cfg.LLMGateway.ServiceToken != "",
		"llm_gateway.token_ttl_seconds", cfg.LLMGateway.TokenTTLSeconds,
		"llm_gateway.auth_cookie_name", cfg.LLMGateway.AuthCookieName,
		"database.driver", cfg.Database.Driver,
		"database.name", cfg.Database.Name,
		"redis.urls", cfg.Redis.Urls,
		"redis.session_prefix", cfg.Redis.SessionPrefix,
		"auth.super_admin_emails", cfg.Auth.SuperAdminEmails,
	)

	return cfg
}
