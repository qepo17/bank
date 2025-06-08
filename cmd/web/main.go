package main

import (
	"bank/account"
	"bank/config"
	"bank/http/handler/customer"
	dbPkg "bank/internal/db"
	"bank/internal/db/sqlc"
	"bank/internal/logger"
	"bank/internal/server"
	"bank/transaction"
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

const gracefulShutdownTimeout = time.Minute

func main() {
	cfg, err := config.Get()
	if err != nil {
		panic("failed to get config: " + err.Error())
	}

	log := logger.NewLogger(cfg.LogLevel)
	ctx := context.Background()

	db, err := dbPkg.New(cfg.DBHost, cfg.DBPort, cfg.DBCustomer, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatal(ctx, "failed to connect to database: %v", err)
	}
	defer db.Close()

	r := chi.NewRouter()
	srv := server.NewServer(net.JoinHostPort("", cfg.Port), r)

	if err := registerDependencies(r, db, cfg, log); err != nil {
		log.Fatal(ctx, "failed to register dependencies: %v", err)
	}

	log.Info(ctx, "starting server on port %s", cfg.Port)

	// Listen for OS interrupt signal
	exitSig := make(chan os.Signal, 1)
	signal.Notify(exitSig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-exitSig
		log.Info(ctx, "received shutdown signal, gracefully shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown server: %v", err)
		}
	}()

	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error(ctx, "serving http server: %v", err)
	}
}

func registerDependencies(r *chi.Mux, db *sql.DB, cfg *config.Config, log *logger.Logger) error {
	sqlc := sqlc.New(db)

	accountDomain, err := account.NewAccountDomain(db, sqlc, log)
	if err != nil {
		return err
	}

	transactionDomain, err := transaction.NewTransactionDomain(db, sqlc, log)
	if err != nil {
		return err
	}

	customerHandler, err := customer.NewHandler(accountDomain, transactionDomain, log)
	if err != nil {
		return err
	}

	customerHandler.RegisterRoutes(r)
	return nil
}
