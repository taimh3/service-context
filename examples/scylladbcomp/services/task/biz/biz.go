package biz

import (
	"context"

	"github.com/taimaifika/service-context/core"
	"github.com/taimaifika/service-context/examples/scylladbcomp/services/task/entity"
)

type TaskScyllaRepository interface {
	ListTasks(ctx context.Context, filter *entity.Filter, paging *core.Paging) ([]entity.Task, error)
	AddNewTask(ctx context.Context, data *entity.TaskCreateRequest) error
	GetTaskById(ctx context.Context, id int) (*entity.Task, error)
	UpdateTask(ctx context.Context, id int, data *entity.TaskUpdateRequest) error
	DeleteTask(ctx context.Context, id int) error

	// gocqlx
	AddNewPerson(ctx context.Context, person *entity.Person) error
	ListPersons(ctx context.Context, firstName, lastName string) (*[]entity.Person, error)
}

type biz struct {
	taskScyllaRepo TaskScyllaRepository
}

func NewBiz(taskScyllaRepo TaskScyllaRepository) *biz {
	return &biz{
		taskScyllaRepo: taskScyllaRepo,
	}
}
