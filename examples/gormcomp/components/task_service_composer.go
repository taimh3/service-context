package composer

import (
	"github.com/gin-gonic/gin"
	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/examples/gormcomp/common"

	taskbiz "github.com/taimaifika/service-context/examples/gormcomp/services/task/biz"
	taskrepo "github.com/taimaifika/service-context/examples/gormcomp/services/task/repository/pg"
	taskapi "github.com/taimaifika/service-context/examples/gormcomp/services/task/transport/api"
)

type TaskService interface {
	ListTaskHdl() func(*gin.Context)
}

func ComposeTaskApiService(serviceCtx sctx.ServiceContext) TaskService {
	// load db
	db := serviceCtx.MustGet(common.KeyCompPostgres).(common.GormComponent)

	taskRepo := taskrepo.NewPgRepo(db.GetDB())

	biz := taskbiz.NewBiz(taskRepo)
	serviceApi := taskapi.NewApi(serviceCtx, biz)

	return serviceApi
}
