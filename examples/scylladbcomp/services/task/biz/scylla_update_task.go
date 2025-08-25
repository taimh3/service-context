package biz

import (
	"context"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"go.opentelemetry.io/otel"
)

func (biz *biz) ScyllaUpdateTask(ctx context.Context, id int, data *entity.TaskUpdateRequest) error {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "ScyllaUpdateTask")
	defer span.End()

	return biz.taskScyllaRepo.UpdateTask(ctx, id, data)
}
