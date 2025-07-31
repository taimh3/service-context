package biz

import (
	"context"

	"github.com/taimaifika/service-context/core"
	"github.com/taimaifika/service-context/examples/gormcomp/services/task/entity"
)

type TaskRepository interface {
	ListTasks(ctx context.Context, filter *entity.Filter, paging *core.Paging) ([]entity.Task, error)
}

type biz struct {
	taskRepo TaskRepository

	tracerName string
}

func NewBiz(taskRepo TaskRepository) *biz {
	return &biz{
		taskRepo:   taskRepo,
		tracerName: "taskBiz",
	}
}
