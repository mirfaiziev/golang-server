package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/mirfaiziev/golang-server/cmd/app/internal/app/http"
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
	// init logger, todo: change to slog
	logger := log.New(os.Stderr, "", 0)

	// init config
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		return fmt.Errorf("failed to load env vars: %w", err)
	}

	logger.Printf("Staring with config %+v", cfg)

	// init http server
	serverErrors := make(chan error, 1)
	var httpserver = http.SetupServer(ctx, cfg.Http)

	go func() {
		// todo: uncomment
		//logger.Info("initializing HTTP server", slog.String("host", cfg.Web.APIHost))

		serverErrors <- httpserver.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		//todo: change to log
		fmt.Errorf("server shutdown: %w", ctx.Err())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := httpserver.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}

		return fmt.Errorf("Successfuly shutdown")
	}
}
