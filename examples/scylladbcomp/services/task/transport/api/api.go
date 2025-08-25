package api

import (
	"context"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	sctx "github.com/taimaifika/service-context"
	"github.com/taimaifika/service-context/core"
)

type Biz interface {
	// Scylla
	ScyllaListTasks(ctx context.Context, filter *entity.Filter, paging *core.Paging) ([]entity.Task, error)
	ScyllaAddNewTask(ctx context.Context, data *entity.TaskCreateRequest) error
	ScyllaGetTaskById(ctx context.Context, id int) (*entity.Task, error)
	ScyllaUpdateTask(ctx context.Context, id int, data *entity.TaskUpdateRequest) error
	ScyllaDeleteTask(ctx context.Context, id int) error

	// Person
	ScyllaAddNewPerson(ctx context.Context, data *entity.PersonCreateRequest) error
	ScyllaListPersons(ctx context.Context, filter *entity.PersonFilter) ([]entity.Person, error)
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
