package entity

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/taimaifika/service-context/core"
	"github.com/taimaifika/service-context/examples/scylladbcomp/common"
)

type TaskType string

const (
	StatusDoing   TaskType = "doing"
	StatusDone    TaskType = "done"
	StatusDeleted TaskType = "deleted"
)

type Task struct {
	core.SQLModel
	Title       string   `json:"title" gorm:"column:title" db:"title"`
	Description string   `json:"description" gorm:"column:description" db:"description"`
	Status      TaskType `json:"status" gorm:"column:status" db:"status"`
}

func (Task) TableName() string {
	return "tasks"
}

func (t *Task) Mask() {
	t.SQLModel.Mask(common.MaskTypeTask)
}

func GenerateTaskID() int {
	// Generate a random 2-byte integer that fits safely in int16
	var bytes [2]byte
	rand.Read(bytes[:])

	// Convert to int64 and ensure it's positive
	return int(binary.BigEndian.Uint16(bytes[:]))
}
