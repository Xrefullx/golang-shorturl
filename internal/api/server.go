package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/Xrefullx/golang-shorturl/internal/service"
	"github.com/Xrefullx/golang-shorturl/internal/storage"
	"github.com/Xrefullx/golang-shorturl/pkg"
	"net/http"
)

type Server struct {
	httpServer http.Server
}

func NewServer(cfg *pkg.Config, db storage.Storage) (*Server, error) {
	svcSht, err := service.NewShortURLService(db)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации handler:%w", err)
	}
	svcUser, err := service.NewUserService(db)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации handler:%w", err)
	}

	handler, err := NewHandler(svcSht, svcUser, cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации handler:%w", err)
	}
	return &Server{
		httpServer: http.Server{
			Addr:    cfg.ServerPort,
			Handler: CreateRouter(handler),
		},
	}, nil
}
func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.ListenAndServe(); err == http.ErrServerClosed {
		return errors.New("http server not runned")
	}
	return s.httpServer.Shutdown(ctx)
}
