package pg

import (
	"context"

	"github.com/pkg/errors"
	"github.com/taimaifika/service-context/examples/gormcomp/services/task/entity"
)

func (repo *pgRepo) DeleteTask(ctx context.Context, id int) error {
	// Soft delete
	if err := repo.db.WithContext(ctx).
		WithContext(ctx).
		Table(entity.Task{}.TableName()).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": entity.StatusDeleted,
		}).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}
