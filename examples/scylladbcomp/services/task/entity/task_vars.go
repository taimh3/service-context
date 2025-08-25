package entity

import (
	"strings"

	"github.com/taimaifika/service-context/core"
	"github.com/taimaifika/service-context/examples/scylladbcomp/common"
)

// TaskCreateRequest is a struct that represents the request to create a new task
type TaskCreateRequest struct {
	core.SQLModel
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (TaskCreateRequest) TableName() string {
	return Task{}.TableName()
}

func (tc *TaskCreateRequest) Prepare(status TaskType) {
	tc.SQLModel = core.NewSQLModel()
}

func (tc *TaskCreateRequest) Mask() {
	tc.SQLModel.Mask(common.MaskTypeTask)
}

func (tc *TaskCreateRequest) Validate() error {
	tc.Title = strings.TrimSpace(tc.Title)

	if err := ValidateTitle(tc.Title); err != nil {
		return err
	}

	tc.Status = strings.ToLower(strings.TrimSpace(tc.Status))

	if err := ValidateStatus(TaskType(tc.Status)); err != nil {
		return err
	}

	return nil
}

// TaskUpdateRequest is a struct that represents the request to update a task
type TaskUpdateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

func (TaskUpdateRequest) TableName() string { return Task{}.TableName() }

func (t *TaskUpdateRequest) Validate() error {
	if title := t.Title; title != nil {
		s := strings.TrimSpace(*title)

		if err := ValidateTitle(s); err != nil {
			return err
		}

		t.Title = &s
	}

	if status := t.Status; status != nil {
		if err := ValidateStatus(TaskType(*status)); err != nil {
			return err
		}
	}

	return nil
}

type Filter struct {
	Status *string `json:"status,omitempty" form:"status"`
}
