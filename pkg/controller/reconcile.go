package controller

import (
	"context"
	"fmt"

	"github.com/axatol/actions-job-dispatcher/pkg/cache"
	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/axatol/actions-job-dispatcher/pkg/gh"
	"github.com/axatol/actions-job-dispatcher/pkg/k8s"
	"github.com/axatol/actions-job-dispatcher/pkg/util"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
)

func Reconcile(ctx context.Context) error {
	jobs, err := k8s.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %s", err)
	}

	for _, runner := range config.Runners {
		if err := reconcileRunner(ctx, runner, jobs); err != nil {
			log.Error().
				Err(err).
				Str("runner_scope", runner.Scope.String()).
				Strs("runner_labels", runner.Labels).
				Msg("failed to reconcile runner")
			return err
		}
	}

	return nil
}

func reconcileRunner(ctx context.Context, runner config.RunnerConfig, jobs []batchv1.Job) error {
	// existing runners
	var existingRunners []batchv1.Job
	for _, job := range jobs {
		labels := k8s.PrefixMapFromLabels(job.Labels).Extract()
		meta := cache.MetaFromStringMap(labels)

		if util.NewSet(runner.Labels...).EqualsStrs(meta.RunnerLabels) {
			existingRunners = append(existingRunners, job)
		}
	}

	// queued/in-progress workflow jobs
	var requestedJobs []cache.WorkflowJobMeta
	for _, meta := range cache.List() {
		// ignore if scope doesn't match
		if meta.Scope.String() != runner.Scope.String() {
			continue
		}

		// ignore if labels don't match
		if !util.NewSet(meta.RunnerLabels...).EqualsStrs(runner.Labels) {
			continue
		}

		// we dispatched a job previously
		if meta.StartedAt.After(meta.CreatedAt) {
			continue
		}

		gh, err := gh.GetClient(ctx, runner.Scope)
		if err != nil {
			return fmt.Errorf("failed to get github client for %s: %s", runner.Scope.String(), err)
		}

		// describe actual job
		job, err := gh.DescribeWorkflowJob(ctx, &meta)
		if err != nil {
			return fmt.Errorf("failed to describe cached job %s: %s", *job.HTMLURL, err)
		}

		switch job.GetStatus() {
		case "queued":
			// a job we should care about
			requestedJobs = append(requestedJobs, meta)

		case "in_progress":
			// ignore

		case "completed":
			// delete if completed (and we somehow missed the webhook)
			cache.Del(job.GetID())
		}
	}

	log := log.With().
		Str("runner_scope", runner.Scope.String()).
		Strs("runner_labels", runner.Labels).
		Int("existing_runner_count", len(existingRunners)).
		Int("requested_job_count", len(requestedJobs)).Logger()

	// cannot exceed limits
	if len(existingRunners) >= runner.MaxReplicas {
		log.Warn().Msg("runner is at maximum replicas")
		return nil
	}

	// enough runners to satisfy jobs
	if len(existingRunners) >= len(requestedJobs) {
		log.Debug().Msg("runner replicas sufficient")
		return nil
	}

	delta := len(requestedJobs) - len(existingRunners)
	count := util.ClampInt(delta, 0, runner.MaxReplicas)
	log.Info().Int("new_jobs", count).Msg("dispatching jobs")

	for i := 0; i < count; i++ {
		if err := Dispatch(ctx, runner); err != nil {
			return err
		}
	}

	return nil
}
