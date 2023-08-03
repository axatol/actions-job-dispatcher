package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/axatol/actions-job-dispatcher/pkg/gh"
	"github.com/axatol/actions-job-dispatcher/pkg/k8s"
	"github.com/axatol/actions-job-dispatcher/pkg/util"
	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog/log"
)

func Dispatch(ctx context.Context, runner config.RunnerConfig) error {
	gh, err := gh.GetClient(ctx, runner.Scope)
	if err != nil {
		return fmt.Errorf("failed to get github client: %s", err)
	}

	job := k8s.NewRunnerJob(runner)

	if config.DryRun {
		job.AddEnv("RUNNER_TOKEN", "DRYRUN")
	} else {
		token, err := gh.CreateRegistrationToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to create runner registration token: %s", err)
		}

		if !config.DryRun && token.Token != nil {
			job.AddEnv("RUNNER_TOKEN", *token.Token)
		}
	}

	tmpl := job.Render()

	if config.DryRun {
		log.Debug().Any("template", tmpl).Msg("dry run enabled: not dispatching")
		return nil
	}

	createdJob, err := k8s.CreateJob(ctx, tmpl)
	if err != nil {
		return fmt.Errorf("failed to dispatch job: %s", err)
	}

	log.Info().
		Str("job_name", createdJob.Name).
		Msg("dispatched job")

	return nil
}

func SelectRunner(event *github.WorkflowJobEvent) (*config.RunnerConfig, error) {
	targetRunnerLabels := util.NewSet(event.WorkflowJob.Labels...)
	for _, runner := range config.Runners {
		if runner.Scope.IsOrg && event.Org == nil {
			continue
		}

		if !runner.Scope.IsOrg && event.Org != nil {
			continue
		}

		if !targetRunnerLabels.EqualsStrs(runner.Labels) {
			continue
		}

		return &runner, nil
	}

	return nil, fmt.Errorf("no matching runner for labels: %s", strings.Join(event.WorkflowJob.Labels, ", "))
}
