package test

import (
	"os"
	"testing"
)

// TestMain sets up and tears down the test environment
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

// Example test using the SetupTest helper
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

	// This data will be automatically rolled back when the test ends
}

// Example test using the RunWithTransaction helper
func TestCustomerUpdate(t *testing.T) {
	RunWithTransaction(t, func(testDB *TestDB) {
		// Insert a customer
		_, err := testDB.Exec(`
			INSERT INTO customers (first_name, last_name, phone_number, email_address) 
			VALUES ($1, $2, $3, $4)`,
			"Jane", "Smith", "+9876543210", "jane.smith@example.com")
		if err != nil {
			t.Fatalf("Failed to insert customer: %v", err)
		}

		// Update the customer
		_, err = testDB.Exec(`
			UPDATE customers 
			SET phone_number = $1 
			WHERE email_address = $2`,
			"+1111111111", "jane.smith@example.com")
		if err != nil {
			t.Fatalf("Failed to update customer: %v", err)
		}

		// Verify the update
		var phoneNumber string
		err = testDB.QueryRow(`
			SELECT phone_number 
			FROM customers 
			WHERE email_address = $1`,
			"jane.smith@example.com").Scan(&phoneNumber)
		if err != nil {
			t.Fatalf("Failed to query customer: %v", err)
		}

		if phoneNumber != "+1111111111" {
			t.Errorf("Expected +1111111111, got %s", phoneNumber)
		}

		// This data will be automatically rolled back when the test ends
	})
}

// Example test for multiple operations in one transaction
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

		// All data will be rolled back automatically
	})
}
