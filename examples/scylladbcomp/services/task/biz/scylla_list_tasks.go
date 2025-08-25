package biz

import (
	"context"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/taimaifika/service-context/core"
	"go.opentelemetry.io/otel"
)

func (biz *biz) ScyllaListTasks(ctx context.Context, filter *entity.Filter, paging *core.Paging) ([]entity.Task, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "ScyllaListTasks")
	defer span.End()

	return biz.taskScyllaRepo.ListTasks(ctx, filter, paging)
}
