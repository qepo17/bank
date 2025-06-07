package main

import (
	"bank/account"
	"bank/config"
	"bank/http/handler/customer"
	dbPkg "bank/internal/db"
	"bank/internal/repository"
	"bank/internal/server"
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
	"github.com/rs/zerolog/log"
)

const gracefulShutdownTimeout = time.Minute

func main() {
	cfg, err := config.Get()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get config")
	}

	db, err := dbPkg.New(cfg.DBHost, cfg.DBPort, cfg.DBCustomer, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	r := chi.NewRouter()
	srv := server.NewServer(net.JoinHostPort("", cfg.Port), r)

	if err := registerDependencies(r, db); err != nil {
		log.Fatal().Err(err).Msg("failed to register dependencies")
	}

	// Listen for OS interrupt signal
	exitSig := make(chan os.Signal, 1)
	signal.Notify(exitSig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-exitSig
		ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed to shutdown server")
		}
	}()

	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("serving http server")
	}
}

func registerDependencies(r *chi.Mux, db *sql.DB) error {
	accountRepository, err := repository.NewAccountRepository(db)
	if err != nil {
		return err
	}

	accountDomain, err := account.NewAccountDomain(db, accountRepository)
	if err != nil {
		return err
	}

	customerHandler, err := customer.NewHandler(accountDomain)
	if err != nil {
		return err
	}

	customerHandler.RegisterRoutes(r)
	return nil
}
