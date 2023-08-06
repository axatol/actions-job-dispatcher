package server

import (
	"fmt"
	"net/http"

	"github.com/axatol/actions-job-dispatcher/pkg/server/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(serverPort int64) *http.Server {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware_Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json", "text/json"))

	router.Get("/ping", handlers.Ping)
	router.Get("/health", handlers.HealthCheck)
	router.Get("/runners", handlers.ListRunners)
	router.Get("/jobs", handlers.ListJobs)
	router.Post("/webhook", handlers.ReceiveGithubWebhook)

	addr := fmt.Sprintf(":%d", serverPort)
	return &http.Server{Addr: addr, Handler: router}
}
