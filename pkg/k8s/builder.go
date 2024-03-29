package k8s

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/axatol/actions-job-dispatcher/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	JobSelectorKey   = "app.kubernetes.io/managed-by"
	JobSelectorValue = "actions-job-dispatcher"
	JobSelector      = labels.Set(map[string]string{JobSelectorKey: JobSelectorValue})
)

type Job struct {
	Env         EnvMap
	Annotations PrefixMap
	Labels      PrefixMap
}

func (j Job) Hash(labels config.Labels) string {
	hasher := sha1.New()
	hasher.Write([]byte(labels.String()))
	hasher.Write([]byte(fmt.Sprint(time.Now().Unix())))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (j Job) AddLabel(key, value string) {
	if j.Labels == nil {
		j.Labels = PrefixMap{}
	}

	j.Labels.Add(key, value)
}

func (j Job) AddAnnotation(key, value string) {
	if j.Annotations == nil {
		j.Annotations = PrefixMap{}
	}

	j.Annotations.Add(key, value)
}

func (j Job) AddEnv(key, value string) {
	if j.Env == nil {
		j.Env = EnvMap{}
	}

	j.Env[key] = value
}

// note: need to include env vars "RUNNER_TOKEN" with a registration token
func (j Job) Render(runner config.RunnerConfig) batchv1.Job {
	name := fmt.Sprintf("runner-%s-%s", runner.Slug(), j.Hash(runner.Labels)[:8])

	// labels
	j.AddLabel("is-org", strconv.FormatBool(runner.Scope.IsOrg))
	j.AddLabel("repository-owner", runner.Scope.Owner)
	j.AddLabel("repository-name", runner.Scope.Repository)
	j.AddLabel("runner-labels", runner.Labels.String())
	j.AddLabel("scope", runner.Scope.String())

	// environment variables
	j.AddEnv("DISABLE_RUNNER_UPDATE", "true")
	j.AddEnv("DOCKERD_IN_RUNNER", "true")
	// j.AddEnv("DOCKER_CERT_PATH", "/certs/client")
	j.AddEnv("DOCKER_ENABLED", "true")
	// j.AddEnv("DOCKER_HOST", "tcp://localhost:2376")
	// j.AddEnv("DOCKER_TLS_VERIFY", "1")
	j.AddEnv("GITHUB_ACTIONS_RUNNER_EXTRA_USER_AGENT", "actions-job-dispatcher/v0.0.1")
	j.AddEnv("GITHUB_URL", "https://github.com/")
	j.AddEnv("MTU", "1400")
	j.AddEnv("RUNNER_EPHEMERAL", "true")
	j.AddEnv("RUNNER_LABELS", runner.Labels.String())
	j.AddEnv("RUNNER_NAME", name)
	j.AddEnv("RUNNER_STATUS_UPDATE_HOOK", "false")
	j.AddEnv("RUNNER_WORKDIR", "/runner/_work")

	if runner.Scope.Repository != "" {
		j.AddEnv("RUNNER_REPO", fmt.Sprintf("%s/%s", runner.Scope.Owner, runner.Scope.Repository))
	} else {
		j.AddEnv("RUNNER_ORG", runner.Scope.Owner)
	}

	return batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: config.Namespace,
			Labels:    j.Labels,
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
					ServiceAccountName:            runner.ServiceAccountName,
					RestartPolicy:                 corev1.RestartPolicyNever,
					DNSPolicy:                     corev1.DNSClusterFirst,
					EnableServiceLinks:            util.Ptr(true),

					Containers: []corev1.Container{{
						Name:            "runner",
						Image:           runner.Image,
						ImagePullPolicy: corev1.PullAlways,

						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(runner.Resources.CPULimit),
								corev1.ResourceMemory: resource.MustParse(runner.Resources.MemoryLimit),
							},
						},

						SecurityContext: &corev1.SecurityContext{
							Privileged: util.Ptr(true),
						},

						// LivenessProbe: ,
						// ReadinessProbe: ,
						// StartupProbe: ,

						Env: j.Env.EnvVarList(),

						EnvFrom: []corev1.EnvFromSource{{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "actions-runner-job-config",
								},
								Optional: util.Ptr(true),
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

func NewRunnerJob() Job {
	return Job{
		Env:         EnvMap{},
		Labels:      PrefixMap{},
		Annotations: PrefixMap{},
	}
}
