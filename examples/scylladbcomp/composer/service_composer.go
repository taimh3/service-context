package composer

import (
	"github.com/taimaifika/service-context/examples/scylladbcomp/common"

	"github.com/gin-gonic/gin"

	taskbiz "github.com/taimaifika/service-context/examples/scylladbcomp/services/task/biz"
	taskscyllarepo "github.com/taimaifika/service-context/examples/scylladbcomp/services/task/repository/scylla"
	taskapi "github.com/taimaifika/service-context/examples/scylladbcomp/services/task/transport/api"

	sctx "github.com/taimaifika/service-context"
)

type TaskService interface {
	// Scylla
	ScyllaListTaskHdl() func(*gin.Context)
	ScyllaGetTaskHdl() func(*gin.Context)
	ScyllaCreateTaskHdl() func(*gin.Context)
	ScyllaUpdateTaskHdl() func(*gin.Context)
	ScyllaDeleteTaskHdl() func(*gin.Context)

	// Person
	ScyllaCreatePersonHdl() func(*gin.Context)
	ScyllaListPersonHdl() func(*gin.Context)
}

func ComposeTaskApiService(serviceCtx sctx.ServiceContext) TaskService {
	scylla := serviceCtx.MustGet(common.KeyCompScylla).(common.ScyllaComponent)
	scyllaSession, _ := scylla.CreateSession()
	scyllaSessionWithGocqlX, _ := scylla.CreateSessionWithGoCqlX()
	scyllaRepo := taskscyllarepo.NewScyllaRepo(
		scylla.GetCluster(),
		scyllaSession,
		scyllaSessionWithGocqlX,
	)

	biz := taskbiz.NewBiz(scyllaRepo)
	serviceApi := taskapi.NewApi(serviceCtx, biz)

	return serviceApi
}
