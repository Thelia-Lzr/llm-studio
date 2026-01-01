package identra

import (
	"context"
	"log/slog"
	"time"

	"github.com/poly-workshop/llm-studio/internal/domain"
	infraConfig "github.com/poly-workshop/llm-studio/internal/infrastructure/config"

	identra_v1_pb "github.com/poly-workshop/identra/gen/go/identra/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is the identra gRPC client implementation for the application-layer port (OAuthGateway).
type Client struct {
	pb identra_v1_pb.IdentraServiceClient
}

func MustNewClient(cfg infraConfig.Config) (*Client, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.Identra.GRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		panic(err)
	}

	slog.Info("identra grpc connected", "addr", cfg.Identra.GRPCAddr)
	return &Client{pb: identra_v1_pb.NewIdentraServiceClient(conn)}, func() { _ = conn.Close() }
}

func (c *Client) GetAuthorizationURL(ctx context.Context, provider string, redirectURL string) (string, string, error) {
	req := &identra_v1_pb.GetOAuthAuthorizationURLRequest{
		Provider: provider,
	}
	if redirectURL != "" {
		req.RedirectUrl = &redirectURL
	}

	resp, err := c.pb.GetOAuthAuthorizationURL(ctx, req)
	if err != nil {
		return "", "", err
	}
	return resp.GetUrl(), resp.GetState(), nil
}

func (c *Client) LoginByOAuth(ctx context.Context, code string, state string) (domain.TokenPair, error) {
	resp, err := c.pb.LoginByOAuth(ctx, &identra_v1_pb.LoginByOAuthRequest{
		Code:  code,
		State: state,
	})
	if err != nil {
		return domain.TokenPair{}, err
	}

	t := resp.GetToken()
	return domain.TokenPair{
		AccessToken: domain.Token{
			Value:     t.GetAccessToken().GetToken(),
			ExpiresAt: t.GetAccessToken().GetExpiresAt(),
		},
		RefreshToken: domain.Token{
			Value:     t.GetRefreshToken().GetToken(),
			ExpiresAt: t.GetRefreshToken().GetExpiresAt(),
		},
		TokenType: t.GetTokenType(),
	}, nil
}

func (c *Client) GetCurrentUserLoginInfo(ctx context.Context, accessToken string) (domain.LoginInfo, error) {
	resp, err := c.pb.GetCurrentUserLoginInfo(ctx, &identra_v1_pb.GetCurrentUserLoginInfoRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return domain.LoginInfo{}, err
	}

	var oauthConns []domain.OAuthConnection
	for _, oc := range resp.GetOauthConnections() {
		oauthConns = append(oauthConns, domain.OAuthConnection{
			Provider:       oc.GetProvider(),
			ProviderUserID: oc.GetProviderUserId(),
		})
	}

	var githubID *string
	if v := resp.GithubId; v != nil {
		githubID = v
	}

	return domain.LoginInfo{
		UserID:           resp.GetUserId(),
		Email:            resp.GetEmail(),
		GithubID:         githubID,
		PasswordEnabled:  resp.GetPasswordEnabled(),
		OAuthConnections: oauthConns,
	}, nil
}
