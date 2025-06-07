package entity

import (
	"time"
)

type Model struct {
	ID        uint64
	CreatedAt time.Time
}

type ModelWithUpdatedAt struct {
	Model
	UpdatedAt time.Time
}
