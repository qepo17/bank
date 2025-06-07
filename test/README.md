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

### 1. Using SetupTestWithTransaction (Recommended for simple tests)

```go
func TestSomething(t *testing.T) {
    testDB, cleanup := SetupTestWithTransaction(t)
    defer cleanup()

    // Your test code here
    // All database operations will be within a transaction
    // Transaction will be rolled back automatically
}
```

### 2. Using SetupTestWithoutTransaction (For tests requiring table truncation)

```go
func TestSomething(t *testing.T) {
    testDB, cleanup := SetupTestWithoutTransaction(t)
    defer cleanup()

    // Your test code here
    // All tables will be truncated after the test
    // Use this when you need to test across multiple transactions
}
```

### 3. Using RunWithTransaction (Recommended for complex tests)

```go
func TestSomething(t *testing.T) {
    RunWithTransaction(t, func(testDB *TestDB) {
        // Your test code here
        // All database operations will be within a transaction
        // Transaction will be rolled back automatically
    })
}
```

### 4. Using RunWithoutTransaction (For tests that have nested transaction)

```go
func TestSomething(t *testing.T) {
    RunWithoutTransaction(t, func(testDB *TestDB) {
        // Your test code here
        // All tables will be truncated after the test
        // Use this when you need to test across multiple transactions
    })
}
```

### 5. Using TestMain for package-level setup

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
- `Rollback(t)` - Manually rollback the transaction (usually handled automatically)

## Transaction vs Non-Transaction Modes

### Transaction Mode (Recommended)
- **Functions**: `SetupTestWithTransaction`, `RunWithTransaction`
- **Cleanup**: Automatic transaction rollback
- **Use case**: Most database tests where isolation is important
- **Performance**: Faster cleanup (rollback vs truncate)

### Non-Transaction Mode
- **Functions**: `SetupTestWithoutTransaction`, `RunWithoutTransaction`
- **Cleanup**: Table truncation (all tables except `goose_db_version`)
- **Use case**: Tests that need to work across multiple transactions or test transaction behavior itself
- **Performance**: Slower cleanup due to table truncation

## Example Usage

### Simple Customer Operations
```go
func TestCustomerOperations(t *testing.T) {
    testDB, cleanup := SetupTestWithTransaction(t)
    defer cleanup()

    // Insert a customer
    _, err := testDB.Exec(`
        INSERT INTO customers (first_name, last_name, phone_number, email_address) 
        VALUES ($1, $2, $3, $4)`,
        "John", "Doe", "+1234567890", "john.doe@example.com")
    if err != nil {
        t.Fatalf("Failed to insert customer: %v", err)
    }

    // Query the customer
    var firstName, lastName string
    err = testDB.QueryRow(`
        SELECT first_name, last_name 
        FROM customers 
        WHERE email_address = $1`,
        "john.doe@example.com").Scan(&firstName, &lastName)
    if err != nil {
        t.Fatalf("Failed to query customer: %v", err)
    }

    // Verify the data
    if firstName != "John" || lastName != "Doe" {
        t.Errorf("Expected John Doe, got %s %s", firstName, lastName)
    }
}
```

### Multiple Operations with RunWithTransaction
```go
func TestMultipleOperations(t *testing.T) {
    RunWithTransaction(t, func(testDB *TestDB) {
        customers := []struct {
            firstName, lastName, email string
        }{
            {"Alice", "Johnson", "alice@example.com"},
            {"Bob", "Wilson", "bob@example.com"},
            {"Charlie", "Brown", "charlie@example.com"},
        }

        // Insert multiple customers
        for _, customer := range customers {
            _, err := testDB.Exec(`
                INSERT INTO customers (first_name, last_name, email_address) 
                VALUES ($1, $2, $3)`,
                customer.firstName, customer.lastName, customer.email)
            if err != nil {
                t.Fatalf("Failed to insert customer %s: %v", customer.firstName, err)
            }
        }

        // Count customers
        var count int
        err := testDB.QueryRow("SELECT COUNT(*) FROM customers").Scan(&count)
        if err != nil {
            t.Fatalf("Failed to count customers: %v", err)
        }

        if count != 3 {
            t.Errorf("Expected 3 customers, got %d", count)
        }
    })
}
```

## Best Practices

1. **Prefer transaction mode**: Use `SetupTestWithTransaction` or `RunWithTransaction` for most tests
2. **Always use cleanup**: Use `defer cleanup()` when using `SetupTestWithTransaction/WithoutTransaction`
3. **Use RunWith* helpers**: Prefer `RunWithTransaction` for complex test scenarios as it handles cleanup automatically
4. **Don't commit manually**: Never call `Commit()` on transactions in tests
5. **Keep test data minimal**: Use focused, easy-to-verify test data
6. **Test both success and failure**: Cover both happy path and error scenarios
7. **Use non-transaction mode sparingly**: Only when you need to test across multiple transactions

## Database Configuration

The test suite automatically:
- Uses `bank_test` database if the main database is `postgres` or empty
- Connects using the same configuration as the main application
- Handles connection pooling and cleanup

## Best Practices

1. Always use `defer cleanup()` when using `SetupTestWithTransaction/WithoutTransaction`
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
- Make sure `bank_test` database is created

### Transaction Issues
- Don't call `Commit()` on transactions in tests
- Make sure all prepared statements are properly closed
- Check for nested transactions (not supported in PostgreSQL)
- Use non-transaction mode if you need to test transaction behavior

### Performance Issues
- Consider using connection pooling settings
- Avoid long-running transactions in tests
- Use prepared statements for repeated queries
- Prefer transaction mode over non-transaction mode for faster cleanup

### Table Truncation Issues (Non-Transaction Mode)
- Ensure foreign key constraints are properly handled with CASCADE
- Check that all tables are being truncated except `goose_db_version`
- Verify that table truncation order doesn't cause constraint violations 