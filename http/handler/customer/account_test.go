package customer_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCreateAccount(t *testing.T) {
	testCases := []struct {
		name           string
		request        string // json
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			request: `{
				"account_id": 123,
				"initial_balance": "100"
			}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request",
			request: `{
				"account_id": 123,
				"initial_balance": "invalid"
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "account_id empty",
			request: `{
				"initial_balance": "100.12345"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"account id is required"}`,
		},
		{
			name: "initial_balance empty",
			request: `{
				"account_id": 123
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"initial balance is required"}`,
		},
		{
			name: "initial_balance negative",
			request: `{
				"account_id": 123,
				"initial_balance": "-100"
			}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"initial balance must be greater than 0"}`,
		},
	}

	for _, tc := range testCases {
		testHandler(t, func(t *testing.T, handler *handlerFixture) {
			t.Run(tc.name, func(t *testing.T) {
				req := createRequest(t, "POST", "/accounts", tc.request)

				rr := httptest.NewRecorder()
				handler.handler.CreateAccount()(rr, req)

				if tc.expectedStatus != 0 && rr.Code != tc.expectedStatus {
					t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
				}

				if tc.expectedBody != "" && rr.Body.String() != tc.expectedBody {
					t.Errorf("expected body %s, got %s", tc.expectedBody, rr.Body.String())
				}

				if tc.expectedStatus != http.StatusCreated {
					return
				}

				// Check if the account and transaction are created
				var accountID int64
				err := handler.db.QueryRow("SELECT id FROM accounts WHERE id = $1", 123).Scan(&accountID)
				if err != nil {
					t.Fatalf("failed to query account: %v", err)
				}

				if accountID != 123 {
					t.Errorf("expected account id %d, got %d", 123, accountID)
				}

				var transactionAmount decimal.Decimal
				err = handler.db.QueryRow("SELECT amount FROM transactions WHERE account_id = $1", 123).Scan(&transactionAmount)
				if err != nil {
					t.Fatalf("failed to query transaction: %v", err)
				}

				if transactionAmount.String() != "100" {
					t.Errorf("expected transaction amount %s, got %s", "100", transactionAmount.String())
				}
			})
		})
	}
}

func TestGetAccountBalance(t *testing.T) {
	testCases := []struct {
		name              string
		accountID         string
		setupDB           func(t *testing.T, db *sql.DB)
		expectedStatus    int
		expectedBody      string
		expectedErrorLogs bool
	}{
		{
			name:      "success - account with balance",
			accountID: "123",
			setupDB: func(t *testing.T, db *sql.DB) {
				// Create account
				_, err := db.Exec("INSERT INTO accounts (id, created_at, updated_at) VALUES ($1, NOW(), NOW())", 123)
				if err != nil {
					t.Fatalf("failed to create account: %v", err)
				}

				// Create balance snapshot
				_, err = db.Exec(`
					INSERT INTO account_balance_snapshots (account_id, balance, last_transaction_id, created_at) 
					VALUES ($1, $2, $3, NOW())
				`, 123, "150.51234", 1)
				if err != nil {
					t.Fatalf("failed to create balance snapshot: %v", err)
				}

				// Create initial transaction (to match the last_transaction_id)
				_, err = db.Exec(`
					INSERT INTO transactions (id, account_id, amount, trx_type, created_at) 
					VALUES ($1, $2, $3, 'CREDIT', NOW())
				`, 1, 123, "150.500000")
				if err != nil {
					t.Fatalf("failed to create transaction: %v", err)
				}

				// Create another transaction
				_, err = db.Exec(`
					INSERT INTO transactions (id, account_id, amount, trx_type, created_at) 
					VALUES ($1, $2, $3, 'CREDIT', NOW())
				`, 2, 123, "100.000000")
				if err != nil {
					t.Fatalf("failed to create transaction: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
			// This should be calculated as 150.51234 + 100.000000 = 250.51234
			expectedBody: `{"account_id":123,"balance":"250.51234"}`,
		},
		{
			name:      "success - account with zero balance",
			accountID: "456",
			setupDB: func(t *testing.T, db *sql.DB) {
				// Create account
				_, err := db.Exec("INSERT INTO accounts (id, created_at, updated_at) VALUES ($1, NOW(), NOW())", 456)
				if err != nil {
					t.Fatalf("failed to create account: %v", err)
				}

				// Create balance snapshot with zero balance
				_, err = db.Exec(`
					INSERT INTO account_balance_snapshots (account_id, balance, last_transaction_id, created_at) 
					VALUES ($1, $2, $3, NOW())
				`, 456, "0.000000", 0)
				if err != nil {
					t.Fatalf("failed to create balance snapshot: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"account_id":456,"balance":"0"}`,
		},
		{
			name:           "invalid account_id - not a number",
			accountID:      "abc",
			setupDB:        func(t *testing.T, db *sql.DB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid account id"}`,
		},
		{
			name:           "invalid account_id - negative number",
			accountID:      "-123",
			setupDB:        func(t *testing.T, db *sql.DB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid account id"}`,
		},
		{
			name:           "invalid account_id - empty",
			accountID:      "",
			setupDB:        func(t *testing.T, db *sql.DB) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid account id"}`,
		},
		{
			name:              "account not found",
			accountID:         "999",
			setupDB:           func(t *testing.T, db *sql.DB) {},
			expectedStatus:    http.StatusBadRequest,
			expectedBody:      `{"error":"invalid account"}`,
			expectedErrorLogs: true,
		},
	}

	for _, tc := range testCases {
		testHandler(t, func(t *testing.T, handler *handlerFixture) {
			t.Run(tc.name, func(t *testing.T) {
				tc.setupDB(t, handler.db)

				// Create request with account_id parameter
				req := createRequest(t, "GET", "/api/accounts/{account_id}/balance", "", requestParam{key: "account_id", value: tc.accountID})

				rr := httptest.NewRecorder()
				handler.handler.GetAccountBalance()(rr, req)

				if rr.Code != tc.expectedStatus {
					t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
				}

				if tc.expectedBody != "" && rr.Body.String() != tc.expectedBody {
					t.Errorf("expected body %s, got %s", tc.expectedBody, rr.Body.String())
				}

				// Verify that error logs are generated when expected
				if tc.expectedErrorLogs && rr.Code == http.StatusInternalServerError {
					// This is a placeholder for verifying error logs
					// In a real scenario, you might want to capture and verify log output
				}
			})
		})
	}
}
