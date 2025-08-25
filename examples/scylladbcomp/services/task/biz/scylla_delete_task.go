package biz

import (
	"context"

	"go.opentelemetry.io/otel"
)

func (biz *biz) ScyllaDeleteTask(ctx context.Context, id int) error {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "ScyllaDeleteTask")
	defer span.End()

	return biz.taskScyllaRepo.DeleteTask(ctx, id)
}
