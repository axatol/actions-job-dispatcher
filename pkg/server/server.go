package server

import (
	"net/http"

	"github.com/axatol/actions-job-dispatcher/pkg/server/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer() *http.Server {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware_Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json", "text/json"))

	router.Get("/api/health", handlers.DescribeHealth)
	router.Get("/api/runners", handlers.ListRunners)
	router.Get("/api/jobs", handlers.ListJobs)
	router.Post("/api/webhook", handlers.ReceiveGithubWebhook)

	return &http.Server{Addr: ":8000", Handler: router}
}
