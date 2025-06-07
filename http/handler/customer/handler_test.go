package customer_test

import (
	"bank/account"
	"bank/http/handler/customer"
	"bank/internal/db/sqlc"
	"bank/internal/logger"
	"bank/test"
	"bytes"
	"database/sql"
	"net/http"
	"testing"
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

		handler, err := customer.NewHandler(accountDomain, testLogger)
		if err != nil {
			t.Fatalf("failed to create handler: %v", err)
		}

		testFunc(t, &handlerFixture{
			db:      testDB.DB,
			handler: handler,
		})
	})
}

func createRequest(t *testing.T, method, path string, body string) *http.Request {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}
