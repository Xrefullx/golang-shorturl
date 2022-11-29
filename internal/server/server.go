package server

import (
	"context"
	"errors"
	"github.com/Xrefullx/golang-shorturl/internal/handlers"
	"github.com/Xrefullx/golang-shorturl/internal/router"
	"net/http"
)

type server struct {
	httpServer http.Server
}

func Createserver(port string, handler handlers.Handler) *server {
	return &server{
		httpServer: http.Server{
			Addr:              port,
			Handler:           router.CreateRouter(handler),
			TLSConfig:         nil,
			ReadTimeout:       0,
			ReadHeaderTimeout: 0,
			WriteTimeout:      0,
			IdleTimeout:       0,
			MaxHeaderBytes:    0,
			TLSNextProto:      nil,
			ConnState:         nil,
			ErrorLog:          nil,
			BaseContext:       nil,
			ConnContext:       nil,
		},
	}
}
func (s *server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.ListenAndServe(); err == http.ErrServerClosed {
		return errors.New("server is not run")
	}
	return s.httpServer.Shutdown(ctx)
}
