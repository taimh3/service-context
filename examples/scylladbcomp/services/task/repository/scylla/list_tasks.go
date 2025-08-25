package scylla

import (
	"context"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/pkg/errors"
	"github.com/taimaifika/service-context/core"
	"go.opentelemetry.io/otel"
)

func (repo *scyllaRepo) ListTasks(ctx context.Context, filter *entity.Filter, paging *core.Paging) ([]entity.Task, error) {
	_, span := otel.Tracer("scyllaRepo").Start(ctx, "ListTasks")
	defer span.End()

	var tasks []entity.Task
	query := `SELECT id, created_at, updated_at, title, description, status FROM tasks`

	var queryArgs []interface{}

	// // Add filter conditions if provided
	// if filter != nil && filter.Status != nil {
	// 	query += ` WHERE status = ?`
	// 	queryArgs = append(queryArgs, *filter.Status)
	// }

	// // Add pagination
	// if paging != nil && paging.Limit > 0 {
	// 	query += ` LIMIT ?`
	// 	queryArgs = append(queryArgs, paging.Limit)
	// }

	q := repo.session.Query(query, queryArgs...).WithContext(ctx)
	defer q.Release() // Release query resources back to the pool

	iter := q.Iter()
	defer iter.Close()

	var task entity.Task
	for iter.Scan(&task.Id, &task.CreatedAt, &task.UpdatedAt, &task.Title, &task.Description, &task.Status) {
		tasks = append(tasks, task)
		task = entity.Task{} // Reset for next iteration
	}

	if err := iter.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	return tasks, nil
}
