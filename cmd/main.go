// Package main provides an HTTP API for movie data.
//
//	@title			Kinosearch API
//	@version		1.0
//	@description	Simple movie JSON API for homework backend practice.
//	@host			localhost:8080
//	@BasePath		/
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/vacmannnn/kinosearch/internal/adapter"
	"github.com/vacmannnn/kinosearch/internal/config"
	"github.com/vacmannnn/kinosearch/internal/handler"
	"github.com/vacmannnn/kinosearch/internal/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("server stopped:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	logLevel, err := parseLogLevel(cfg.LoggerLevel)
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	pool, err := initPostgreSQL(cfg.DatabaseURL, cfg.DbTimeout)
	if err != nil {
		return err
	}
	defer pool.Close()

	moviesStorage := adapter.NewMoviesStorage(pool, logger)
	appService := service.New(moviesStorage, logger)
	httpHandler := handler.New(appService, logger)

	server := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: handler.NewRouter(httpHandler),
	}

	serverErrCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrCh:
		if err != nil {
			return err
		}
	case s := <-signalCh:
		fmt.Println("\ncatched signal:", s)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.ServerTimeout)*time.Second)
	defer shutdownCancel()

	return server.Shutdown(shutdownCtx)
}

func parseLogLevel(value string) (slog.Level, error) {
	var logLevel slog.Level
	if value == "" {
		return slog.LevelInfo, nil
	}

	if err := logLevel.UnmarshalText([]byte(strings.ToUpper(value))); err != nil {
		return slog.LevelInfo, fmt.Errorf("invalid logger_level %q: %w", value, err)
	}

	return logLevel, nil
}

func runMigrations(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "migrations")
}

func initPostgreSQL(databaseURL string, timeout int) (*pgxpool.Pool, error) {
	dbCtx, dbCancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer dbCancel()

	pool, err := pgxpool.New(dbCtx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(dbCtx); err != nil {
		pool.Close()
		return nil, err
	}

	if err := runMigrations(databaseURL); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
