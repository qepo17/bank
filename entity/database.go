package entity

import (
	"database/sql"
	"time"
)

type Model struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
}
