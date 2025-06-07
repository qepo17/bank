package test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"bank/config"
	"bank/internal/db"
)

// TestDB wraps database connection with transaction support for testing
type TestDB struct {
	DB *sql.DB
	Tx *sql.Tx
}

// TestSuite provides test environment setup and teardown
type TestSuite struct {
	config *config.Config
	db     *sql.DB
}

// NewTestSuite creates a new test suite with database connection
func NewTestSuite() (*TestSuite, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// Use test database if not already configured
	if cfg.DBName == "postgres" || cfg.DBName == "" {
		cfg.DBName = "bank_test"
	}

	database, err := db.New(cfg.DBHost, cfg.DBPort, cfg.DBCustomer, cfg.DBPassword, cfg.DBName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &TestSuite{
		config: cfg,
		db:     database,
	}, nil
}

// BeginTransaction starts a new transaction for a test
func (ts *TestSuite) BeginTransaction(t *testing.T) *TestDB {
	t.Helper()

	tx, err := ts.db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	return &TestDB{
		DB: ts.db,
		Tx: tx,
	}
}

// BeginWithoutTransaction starts a new transaction for a test
func (ts *TestSuite) BeginWithoutTransaction(t *testing.T) *TestDB {
	t.Helper()

	return &TestDB{
		DB: ts.db,
	}
}

// Close closes the database connection
func (ts *TestSuite) Close() error {
	if ts.db != nil {
		return ts.db.Close()
	}
	return nil
}

// Rollback rolls back the transaction and cleans up
func (tdb *TestDB) Rollback(t *testing.T) {
	t.Helper()

	if tdb.Tx != nil {
		if err := tdb.Tx.Rollback(); err != nil {
			t.Errorf("failed to rollback transaction: %v", err)
		}
	}
}

// Exec executes a query within the transaction
func (tdb *TestDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tdb.Tx.Exec(query, args...)
}

// Query executes a query within the transaction and returns rows
func (tdb *TestDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tdb.Tx.Query(query, args...)
}

// QueryRow executes a query within the transaction and returns a single row
func (tdb *TestDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return tdb.Tx.QueryRow(query, args...)
}

// Prepare prepares a statement within the transaction
func (tdb *TestDB) Prepare(query string) (*sql.Stmt, error) {
	return tdb.Tx.Prepare(query)
}

// SetupTestWithTransaction is a helper function to set up a test with transaction isolation
func SetupTestWithTransaction(t *testing.T) (*TestDB, func()) {
	t.Helper()

	suite, err := NewTestSuite()
	if err != nil {
		t.Fatalf("failed to create test suite: %v", err)
	}

	testDB := suite.BeginTransaction(t)

	cleanup := func() {
		testDB.Rollback(t)
		suite.Close()
	}

	return testDB, cleanup
}

// SetupTestWithoutTransaction is a helper function to set up a test without transaction isolation
func SetupTestWithoutTransaction(t *testing.T) (*TestDB, func()) {
	t.Helper()

	suite, err := NewTestSuite()
	if err != nil {
		t.Fatalf("failed to create test suite: %v", err)
	}

	testDB := suite.BeginWithoutTransaction(t)

	cleanup := func() {
		tables, err := getAllTables(suite.db)
		if err != nil {
			t.Fatalf("failed to get all tables: %v", err)
		}

		for _, table := range tables {
			_, err := testDB.DB.ExecContext(context.Background(), "TRUNCATE TABLE "+table+" CASCADE")
			if err != nil {
				t.Fatalf("failed to truncate table %s: %v", table, err)
			}
		}

		suite.Close()
	}

	return testDB, cleanup
}

func getAllTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// RunWithTransaction runs a test function with transaction isolation
func RunWithTransaction(t *testing.T, testFunc func(*TestDB)) {
	t.Helper()

	testDB, cleanup := SetupTestWithTransaction(t)
	defer cleanup()

	testFunc(testDB)
}

// RunWithoutTransaction runs a test function without transaction isolation
// Cleanup by truncating the database
func RunWithoutTransaction(t *testing.T, testFunc func(*TestDB)) {
	t.Helper()

	testDB, cleanup := SetupTestWithoutTransaction(t)
	defer cleanup()

	testFunc(testDB)
}

// SetupSuite sets up the test suite for the entire test package
// This should be called in TestMain to ensure proper setup/teardown
func SetupSuite() (*TestSuite, error) {
	return NewTestSuite()
}

// TeardownSuite cleans up the test suite
func TeardownSuite(suite *TestSuite) {
	if suite != nil {
		suite.Close()
	}
}
