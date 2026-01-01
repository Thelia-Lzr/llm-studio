package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/poly-workshop/llm-studio/internal/domain"
	infraConfig "github.com/poly-workshop/llm-studio/internal/infrastructure/config"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	rdb           redis.UniversalClient
	sessionPrefix string
}

func NewRedisStore(rdb redis.UniversalClient, cfg infraConfig.Config) *RedisStore {
	return &RedisStore{
		rdb:           rdb,
		sessionPrefix: cfg.Redis.SessionPrefix,
	}
}

func (s *RedisStore) Save(ctx context.Context, sessionID string, session domain.Session, ttl time.Duration) error {
	b, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.sessionPrefix+sessionID, b, ttl).Err()
}

func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.rdb.Del(ctx, s.sessionPrefix+sessionID).Err()
}

func (s *RedisStore) Get(ctx context.Context, sessionID string) (domain.Session, error) {
	b, err := s.rdb.Get(ctx, s.sessionPrefix+sessionID).Bytes()
	if err != nil {
		return domain.Session{}, err
	}
	var sess domain.Session
	if err := json.Unmarshal(b, &sess); err != nil {
		return domain.Session{}, err
	}
	return sess, nil
}
