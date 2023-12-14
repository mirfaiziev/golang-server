package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/mirfaiziev/golang-server/internal/app/http/handler"
	"github.com/mirfaiziev/golang-server/internal/infra/log"
)

type Config struct {
	ServerAddr      string        `split_words:"true" default:"0.0.0.0:8080"`
	ShutdownTimeout time.Duration `split_words:"true" default:"30s"`
}

func SetupServer(ctx context.Context, cfg Config) *http.Server {
	return &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router(),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
}

func router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(log.RequestLogger)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/hello", handler.Hello)

	return r
}
