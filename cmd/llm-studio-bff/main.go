package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/poly-workshop/go-webmods/app"
	gormclient "github.com/poly-workshop/go-webmods/gormclient"
	redisclient "github.com/poly-workshop/go-webmods/redisclient"

	"github.com/poly-workshop/llm-studio/internal/application/auth"
	"github.com/poly-workshop/llm-studio/internal/application/llmadmin"
	"github.com/poly-workshop/llm-studio/internal/application/llmtoken"
	"github.com/poly-workshop/llm-studio/internal/application/rbac"
	infraAuth "github.com/poly-workshop/llm-studio/internal/infrastructure/auth"
	infraConfig "github.com/poly-workshop/llm-studio/internal/infrastructure/config"
	infraIdentra "github.com/poly-workshop/llm-studio/internal/infrastructure/identra"
	infraLLMGatewayAdmin "github.com/poly-workshop/llm-studio/internal/infrastructure/llmgatewayadmin"
	infraPersistence "github.com/poly-workshop/llm-studio/internal/infrastructure/persistence"
	infraSession "github.com/poly-workshop/llm-studio/internal/infrastructure/session"
	"github.com/poly-workshop/llm-studio/internal/interfaces/httpapi"
)

func main() {
	app.Init("llm-studio-bff")

	cfg := infraConfig.Load()

	db := gormclient.NewDB(gormclient.Config{
		Driver:   cfg.Database.Driver,
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Username: cfg.Database.Username,
		Password: cfg.Database.Password,
		DbName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err := db.AutoMigrate(&infraPersistence.UserModel{}, &infraPersistence.LoginInfoModel{}); err != nil {
		panic(err)
	}
	userRepo := infraPersistence.NewUserRepository(db)

	identraClient, identraClose := infraIdentra.MustNewClient(cfg)
	defer identraClose()

	llmGatewayAdminClient, llmGatewayAdminClose := infraLLMGatewayAdmin.MustNewClient(cfg)
	defer llmGatewayAdminClose()

	rdb := redisclient.NewRDB(redisclient.Config{
		Urls:     cfg.Redis.Urls,
		Password: cfg.Redis.Password,
	})

	sessionStore := infraSession.NewRedisStore(rdb, cfg)

	uidExtractor := infraAuth.NewJWTUIDExtractor()

	authUC := auth.New(identraClient, sessionStore, uidExtractor, userRepo, cfg.Auth.SuperAdminEmails)
	rbacUC := rbac.New(sessionStore, userRepo)
	llmAdminUC := llmadmin.New(llmGatewayAdminClient, sessionStore, userRepo)
	llmTokenUC := llmtoken.New(sessionStore, llmGatewayAdminClient, cfg.LLMGateway.TokenTTLSeconds, nil)

	srv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           httpapi.NewMux(cfg, authUC, rbacUC, llmAdminUC, llmTokenUC),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("BFF listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("BFF server error", "err", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	slog.Info("BFF stopped")
}
