package biz

import (
	"context"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"go.opentelemetry.io/otel"
)

func (biz *biz) ScyllaGetTaskById(ctx context.Context, id int) (*entity.Task, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "ScyllaGetTaskById")
	defer span.End()

	return biz.taskScyllaRepo.GetTaskById(ctx, id)
}
