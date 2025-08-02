package pg

import "gorm.io/gorm"

type pgRepo struct {
	db *gorm.DB

	tracerName string
}

func NewPgRepo(db *gorm.DB) *pgRepo {
	return &pgRepo{
		db:         db,
		tracerName: "pgRepo",
	}
}
