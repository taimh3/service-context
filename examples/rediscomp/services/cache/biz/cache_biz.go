package biz

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/taimaifika/service-context/examples/rediscomp/services/cache/entity"
)

type CacheRepo interface {
	Set(ctx context.Context, item *entity.CacheItem) error
	Get(ctx context.Context, key string) (*entity.CacheItem, error)
	Delete(ctx context.Context, key string) (int64, error)
	Exists(ctx context.Context, key string) (bool, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
}

type cacheBiz struct {
	repo CacheRepo
}

func NewCacheBiz(repo CacheRepo) *cacheBiz {
	return &cacheBiz{repo: repo}
}

func (b *cacheBiz) SetCache(ctx context.Context, req *entity.SetCacheRequest) error {
	ctx, span := otel.Tracer("cache-biz").Start(ctx, "SetCache")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.key", req.Key),
		attribute.String("cache.operation", "set"),
	)

	if req.Key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	item := &entity.CacheItem{
		Key:   req.Key,
		Value: req.Value,
		TTL:   req.TTL,
	}

	return b.repo.Set(ctx, item)
}

func (b *cacheBiz) GetCache(ctx context.Context, key string) (*entity.CacheItem, error) {
	ctx, span := otel.Tracer("cache-biz").Start(ctx, "GetCache")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.key", key),
		attribute.String("cache.operation", "get"),
	)

	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	return b.repo.Get(ctx, key)
}

func (b *cacheBiz) DeleteCache(ctx context.Context, key string) (int64, error) {
	ctx, span := otel.Tracer("cache-biz").Start(ctx, "DeleteCache")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.key", key),
		attribute.String("cache.operation", "delete"),
	)

	if key == "" {
		return 0, fmt.Errorf("key cannot be empty")
	}

	return b.repo.Delete(ctx, key)
}

func (b *cacheBiz) ExistsCache(ctx context.Context, key string) (bool, error) {
	ctx, span := otel.Tracer("cache-biz").Start(ctx, "ExistsCache")
	defer span.End()

	span.SetAttributes(
		attribute.String("cache.key", key),
		attribute.String("cache.operation", "exists"),
	)

	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	return b.repo.Exists(ctx, key)
}

func (b *cacheBiz) ListKeys(ctx context.Context, filter *entity.CacheFilter) ([]string, error) {
	ctx, span := otel.Tracer("cache-biz").Start(ctx, "ListKeys")
	defer span.End()

	pattern := "*"
	if filter != nil && filter.Pattern != "" {
		pattern = filter.Pattern
	}

	span.SetAttributes(
		attribute.String("cache.pattern", pattern),
		attribute.String("cache.operation", "keys"),
	)

	return b.repo.Keys(ctx, pattern)
}
