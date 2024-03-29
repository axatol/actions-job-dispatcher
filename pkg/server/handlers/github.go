package handlers

import (
	"net/http"

	"github.com/axatol/actions-job-dispatcher/pkg/cache"
	"github.com/axatol/actions-job-dispatcher/pkg/config"
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
			Str("repository_name", e.GetRepo().GetName()).
			Str("repository_owner", e.GetRepo().GetOwner().GetLogin()).
			Str("workflow_job_status", e.GetWorkflowJob().GetStatus()).
			Str("workflow_job_conclusion", e.GetWorkflowJob().GetConclusion()).
			Int("workflow_job_id", int(e.GetWorkflowJob().GetID())).
			Int("workflow_run_id", int(e.GetWorkflowJob().GetRunID())).
			Int("workflow_job_run_attempt", int(e.GetWorkflowJob().GetRunAttempt())).
			Str("workflow_job_html_url", e.GetWorkflowJob().GetHTMLURL()).
			Str("workflow_job_workflow_name", e.GetWorkflowJob().GetWorkflowName()).
			Str("workflow_job_name", e.GetWorkflowJob().GetName()).
			Strs("workflow_job_labels", e.GetWorkflowJob().Labels).
			Logger()

		runner, err := controller.SelectRunner(e)
		if err != nil {
			log.Debug().Err(err).Msg("ignoring workflow_job webhook")
			ResponseOK().Write(w)
			return
		}

		cache.CacheWorkflowJobEvent(e)
		if e.GetAction() == "queued" {
			runner = &config.RunnerConfig{
				Image:              runner.Image,
				Labels:             runner.Labels,
				Resources:          runner.Resources,
				ServiceAccountName: runner.ServiceAccountName,
				Scope: config.Scope{
					IsOrg:      runner.Scope.IsOrg,
					Owner:      runner.Scope.Owner,
					Repository: e.GetRepo().GetName(),
				},
			}

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
