package job

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/axatol/actions-runner-broker/pkg/config"
	"github.com/axatol/actions-runner-broker/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	jobSelectorKey   = "app.kubernetes.io/managed-by"
	jobSelectorValue = "actions-runner-broker"
	runnerLabelKey   = "actions-runner-broker/runner-label"
)

type Job struct {
	runner      config.RunnerConfig
	env         EnvMap
	annotations PrefixMap
	labels      PrefixMap
}

func (j Job) Hash() string {
	hasher := sha1.New()
	hasher.Write([]byte(j.runner.Labels.String()))
	hasher.Write([]byte(fmt.Sprint(time.Now().Unix())))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (j Job) AddLabel(key, value string) {
	if j.labels == nil {
		j.labels = PrefixMap{}
	}

	j.labels.Add(key, value)
}

func (j Job) AddAnnotation(key, value string) {
	if j.annotations == nil {
		j.annotations = PrefixMap{}
	}

	j.annotations.Add(key, value)
}

func (j Job) AddEnv(key, value string) {
	if j.env == nil {
		j.env = EnvMap{}
	}

	j.env[key] = value
}

// note: need to include env var "RUNNER_TOKEN" for the runner to authenticate
func (j Job) Build() batchv1.Job {
	name := fmt.Sprintf("%s-%s", j.runner.Labels, j.Hash())

	j.AddLabel("runner-labels", j.runner.Labels.String())

	j.AddEnv("RUNNER_NAME", name)
	j.AddEnv("DISABLE_RUNNER_UPDATE", "true")
	j.AddEnv("RUNNER_LABELS", j.runner.Labels.String())
	j.AddEnv("DOCKER_ENABLED", "true")
	j.AddEnv("DOCKERD_IN_RUNNER", "true")
	j.AddEnv("GITHUB_URL", "https://github.com/")
	j.AddEnv("RUNNER_WORKDIR", "/runner/_work")
	j.AddEnv("RUNNER_EPHEMERAL", "true")
	j.AddEnv("RUNNER_STATUS_UPDATE_HOOK", "true")
	j.AddEnv("GITHUB_ACTIONS_RUNNER_EXTRA_USER_AGENT", "actions-runner-broker/v0.0.1")
	j.AddEnv("MTU", "1400")
	j.AddEnv("DOCKER_HOST", "tcp://localhost:2376")
	j.AddEnv("DOCKER_TLS_VERIFY", "1")
	j.AddEnv("DOCKER_CERT_PATH", "/certs/client")

	if j.runner.Scope.IsOrg() {
		j.AddEnv("RUNNER_ORG", j.runner.Scope.Organisation)
	}

	if j.runner.Scope.IsRepo() {
		j.AddEnv("RUNNER_REPO", j.runner.Scope.Repository)
	}

	return batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: config.Namespace.Value(),
			Labels:    j.labels,
		},

		Spec: batchv1.JobSpec{
			Parallelism:  util.Ptr(int32(1)),
			Completions:  util.Ptr(int32(1)),
			BackoffLimit: util.Ptr(int32(0)),

			ActiveDeadlineSeconds:   util.Ptr(int64(time.Hour.Seconds())),
			TTLSecondsAfterFinished: util.Ptr(int32(time.Minute.Seconds())),

			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: util.Ptr(int64((time.Minute * 5).Seconds())),
					// NodeSelector: map[string]string{},
					// ImagePullSecrets: []corev1.LocalObjectReference{},
					// Affinity: &corev1.Affinity{},
					// Tolerations: []corev1.Toleration{},
					// SchedulingGates: []corev1.PodSchedulingGate{},

					ServiceAccountName: j.runner.ServiceAccountName,
					PreemptionPolicy:   util.Ptr(corev1.PreemptNever),
					RestartPolicy:      corev1.RestartPolicyNever,

					// InitContainers: []corev1.Container{{
					// 	Name: "registration",
					// }},

					Containers: []corev1.Container{{
						Name:            "runner",
						Image:           j.runner.Image,
						ImagePullPolicy: corev1.PullAlways,

						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(j.runner.Resources.CPULimit),
								corev1.ResourceMemory: resource.MustParse(j.runner.Resources.MemoryLimit),
							},
						},

						// LivenessProbe: ,
						// ReadinessProbe: ,
						// StartupProbe: ,

						Env: j.env.EnvVarList(),

						EnvFrom: []corev1.EnvFromSource{{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "actions-runner-job-cm",
								},
							},
						}},

						VolumeMounts: []corev1.VolumeMount{
							{MountPath: "/runner", Name: "runner"},
							{MountPath: "/runner/_work", Name: "work"},
						},
					}},

					Volumes: []corev1.Volume{
						{Name: "runner", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
						{Name: "work", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
					},
				},
			},
		},
	}
}

func NewRunnerJob(runner config.RunnerConfig) Job {
	job := Job{
		runner:      runner,
		env:         EnvMap{},
		labels:      PrefixMap{},
		annotations: PrefixMap{},
	}

	return job
}
