package scylla

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
)

func (repo *scyllaRepo) DeleteTask(ctx context.Context, id int) error {
	_, span := otel.Tracer("scyllaRepo").Start(ctx, "DeleteTask")
	defer span.End()

	query := `DELETE FROM tasks WHERE id = ?`

	q := repo.session.Query(query, id).WithContext(ctx)
	defer q.Release() // Release query resources back to the pool

	if err := q.Exec(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
