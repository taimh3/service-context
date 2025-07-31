package entity

type TaskType string

const (
	StatusDoing   TaskType = "doing"
	StatusDone    TaskType = "done"
	StatusDeleted TaskType = "deleted"
)

type Task struct {
	ID          int      `json:"id" gorm:"column:id" db:"id"`
	Title       string   `json:"title" gorm:"column:title" db:"title"`
	Description string   `json:"description" gorm:"column:description" db:"description"`
	Status      TaskType `json:"status" gorm:"column:status" db:"status"`
}

func (Task) TableName() string {
	return "tasks"
}
