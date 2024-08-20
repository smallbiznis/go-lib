package server

import (
	"context"
	"net/http"
	"os"

	"go.uber.org/fx"
)

var (
	port = ":8080"
)

func init() {
	newPort, ok := os.LookupEnv("PORT")
	if ok {
		port = newPort
	}
}

var Module = fx.Module("http.server", fx.Options(
	fx.Provide(
		NewServer,
	),
))

type IServer interface {
	RunTLS(string, string) error
	Run() error
	Down(ctx context.Context) error
}

type server struct {
	*http.Server
}

func NewServer(h http.Handler) IServer {
	srv := &http.Server{
		Addr:    port,
		Handler: h,
	}
	return &server{Server: srv}
}

// RunTLS
func (s *server) RunTLS(cert, key string) (err error) {
	return s.ListenAndServeTLS(cert, key)
}

// Run
func (s *server) Run() (err error) {
	return s.ListenAndServe()
}

// Down
func (s *server) Down(ctx context.Context) (err error) {
	return s.Shutdown(ctx)
}
