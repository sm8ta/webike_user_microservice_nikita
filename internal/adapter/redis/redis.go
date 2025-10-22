package redis

import (
	"context"
	"time"
	"webike_services/webike_User-microservice_Nikita/internal/core/ports"

	"github.com/redis/go-redis/v9"
)

type RedisAdapter struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisAdapter(client *redis.Client) ports.CachePort {
	return &RedisAdapter{
		client: client,
		ctx:    context.Background(),
	}
}

func (r *RedisAdapter) Get(key string) ([]byte, error) {
	result, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return []byte(result), nil
}

func (r *RedisAdapter) Set(key string, value []byte, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

func (r *RedisAdapter) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

var _ ports.CachePort = (*RedisAdapter)(nil)
