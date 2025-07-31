package biz

import (
	"context"

	"github.com/taimaifika/service-context/core"
	"github.com/taimaifika/service-context/examples/gormcomp/services/task/entity"
	"go.opentelemetry.io/otel"
)

func (b *biz) ListTasks(ctx context.Context, filter *entity.Filter, paging *core.Paging) ([]entity.Task, error) {
	ctx, span := otel.Tracer(b.tracerName).Start(ctx, "ListTasks")
	defer span.End()

	tasks, err := b.taskRepo.ListTasks(ctx, filter, paging)

	if err != nil {
		return nil, err
	}

	return tasks, nil
}
