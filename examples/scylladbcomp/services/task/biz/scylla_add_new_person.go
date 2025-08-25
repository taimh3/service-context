package biz

import (
	"context"

	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"
	"go.opentelemetry.io/otel"
)

func (biz *biz) ScyllaAddNewPerson(ctx context.Context, data *entity.PersonCreateRequest) error {
	ctx, span := otel.Tracer("auth-service").Start(ctx, "ScyllaAddNewPerson")
	defer span.End()

	person := &entity.Person{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
	}

	return biz.taskScyllaRepo.AddNewPerson(ctx, person)
}
