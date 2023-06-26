package job

import (
	"context"
	"fmt"
	"sync"

	"github.com/axatol/actions-job-dispatcher/pkg/cache"
	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/axatol/actions-job-dispatcher/pkg/util"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func Reconcile(ctx context.Context, client *kubernetes.Clientset) error {
	opts := metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{
			jobSelectorKey: jobSelectorValue,
		}).String(),
	}

	jobs, err := client.BatchV1().Jobs(config.Namespace.Value()).List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %s", err)
	}

	wg := new(sync.WaitGroup)
	for _, runner := range config.Runners {
		wg.Add(1)
		go func(runner config.RunnerConfig) {
			defer wg.Done()

			if err := reconcileRunner(ctx, client, runner, jobs.Items); err != nil {
				log.Error().
					Err(err).
					Strs("runner_labels", runner.Labels).
					Msg("failed to reconcile runner")
			}
		}(runner)
	}

	wg.Wait()

	return nil
}

func reconcileRunner(
	ctx context.Context,
	client *kubernetes.Clientset,
	runner config.RunnerConfig,
	jobs []batchv1.Job,
) error {
	// existing runners
	matchingJobs := 0
	for _, job := range jobs {
		if label, ok := job.Labels[runnerLabelKey]; ok && label == runner.Labels.String() {
			matchingJobs += 1
		}
	}

	// queued/in-progress workflow jobs
	matchingMeta := 0
	for _, meta := range cache.List() {
		if util.NewSet(meta.Labels...).EqualsStrs(runner.Labels) {
			matchingMeta += 1
		}
	}

	event := log.Info().
		Strs("runner_label", runner.Labels).
		Int("jobs", matchingJobs).
		Int("requests", matchingMeta)

	// cannot exceed limits
	if matchingJobs >= runner.MaxReplicas {
		event.Msg("runner is at maximum replicas")
		return nil
	}

	// enough runners to satisfy jobs
	if matchingJobs >= matchingMeta {
		event.Msg("runner replicas sufficient")
		return nil
	}

	count := util.MinInt(matchingMeta-matchingJobs, runner.MaxReplicas)
	event.Int("new_jobs", count).Msg("dispatching jobs")

	for i := 0; i < count; i++ {
		if err := Dispatch(ctx, client, runner); err != nil {
			return err
		}
	}

	return nil
}
