package scylla

import (
	"context"
	"strings"
	"time"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
)

func (repo *scyllaRepo) UpdateTask(ctx context.Context, id int, data *entity.TaskUpdateRequest) error {
	_, span := otel.Tracer("scyllaRepo").Start(ctx, "UpdateTask")
	defer span.End()

	var setParts []string
	var queryArgs []interface{}

	if data.Title != nil {
		setParts = append(setParts, "title = ?")
		queryArgs = append(queryArgs, *data.Title)
	}

	if data.Description != nil {
		setParts = append(setParts, "description = ?")
		queryArgs = append(queryArgs, *data.Description)
	}

	if data.Status != nil {
		setParts = append(setParts, "status = ?")
		queryArgs = append(queryArgs, *data.Status)
	}

	if len(setParts) == 0 {
		return errors.New("no fields to update")
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = ?")
	queryArgs = append(queryArgs, time.Now())

	query := `UPDATE tasks SET ` + strings.Join(setParts, ", ") + ` WHERE id = ?`
	queryArgs = append(queryArgs, id)

	q := repo.session.Query(query, queryArgs...).WithContext(ctx)
	defer q.Release() // Release query resources back to the pool

	if err := q.Exec(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
