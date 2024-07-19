package database

import (
	"time"

	"github.com/google/uuid"
)

// object mapping to table dummy
type DummyOrm struct {
	UserId    uuid.UUID `gorm:"column:user_id;primaryKey"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (DummyOrm) TableName() string {
	return "dummy"
}
