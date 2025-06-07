package customer_test

import (
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
