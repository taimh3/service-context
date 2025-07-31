package pg

import (
	"context"

	"github.com/pkg/errors"
	"github.com/taimaifika/service-context/examples/gormcomp/services/task/entity"
	"go.opentelemetry.io/otel"
)

func (repo *pgRepo) AddNewTask(ctx context.Context, data *entity.Task) error {
	_, span := otel.Tracer("pgRepo").Start(ctx, "pgRepo.AddNewTask")
	defer span.End()

	if err := repo.db.WithContext(ctx).WithContext(ctx).Create(data).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}
