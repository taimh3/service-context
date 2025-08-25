package scylla

import (
	"context"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
)

func (repo *scyllaRepo) GetTaskById(ctx context.Context, id int) (*entity.Task, error) {
	_, span := otel.Tracer("scyllaRepo").Start(ctx, "GetTaskById")
	defer span.End()

	var task entity.Task
	query := `SELECT id, created_at, updated_at, title, description, status FROM tasks WHERE id = ?`

	q := repo.session.Query(query, id).WithContext(ctx)
	defer q.Release() // Release query resources back to the pool

	if err := q.Scan(
		&task.Id,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.Title,
		&task.Description,
		&task.Status,
	); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.New("task not found")
		}
		return nil, errors.WithStack(err)
	}

	return &task, nil
}
