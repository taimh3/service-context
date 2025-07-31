package api

import (
	"context"

	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/core"
	"github.com/taimaifika/service-context/examples/gormcomp/services/task/entity"
)

type Biz interface {
	ListTasks(ctx context.Context, filter *entity.Filter, paging *core.Paging) ([]entity.Task, error)
}

type api struct {
	serviceCtx sctx.ServiceContext
	biz        Biz
}

func NewApi(serviceCtx sctx.ServiceContext, biz Biz) *api {
	return &api{
		serviceCtx: serviceCtx,
		biz:        biz,
	}
}
