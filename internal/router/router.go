package router

import (
	"github.com/Xrefullx/golang-shorturl/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func CreateRouter(handler handlers.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/{shortID}", handler.GetHandler)
	r.Post("/", handler.SaveHandler)
	r.Post("/api/shorten", handler.JsonSave)
	return r
}
