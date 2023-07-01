package handlers

import (
	"net/http"

	"github.com/axatol/actions-job-dispatcher/pkg/cache"
	"github.com/axatol/actions-job-dispatcher/pkg/controller"
	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog/log"
)

func ReceiveGithubWebhook(w http.ResponseWriter, r *http.Request) {
	webhookType := github.WebHookType(r)
	log := log.With().
		Str("github_webhook_type", webhookType).
		Str("github_delivery_id", github.DeliveryID(r)).
		Str("github_hook_id", r.Header.Get("X-GitHub-Hook-ID")).
		Logger()

	payload, err := github.ValidatePayload(r, []byte{})
	if err != nil {
		ResponseErr(err).SetMessage("received invalid payload").Write(w, log)
		return
	}

	event, err := github.ParseWebHook(webhookType, payload)
	if err != nil {
		ResponseErr(err).SetMessage("could not parse webhook").Write(w, log)
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
		log := log.With().
			Str("job_status", e.GetWorkflowJob().GetStatus()).
			Str("job_conclusion", e.GetWorkflowJob().GetConclusion()).
			Int("job_id", int(e.GetWorkflowJob().GetID())).
			Int("run_id", int(e.GetWorkflowJob().GetRunID())).
			Int("run_attempt", int(e.GetWorkflowJob().GetRunAttempt())).
			Str("html_url", e.GetWorkflowJob().GetHTMLURL()).
			Str("workflow_name", e.GetWorkflowJob().GetWorkflowName()).
			Str("job_name", e.GetWorkflowJob().GetName()).
			Strs("workflow_job_labels", e.GetWorkflowJob().Labels).
			Logger()
			//Msg("handling workflow job webhook")

		runner, err := controller.SelectRunner(e)
		if err != nil {
			log.Debug().Msgf("ignoring workflow_job webhook: %s", err)
			ResponseOK().Write(w)
			return
		}

		cache.CacheWorkflowJobEvent(e)
		if e.GetAction() == "queued" {
			if err := controller.Dispatch(r.Context(), *runner); err != nil {
				ResponseErr(err).SetMessage("failed to dispatch job").Write(w, log)
				return
			}
		}

	default:
		log.Info().Str("event_type", webhookType).Msg("ignoring webhook")
		ResponseOK().Write(w)
		return
	}
}
