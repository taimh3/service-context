package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"

	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/core"
	"github.com/taimaifika/service-context/examples/rediscomp/services/cache/entity"
)

type CacheBiz interface {
	SetCache(ctx context.Context, req *entity.SetCacheRequest) error
	GetCache(ctx context.Context, key string) (*entity.CacheItem, error)
	DeleteCache(ctx context.Context, key string) (int64, error)
	ExistsCache(ctx context.Context, key string) (bool, error)
	ListKeys(ctx context.Context, filter *entity.CacheFilter) ([]string, error)
}

type cacheApi struct {
	serviceCtx sctx.ServiceContext
	biz        CacheBiz
}

func NewCacheApi(serviceCtx sctx.ServiceContext, biz CacheBiz) *cacheApi {
	return &cacheApi{
		serviceCtx: serviceCtx,
		biz:        biz,
	}
}

// SetCacheHandler handles setting cache data
func (a *cacheApi) SetCacheHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("cache-api").Start(c.Request.Context(), "SetCacheHandler")
		defer span.End()

		var req entity.SetCacheRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			core.WriteStandardErrorResponse(c, http.StatusBadRequest, core.GlobalErrorContext.BadRequestError(
				core.ErrBadRequest.Error(),
				err.Error(),
			))
			return
		}

		// Convert TTL from seconds to time.Duration if provided
		if req.TTL > 0 {
			req.TTL = req.TTL * time.Second
		}

		if err := a.biz.SetCache(ctx, &req); err != nil {
			core.WriteStandardErrorResponse(c, http.StatusInternalServerError, core.GlobalErrorContext.InternalServerError(
				core.ErrInternalServerError.Error(),
				err.Error(),
			))
			return
		}

		core.WriteSuccessResponse(c, gin.H{"message": "Cache set successfully"})
	}
}

// GetCacheHandler handles getting cache data
func (a *cacheApi) GetCacheHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("cache-api").Start(c.Request.Context(), "GetCacheHandler")
		defer span.End()

		key := c.Param("key")
		if key == "" {
			core.WriteStandardErrorResponse(c, http.StatusBadRequest, core.GlobalErrorContext.BadRequestError(
				core.ErrBadRequest.Error(),
				"key is required",
			))
			return
		}

		item, err := a.biz.GetCache(ctx, key)
		if err != nil {
			core.WriteStandardErrorResponse(c, http.StatusInternalServerError, core.GlobalErrorContext.InternalServerError(
				core.ErrInternalServerError.Error(),
				err.Error(),
			))
			return
		}

		if item == nil {
			core.WriteStandardErrorResponse(c, http.StatusNotFound, core.GlobalErrorContext.NotFoundError(
				core.ErrNotFound.Error(),
				"key not found",
			))
			return
		}

		core.WriteSuccessResponse(c, item)
	}
}

// DeleteCacheHandler handles deleting cache data
func (a *cacheApi) DeleteCacheHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("cache-api").Start(c.Request.Context(), "DeleteCacheHandler")
		defer span.End()

		key := c.Param("key")
		if key == "" {
			core.WriteStandardErrorResponse(c, http.StatusBadRequest, core.GlobalErrorContext.BadRequestError(
				core.ErrBadRequest.Error(),
				"key is required",
			))
			return
		}

		deleted, err := a.biz.DeleteCache(ctx, key)
		if err != nil {
			core.WriteStandardErrorResponse(c, http.StatusInternalServerError, core.GlobalErrorContext.InternalServerError(
				core.ErrInternalServerError.Error(),
				err.Error(),
			))
			return
		}

		if deleted == 0 {
			core.WriteStandardErrorResponse(c, http.StatusNotFound, core.GlobalErrorContext.NotFoundError(
				core.ErrNotFound.Error(),
				"key not found",
			))
			return
		}

		core.WriteSuccessResponse(c, gin.H{"deleted": deleted})
	}
}

// ExistsCacheHandler handles checking if cache key exists
func (a *cacheApi) ExistsCacheHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("cache-api").Start(c.Request.Context(), "ExistsCacheHandler")
		defer span.End()

		key := c.Param("key")
		if key == "" {
			core.WriteStandardErrorResponse(c, http.StatusBadRequest, core.GlobalErrorContext.BadRequestError(
				core.ErrBadRequest.Error(),
				"key is required",
			))
			return
		}

		exists, err := a.biz.ExistsCache(ctx, key)
		if err != nil {
			core.WriteStandardErrorResponse(c, http.StatusInternalServerError, core.GlobalErrorContext.InternalServerError(
				core.ErrInternalServerError.Error(),
				err.Error(),
			))
			return
		}

		core.WriteSuccessResponse(c, gin.H{"exists": exists})
	}
}

// ListKeysHandler handles listing cache keys
func (a *cacheApi) ListKeysHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		ctx, span := otel.Tracer("cache-api").Start(c.Request.Context(), "ListKeysHandler")
		defer span.End()

		pattern := c.Query("pattern")
		filter := &entity.CacheFilter{
			Pattern: pattern,
		}

		keys, err := a.biz.ListKeys(ctx, filter)
		if err != nil {
			core.WriteStandardErrorResponse(c, http.StatusInternalServerError, core.GlobalErrorContext.InternalServerError(
				core.ErrInternalServerError.Error(),
				err.Error(),
			))
			return
		}

		core.WriteSuccessResponse(c, gin.H{"keys": keys})
	}
}
