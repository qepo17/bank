# Test Setup

This directory contains the test setup for the bank project with transaction-based test isolation.

## Overview

The test setup provides transaction-based isolation for database tests, ensuring that:
- Each test runs in its own transaction
- All database changes are automatically rolled back at the end of each test
- Tests are completely isolated from each other
- No test data persists between test runs

## Files

- `setup.go` - Core test setup and transaction management
- `example_test.go` - Example tests demonstrating usage patterns
- `README.md` - This documentation file

## Usage Patterns

### 1. Using SetupTest (Recommended for simple tests)

```go
func TestSomething(t *testing.T) {
    testDB, cleanup := SetupTest(t)
    defer cleanup()

    // Your test code here
    // All database operations will be within a transaction
    // Transaction will be rolled back automatically
}
```

### 2. Using RunWithTransaction (Recommended for complex tests)

```go
func TestSomething(t *testing.T) {
    RunWithTransaction(t, func(testDB *TestDB) {
        // Your test code here
        // All database operations will be within a transaction
        // Transaction will be rolled back automatically
    })
}
```

### 3. Using TestMain for package-level setup

```go
func TestMain(m *testing.M) {
    // Setup test suite
    suite, err := SetupSuite()
    if err != nil {
        panic("Failed to setup test suite: " + err.Error())
    }

    // Run tests
    code := m.Run()

    // Teardown
    TeardownSuite(suite)

    os.Exit(code)
}
```

## Database Operations

The `TestDB` struct provides methods for database operations within transactions:

- `Exec(query, args...)` - Execute a query (INSERT, UPDATE, DELETE)
- `Query(query, args...)` - Execute a query that returns multiple rows
- `QueryRow(query, args...)` - Execute a query that returns a single row
- `Prepare(query)` - Prepare a statement within the transaction

## Best Practices

1. Always use `defer cleanup()` when using `SetupTest`
2. Use `RunWithTransaction` for complex test scenarios
3. Don't commit transactions manually in tests
4. Keep test data minimal and focused
5. Use meaningful test data that's easy to verify
6. Test both success and failure scenarios

## Troubleshooting

### Connection Issues
- Ensure PostgreSQL is running
- Check environment variables
- Verify test database exists and is accessible

### Transaction Issues
- Don't call `Commit()` on transactions in tests
- Make sure all prepared statements are properly closed
- Check for nested transactions (not supported in PostgreSQL)

### Performance Issues
- Consider using connection pooling settings
- Avoid long-running transactions in tests
- Use prepared statements for repeated queries 