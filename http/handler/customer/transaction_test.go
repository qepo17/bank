package customer_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCreateTransferFunds(t *testing.T) {
	testCases := []struct {
		name           string
		request        string // json
		setupDB        func(t *testing.T, handler *handlerFixture)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success - valid transfer",
			request: `{
				"source_account_id": 100,
				"destination_account_id": 200,
				"amount": "50.123456"
			}`,
			setupDB: func(t *testing.T, handler *handlerFixture) {
				// Create source account with sufficient balance
				_, err := handler.db.Exec("INSERT INTO accounts (id, created_at, updated_at) VALUES ($1, NOW(), NOW())", 100)
				if err != nil {
					t.Fatalf("failed to create source account: %v", err)
				}

				// Create destination account
				_, err = handler.db.Exec("INSERT INTO accounts (id, created_at, updated_at) VALUES ($1, NOW(), NOW())", 200)
				if err != nil {
					t.Fatalf("failed to create destination account: %v", err)
				}

				// Create initial transaction for source account
				var transactionID uint64
				err = handler.db.QueryRow(`
					INSERT INTO transactions (account_id, amount, trx_type, created_at) 
					VALUES ($1, $2, 'CREDIT', NOW()) RETURNING id
				`, 100, "100.000000").Scan(&transactionID)
				if err != nil {
					t.Fatalf("failed to create initial transaction: %v", err)
				}

				// Create balance snapshot for source account
				_, err = handler.db.Exec(`
					INSERT INTO account_balance_snapshots (account_id, balance, last_transaction_id, created_at) 
					VALUES ($1, $2, $3, NOW())
				`, 100, "100.000000", transactionID)
				if err != nil {
					t.Fatalf("failed to create initial transaction: %v", err)
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid request - malformed JSON",
			request: `{
				"source_account_id": "invalid",
				"destination_account_id": 200,
				"amount": "50.123456"
			}`,
			setupDB:        func(t *testing.T, handler *handlerFixture) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "validation error - missing source_account_id",
			request: `{
				"destination_account_id": 200,
				"amount": "50.123456"
			}`,
			setupDB:        func(t *testing.T, handler *handlerFixture) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "validation error - missing destination_account_id",
			request: `{
				"source_account_id": 100,
				"amount": "50.123456"
			}`,
			setupDB:        func(t *testing.T, handler *handlerFixture) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "validation error - missing amount",
			request: `{
				"source_account_id": 100,
				"destination_account_id": 200
			}`,
			setupDB:        func(t *testing.T, handler *handlerFixture) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "validation error - negative amount",
			request: `{
				"source_account_id": 100,
				"destination_account_id": 200,
				"amount": "-50.123456"
			}`,
			setupDB:        func(t *testing.T, handler *handlerFixture) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "validation error - zero amount",
			request: `{
				"source_account_id": 100,
				"destination_account_id": 200,
				"amount": "0"
			}`,
			setupDB:        func(t *testing.T, handler *handlerFixture) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "validation error - too many decimal places",
			request: `{
				"source_account_id": 100,
				"destination_account_id": 200,
				"amount": "50.1234567"
			}`,
			setupDB:        func(t *testing.T, handler *handlerFixture) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name: "insufficient funds",
			request: `{
				"source_account_id": 100,
				"destination_account_id": 200,
				"amount": "10000.000000"
			}`,
			setupDB: func(t *testing.T, handler *handlerFixture) {
				// Create source account with insufficient balance
				_, err := handler.db.Exec("INSERT INTO accounts (id, created_at, updated_at) VALUES ($1, NOW(), NOW())", 100)
				if err != nil {
					t.Fatalf("failed to create source account: %v", err)
				}

				// Create destination account
				_, err = handler.db.Exec("INSERT INTO accounts (id, created_at, updated_at) VALUES ($1, NOW(), NOW())", 200)
				if err != nil {
					t.Fatalf("failed to create destination account: %v", err)
				}

				// Create initial transaction for source account with low balance
				var transactionID uint64
				err = handler.db.QueryRow(`
					INSERT INTO transactions (account_id, amount, trx_type, created_at) 
					VALUES ($1, $2, 'CREDIT', NOW()) RETURNING id
				`, 100, "100.000000").Scan(&transactionID)
				if err != nil {
					t.Fatalf("failed to create initial transaction: %v", err)
				}

				// Create balance snapshot for source account
				_, err = handler.db.Exec(`
					INSERT INTO account_balance_snapshots (account_id, balance, last_transaction_id, created_at) 
					VALUES ($1, $2, $3, NOW())
				`, 100, "100.000000", transactionID)
				if err != nil {
					t.Fatalf("failed to create balance snapshot: %v", err)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"your account has insufficient funds"}`,
		},
		{
			name: "invalid account - source account does not exist",
			request: `{
				"source_account_id": 999,
				"destination_account_id": 200,
				"amount": "50.123456"
			}`,
			setupDB: func(t *testing.T, handler *handlerFixture) {
				// Create destination account only
				_, err := handler.db.Exec("INSERT INTO accounts (id, created_at, updated_at) VALUES ($1, NOW(), NOW())", 200)
				if err != nil {
					t.Fatalf("failed to create destination account: %v", err)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid account"}`,
		},
		{
			name: "invalid account - destination account does not exist",
			request: `{
				"source_account_id": 100,
				"destination_account_id": 999,
				"amount": "50.123456"
			}`,
			setupDB: func(t *testing.T, handler *handlerFixture) {
				// Create source account only
				_, err := handler.db.Exec("INSERT INTO accounts (id, created_at, updated_at) VALUES ($1, NOW(), NOW())", 100)
				if err != nil {
					t.Fatalf("failed to create source account: %v", err)
				}

				// Create initial transaction for source account
				var transactionID uint64
				err = handler.db.QueryRow(`
					INSERT INTO transactions (account_id, amount, trx_type, created_at) 
					VALUES ($1, $2, 'CREDIT', NOW()) RETURNING id
				`, 100, "100.000000").Scan(&transactionID)
				if err != nil {
					t.Fatalf("failed to create initial transaction: %v", err)
				}

				// Create balance snapshot for source account
				_, err = handler.db.Exec(`
					INSERT INTO account_balance_snapshots (account_id, balance, last_transaction_id, created_at) 
					VALUES ($1, $2, $3, NOW())
				`, 100, "100.000000", transactionID)
				if err != nil {
					t.Fatalf("failed to create balance snapshot: %v", err)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid account"}`,
		},
		{
			name: "same account transfer",
			request: `{
				"source_account_id": 100,
				"destination_account_id": 100,
				"amount": "50.123456"
			}`,
			setupDB: func(t *testing.T, handler *handlerFixture) {
				// Create account
				_, err := handler.db.Exec("INSERT INTO accounts (id, created_at, updated_at) VALUES ($1, NOW(), NOW())", 100)
				if err != nil {
					t.Fatalf("failed to create account: %v", err)
				}

				// Create initial transaction
				var transactionID uint64
				err = handler.db.QueryRow(`
					INSERT INTO transactions (account_id, amount, trx_type, created_at) 
					VALUES ($1, $2, 'CREDIT', NOW()) RETURNING id
				`, 100, "100.000000").Scan(&transactionID)
				if err != nil {
					t.Fatalf("failed to create initial transaction: %v", err)
				}

				// Create balance snapshot
				_, err = handler.db.Exec(`
					INSERT INTO account_balance_snapshots (account_id, balance, last_transaction_id, created_at) 
					VALUES ($1, $2, $3, NOW())
				`, 100, "100.000000", transactionID)
				if err != nil {
					t.Fatalf("failed to create balance snapshot: %v", err)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"validation error: Cannot transfer to the same account"}`,
		},
	}

	for _, tc := range testCases {
		testHandler(t, func(t *testing.T, handler *handlerFixture) {
			t.Run(tc.name, func(t *testing.T) {
				// Setup database state
				tc.setupDB(t, handler)

				req := createRequest(t, "POST", "/transfer", tc.request)
				rr := httptest.NewRecorder()
				handler.handler.CreateTransferFunds()(rr, req)

				if tc.expectedStatus != 0 && rr.Code != tc.expectedStatus {
					t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
				}

				if tc.expectedBody != "" && rr.Body.String() != tc.expectedBody {
					t.Errorf("expected body %s, got %s", tc.expectedBody, rr.Body.String())
				}

				if tc.expectedStatus != http.StatusOK {
					return
				}

				// For successful transfers, verify the database state
				verifySuccessfulTransfer(t, handler, 100, 200, decimal.RequireFromString("50.123456"))
			})
		})
	}
}

// verifySuccessfulTransfer checks that a successful transfer created the correct database records
func verifySuccessfulTransfer(t *testing.T, handler *handlerFixture, sourceAccountID, destAccountID uint64, amount decimal.Decimal) {
	// Check that a transfer record was created
	var transferID int64
	err := handler.db.QueryRow(`
		SELECT id FROM transfers 
		WHERE from_account_id = $1 AND to_account_id = $2
	`, sourceAccountID, destAccountID).Scan(&transferID)
	if err != nil {
		t.Fatalf("failed to query transfer record: %v", err)
	}

	if transferID == 0 {
		t.Error("expected transfer record to be created")
	}

	// Check that debit transaction was created for source account
	var debitAmount decimal.Decimal
	var debitTrxType string
	err = handler.db.QueryRow(`
		SELECT amount, trx_type FROM transactions 
		WHERE account_id = $1 AND transfer_id = $2 AND trx_type = 'DEBIT'
	`, sourceAccountID, transferID).Scan(&debitAmount, &debitTrxType)
	if err != nil {
		t.Fatalf("failed to query debit transaction: %v", err)
	}

	if !debitAmount.Equal(amount) {
		t.Errorf("expected debit amount %s, got %s", amount.String(), debitAmount.String())
	}

	// Check that credit transaction was created for destination account
	var creditAmount decimal.Decimal
	var creditTrxType string
	err = handler.db.QueryRow(`
		SELECT amount, trx_type FROM transactions 
		WHERE account_id = $1 AND transfer_id = $2 AND trx_type = 'CREDIT'
	`, destAccountID, transferID).Scan(&creditAmount, &creditTrxType)
	if err != nil {
		t.Fatalf("failed to query credit transaction: %v", err)
	}

	if !creditAmount.Equal(amount) {
		t.Errorf("expected credit amount %s, got %s", amount.String(), creditAmount.String())
	}
}
