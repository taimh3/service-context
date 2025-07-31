package pg

import (
	"context"

	"github.com/taimaifika/service-context/core"
	"github.com/taimaifika/service-context/examples/gormcomp/services/task/entity"
	"go.opentelemetry.io/otel"
)

func (repo *pgRepo) ListTasks(ctx context.Context, filter *entity.Filter, paging *core.Paging) ([]entity.Task, error) {
	ctx, span := otel.Tracer(repo.tracerName).Start(ctx, "ListTasks")
	defer span.End()

	var tasks []entity.Task
	query := repo.db.WithContext(ctx).Find(&tasks)
	if filter != nil {
		// TODO: Apply filter logic here
	}
	if paging != nil {
		// TODO: Apply paging logic here
	}

	if err := query.Error; err != nil {
		return nil, err
	}

	return tasks, nil
}
