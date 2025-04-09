package auth

import (
	"context"
	"encoding/json"
	"time"

	auth_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/auth"
	"github.com/redis/go-redis/v9"
)

type SessionRepository interface {
	Create(ctx context.Context, session *auth_model.Session) error
	Get(ctx context.Context, sessionID string) (*auth_model.Session, error)
	Delete(ctx context.Context, sessionID string) error
}

type redisRepository struct {
	client  *redis.Client
	timeout time.Duration
}

func NewSessionRepository(client *redis.Client, timeout time.Duration) SessionRepository {
	return &redisRepository{
		client:  client,
		timeout: timeout,
	}
}

func (r *redisRepository) Create(ctx context.Context, session *auth_model.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	expiration := time.Until(session.ExpiresAt)
	return r.client.Set(ctx, session.ID, data, expiration).Err()
}

func (r *redisRepository) Get(ctx context.Context, sessionID string) (*auth_model.Session, error) {
	data, err := r.client.Get(ctx, sessionID).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var session auth_model.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *redisRepository) Delete(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, sessionID).Err()
}
