package biz

import (
	"context"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"
	"go.opentelemetry.io/otel"
)

func (biz *biz) ScyllaAddNewTask(ctx context.Context, data *entity.TaskCreateRequest) error {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "ScyllaAddNewTask")
	defer span.End()

	return biz.taskScyllaRepo.AddNewTask(ctx, data)
}
