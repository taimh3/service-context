package pg

import (
	"context"

	"github.com/pkg/errors"
	"github.com/taimaifika/service-context/core"
	"github.com/taimaifika/service-context/examples/gormcomp/services/task/entity"
	"gorm.io/gorm"
)

func (repo *pgRepo) GetTaskById(ctx context.Context, id int) (*entity.Task, error) {
	var data entity.Task

	if err := repo.db.WithContext(ctx).
		WithContext(ctx).
		Table(data.TableName()).
		Where("id = ?", id).
		First(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, core.ErrRecordNotFound
		}

		return nil, errors.WithStack(err)
	}

	return &data, nil
}
