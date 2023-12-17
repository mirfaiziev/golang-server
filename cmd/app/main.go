package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/mirfaiziev/golang-server/internal/app/http"
)

type config struct {
	LogLevel string `split_words:"true" default:"info"`
	Http     http.Config
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// init config
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		return fmt.Errorf("failed to load env vars: %w", err)
	}

	logger.Info("Staring with config", slog.Any("config", cfg))

	// init http server
	serverErrors := make(chan error, 1)
	var httpserver = http.SetupServer(ctx, logger, cfg.Http)

	go func() {
		logger.Info("initializing HTTP server", slog.String("Host", cfg.Http.ServerAddr))
		serverErrors <- httpserver.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():

		logger.Error("server shutdown: %w", ctx.Err())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := httpserver.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}

		return fmt.Errorf("Successfully shutdown")
	}
}
