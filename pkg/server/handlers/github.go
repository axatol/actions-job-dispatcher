package handlers

import (
	"fmt"
	"net/http"

	"github.com/axatol/actions-job-dispatcher/pkg/cache"
	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/axatol/actions-job-dispatcher/pkg/job"
	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog/log"
)

func ReceiveGithubWebhook(w http.ResponseWriter, r *http.Request) {
	client := config.KubeClientFromContext(r.Context())
	if client == nil {
		ResponseErr(fmt.Errorf("no kubernetes client in context"), "")
	}

	webhookType := github.WebHookType(r)
	log := log.With().
		Str("github_webhook_type", webhookType).
		Str("github_delivery_id", github.DeliveryID(r)).
		Str("github_hook_id", r.Header.Get("X-GitHub-Hook-ID")).
		Logger()

	payload, err := github.ValidatePayload(r, []byte{})
	if err != nil {
		ResponseErr(err, "received invalid payload").Write(w, log)
		return
	}

	event, err := github.ParseWebHook(webhookType, payload)
	if err != nil {
		ResponseErr(err, "could not parse webhook").Write(w, log)
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
		log.Info().
			Str("job_status", e.GetWorkflowJob().GetStatus()).
			Str("job_conclusion", e.GetWorkflowJob().GetConclusion()).
			Int("job_id", int(e.GetWorkflowJob().GetID())).
			Int("run_id", int(e.GetWorkflowJob().GetRunID())).
			Int("run_attempt", int(e.GetWorkflowJob().GetRunAttempt())).
			Str("html_url", e.GetWorkflowJob().GetHTMLURL()).
			Str("workflow_name", e.GetWorkflowJob().GetWorkflowName()).
			Str("job_name", e.GetWorkflowJob().GetName()).
			Msg("handling workflow job webhook")

		cache.CacheWorkflowJobEvent(e)
		if e.GetAction() == "queued" {
			if err := job.DispatchByEvent(r.Context(), client, e); err != nil {
				ResponseErr(err, "failed to dispatch job").Write(w, log)
				return
			}
		}

	default:
		log.Info().Str("event_type", webhookType).Msg("ignoring webhook")
		ResponseOK("")
		return
	}
}
