package http

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/mirfaiziev/golang-server/internal/app/http/handler"
	"github.com/mirfaiziev/golang-server/internal/app/stats"
	"github.com/mirfaiziev/golang-server/internal/infra/log"
)

type Config struct {
	ServerAddr      string        `split_words:"true" default:"0.0.0.0:8080"`
	ShutdownTimeout time.Duration `split_words:"true" default:"30s"`
}

func SetupServer(ctx context.Context, logger *slog.Logger, cfg Config) *http.Server {
	return &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router(logger),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
}

func router(logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(log.RequestLogger)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Post("/analyze", handler.NewAnalyzeHandler(stats.NewService(), logger).Analyze)

	return r
}
