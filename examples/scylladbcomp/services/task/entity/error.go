package entity

import "errors"

var (
	ErrTaskDeleted      = errors.New("task has been deleted")
	ErrTaskNotFound     = errors.New("task not found")
	ErrCannotCreateTask = errors.New("cannot create task")
	ErrCannotUpdateTask = errors.New("cannot update task")
	ErrCannotDeleteTask = errors.New("cannot update task")
	ErrCannotListTask   = errors.New("cannot list tasks")
	ErrCannotGetTask    = errors.New("cannot get task details")

	ErrTitleCannotBeBlank = errors.New("title cannot be blank")

	ErrStatusCannotBeBlank = errors.New("status cannot be blank")
	ErrStatusNotValid      = errors.New("status must be 'doing' or 'done'")
)
