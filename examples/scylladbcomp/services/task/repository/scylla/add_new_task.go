package scylla

import (
	"context"
	"time"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
)

func (repo *scyllaRepo) AddNewTask(ctx context.Context, data *entity.TaskCreateRequest) error {
	ctx, span := otel.Tracer("scyllaRepo").Start(ctx, "AddNewTask")
	defer span.End()

	query := `INSERT INTO tasks (id, created_at, updated_at, title, description, status) VALUES (?, ?, ?, ?, ?, ?)`
	// Ensure the task ID is set
	if data.Id == 0 {
		data.Id = entity.GenerateTaskID() // Assuming GenerateTaskID is a function that generates a new task ID
	}
	// Ensure the created_at and updated_at timestamps are set
	now := time.Now()
	if data.CreatedAt == nil || data.CreatedAt.IsZero() {
		data.CreatedAt = &now
	}
	if data.UpdatedAt == nil || data.UpdatedAt.IsZero() {
		data.UpdatedAt = &now
	}

	q := repo.session.Query(query,
		data.Id,
		data.CreatedAt,
		data.UpdatedAt,
		data.Title,
		data.Description,
		data.Status,
	).WithContext(ctx)
	defer q.Release() // Release query resources back to the pool

	if err := q.Exec(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
