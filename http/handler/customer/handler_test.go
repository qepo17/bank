package customer_test

import (
	"bank/account"
	"bank/http/handler/customer"
	"bank/internal/db/sqlc"
	"bank/internal/logger"
	"bank/test"
	"bank/transaction"
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

type handlerFixture struct {
	db      *sql.DB
	handler *customer.Handler
}

type testHandlerFunc func(t *testing.T, handler *handlerFixture)

func testHandler(t *testing.T, testFunc testHandlerFunc) {
	test.RunWithoutTransaction(t, func(testDB *test.TestDB) {
		testLogger := logger.NewLogger("debug")
		accountDomain, err := account.NewAccountDomain(testDB.DB, sqlc.New(testDB.DB), testLogger)
		if err != nil {
			t.Fatalf("failed to create account domain: %v", err)
		}

		transactionDomain, err := transaction.NewTransactionDomain(testDB.DB, sqlc.New(testDB.DB), testLogger)
		if err != nil {
			t.Fatalf("failed to create transaction domain: %v", err)
		}

		handler, err := customer.NewHandler(accountDomain, transactionDomain, testLogger)
		if err != nil {
			t.Fatalf("failed to create handler: %v", err)
		}

		testFunc(t, &handlerFixture{
			db:      testDB.DB,
			handler: handler,
		})
	})
}

type requestParam struct {
	key   string
	value string
}

func createRequest(t *testing.T, method, path string, body string, params ...requestParam) *http.Request {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	paramCtx := chi.NewRouteContext()
	for _, param := range params {
		paramCtx.URLParams.Add(param.key, param.value)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, paramCtx))
	req.Header.Set("Content-Type", "application/json")
	return req
}
