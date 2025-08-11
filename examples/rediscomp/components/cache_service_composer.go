package composer

import (
	"github.com/gin-gonic/gin"

	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/examples/rediscomp/common"
	cachebiz "github.com/taimaifika/service-context/examples/rediscomp/services/cache/biz"
	cacherepo "github.com/taimaifika/service-context/examples/rediscomp/services/cache/repository/redis"
	cacheapi "github.com/taimaifika/service-context/examples/rediscomp/services/cache/transport/api"
)

type CacheService interface {
	SetCacheHandler() func(*gin.Context)
	GetCacheHandler() func(*gin.Context)
	DeleteCacheHandler() func(*gin.Context)
	ExistsCacheHandler() func(*gin.Context)
	ListKeysHandler() func(*gin.Context)
}

func ComposeCacheApiService(serviceCtx sctx.ServiceContext) CacheService {
	// load redis client
	redisComp := serviceCtx.MustGet(common.KeyCompRedis).(common.RedisComponent)

	// create repository
	cacheRepo := cacherepo.NewRedisRepo(redisComp.GetRedis())

	// create business logic
	biz := cachebiz.NewCacheBiz(cacheRepo)

	// create API service
	serviceApi := cacheapi.NewCacheApi(serviceCtx, biz)

	return serviceApi
}
