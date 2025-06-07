package db

import (
	"database/sql"
)

func New(host, port, customer, pass, name string) (*sql.DB, error) {
	// Connect to the database here
	return &sql.DB{}, nil
}
