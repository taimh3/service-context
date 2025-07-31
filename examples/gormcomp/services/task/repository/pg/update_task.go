package pg

import (
	"context"

	"github.com/pkg/errors"
	"github.com/taimaifika/service-context/examples/gormcomp/services/task/entity"
)

func (repo *pgRepo) UpdateTask(ctx context.Context, id int, data *entity.Task) error {
	if err := repo.db.WithContext(ctx).
		WithContext(ctx).
		Table(data.TableName()).
		Where("id = ?", id).
		Updates(data).Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}
