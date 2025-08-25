package biz

import (
	"context"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"
	"go.opentelemetry.io/otel"
)

func (biz *biz) ScyllaListPersons(ctx context.Context, filter *entity.PersonFilter) ([]entity.Person, error) {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "ScyllaListPersons")
	defer span.End()

	persons, err := biz.taskScyllaRepo.ListPersons(ctx, filter.FirstName, filter.LastName)
	if err != nil {
		return nil, err
	}

	if persons == nil {
		return []entity.Person{}, nil
	}

	return *persons, nil
}
