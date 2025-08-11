package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/taimaifika/service-context/examples/rediscomp/services/cache/entity"
)

type redisRepo struct {
	client *redis.ClusterClient
}

func NewRedisRepo(client *redis.ClusterClient) *redisRepo {
	return &redisRepo{client: client}
}

func (r *redisRepo) Set(ctx context.Context, item *entity.CacheItem) error {
	ctx, span := otel.Tracer("redis-repo").Start(ctx, "Set")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.key", item.Key),
		attribute.String("redis.operation", "set"),
	)

	return r.client.Set(ctx, item.Key, item.Value, item.TTL).Err()
}

func (r *redisRepo) Get(ctx context.Context, key string) (*entity.CacheItem, error) {
	ctx, span := otel.Tracer("redis-repo").Start(ctx, "Get")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.key", key),
		attribute.String("redis.operation", "get"),
	)

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Key not found
		}
		return nil, err
	}

	// Get TTL if exists
	ttl, _ := r.client.TTL(ctx, key).Result()
	var expiresAt *time.Time
	if ttl > 0 {
		exp := time.Now().Add(ttl)
		expiresAt = &exp
	}

	return &entity.CacheItem{
		Key:       key,
		Value:     result,
		TTL:       ttl,
		ExpiresAt: expiresAt,
	}, nil
}

func (r *redisRepo) Delete(ctx context.Context, key string) (int64, error) {
	ctx, span := otel.Tracer("redis-repo").Start(ctx, "Delete")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.key", key),
		attribute.String("redis.operation", "delete"),
	)

	return r.client.Del(ctx, key).Result()
}

func (r *redisRepo) Exists(ctx context.Context, key string) (bool, error) {
	ctx, span := otel.Tracer("redis-repo").Start(ctx, "Exists")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.key", key),
		attribute.String("redis.operation", "exists"),
	)

	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *redisRepo) Keys(ctx context.Context, pattern string) ([]string, error) {
	ctx, span := otel.Tracer("redis-repo").Start(ctx, "Keys")
	defer span.End()

	span.SetAttributes(
		attribute.String("redis.pattern", pattern),
		attribute.String("redis.operation", "keys"),
	)

	return r.client.Keys(ctx, pattern).Result()
}
