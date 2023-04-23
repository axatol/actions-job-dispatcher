package job

import (
	"context"
	"fmt"
	"sync"

	"github.com/axatol/actions-runner-broker/pkg/cache"
	"github.com/axatol/actions-runner-broker/pkg/config"
	"github.com/axatol/actions-runner-broker/pkg/util"
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

	jobs, err := client.BatchV1().Jobs(config.Namespace).List(ctx, opts)
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
					Str("runner_label", runner.RunnerLabel).
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
		if label, ok := job.Labels[runnerLabelKey]; ok && label == runner.RunnerLabel {
			matchingJobs += 1
		}
	}

	// cannot exceed limits
	if matchingJobs >= runner.MaxReplicas {
		return nil
	}

	// queued/in-progress workflow jobs
	matchingMeta := 0
	for _, meta := range cache.List() {
		if meta.RunnerLabel == runner.RunnerLabel {
			matchingMeta += 1
		}
	}

	// enough runners to satisfy jobs
	if matchingMeta <= matchingJobs {
		return nil
	}

	count := util.MinInt(matchingMeta-matchingJobs, runner.MaxReplicas)

	log.Info().
		Str("runner_label", runner.RunnerLabel).
		Int("count", count).
		Msg("dispatching jobs")

	for i := 0; i < count; i++ {
		if err := Dispatch(ctx, client, runner); err != nil {
			return err
		}
	}

	return nil
}
