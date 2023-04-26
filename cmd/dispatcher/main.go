package main

import (
	"context"
	"net/http"
	"time"

	"github.com/axatol/actions-runner-broker/pkg/cache"
	"github.com/axatol/actions-runner-broker/pkg/config"
	"github.com/axatol/actions-runner-broker/pkg/job"
	"github.com/axatol/actions-runner-broker/pkg/server"
	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
)

func main() {
	ctx := context.Background()
	config.LoadConfig()

	if len(config.Runners) < 1 {
		log.Fatal().Msg("no runners configured")
	}

	client, err := config.KubeClient()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	serverVersion, err := client.ServerVersion()
	if err != nil {
		log.Fatal().Err(err).Msg("could not retrieve server details")
	}

	log.Info().
		Str("kubernetes_server", serverVersion.GitVersion).
		Strs("serving_runner_labels", config.Runners.Strs()).
		Bool("dry_run", config.DryRun.Value()).
		Int("sync_interval", config.SyncInterval.Value()).
		Send()

	router := server.NewRouter()
	router.Get("/api/health", handle_Health(client))
	router.Get("/api/runners", handle_ListRunners)
	router.Get("/api/jobs", handle_ListJobs)
	router.Post("/api/webhook", handle_GithubWebhook(client))

	done := make(chan struct{})
	go func() {
		err := server.NewServer(router).ListenAndServe()

		event := log.Info()
		if err != http.ErrServerClosed {
			event = log.Error()
		}

		event.Err(err).Msg("server exited")
		done <- struct{}{}
	}()

	ticker := time.Ticker{}
	for loop := true; loop; {
		select {
		case <-done:
			loop = false
		case <-ticker.C:
			if err := job.Reconcile(ctx, client); err != nil {
				log.Error().Err(err).Msg("could not reconcile")
			}
		}
	}
}

func handle_Health(client *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := client.ServerVersion(); err != nil {
			server.ResponseErr(err, "could not retrieve server details").Write(w)
			return
		}

		server.ResponseOK("").Write(w)
	}
}

func handle_ListRunners(w http.ResponseWriter, r *http.Request) {
	resp := server.ResponseOK("")
	resp.Data = config.Runners
	resp.Write(w)
}

func handle_ListJobs(w http.ResponseWriter, r *http.Request) {
	resp := server.ResponseOK("")
	resp.Data = cache.List()
	resp.Write(w)
}

func handle_GithubWebhook(client *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		webhookType := github.WebHookType(r)
		log := log.With().
			Str("github_webhook_type", webhookType).
			Str("github_delivery_id", github.DeliveryID(r)).
			Str("github_hook_id", r.Header.Get("X-GitHub-Hook-ID")).
			Logger()

		payload, err := github.ValidatePayload(r, []byte{})
		if err != nil {
			server.ResponseErr(err, "received invalid payload").Write(w, log)
			return
		}

		event, err := github.ParseWebHook(webhookType, payload)
		if err != nil {
			server.ResponseErr(err, "could not parse webhook").Write(w, log)
			return
		}

		switch e := event.(type) {
		case *github.PingEvent:
			log.Info().Msg("responding to ping")
			// wants a raw "pong"
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("pong"))
			return

		case *github.WorkflowJobEvent:
			log.Info().Msg("handling workflow job webhook")
			if cache.CacheWorkflowJobEvent(e) {
				if err := job.DispatchByEvent(r.Context(), client, e); err != nil {
					server.ResponseErr(err, "failed to dispatch job").Write(w, log)
					return
				}
			}

		default:
			log.Info().Str("event_type", webhookType).Msg("ignoring webhook")
			server.ResponseOK("")
			return
		}
	}
}
