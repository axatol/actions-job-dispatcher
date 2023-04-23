package job

import (
	"context"
	"fmt"
	"strings"

	"github.com/axatol/actions-runner-broker/pkg/config"
	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Dispatch(ctx context.Context, client *kubernetes.Clientset, runner config.RunnerConfig) error {
	job := NewRunnerJob(runner).Build()

	response, err := client.BatchV1().Jobs(job.ObjectMeta.Namespace).Create(ctx, &job, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	log.Info().
		Str("job_name", response.Name).
		Msg("dispatched job")

	return nil

}

func DispatchByEvent(ctx context.Context, client *kubernetes.Clientset, event github.WorkflowJobEvent) error {
	runner := selectRunner(event.WorkflowJob.Labels)
	if runner == nil {
		return fmt.Errorf("no matching runner for labels: %s", strings.Join(event.WorkflowJob.Labels, ", "))
	}

	return Dispatch(ctx, client, *runner)
}
