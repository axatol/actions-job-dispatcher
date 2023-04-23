package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.AllowContentType("application/json", "text/json"))
	return r
}

func NewServer(r *chi.Mux) *http.Server {
	if r == nil {
		r = NewRouter()
	}

	return &http.Server{Addr: ":8000", Handler: r}
}
