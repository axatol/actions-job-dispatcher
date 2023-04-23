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
	hasher.Write([]byte(j.runner.RunnerLabel))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (j Job) AddLabel(key, value string) {
	j.labels.Add(key, value)
}

func (j Job) AddAnnotation(key, value string) {
	j.annotations.Add(key, value)
}

func (j Job) AddEnv(key, value string) {
	j.env[key] = value
}

func (j Job) Build() batchv1.Job {
	name := fmt.Sprintf("%s-%s", j.runner.RunnerLabel, j.Hash())
	j.env["RUNNER_NAME"] = name
	// TODO
	// j.env["RUNNER_TOKEN"] =

	return batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: config.Namespace,
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
		runner: runner,
		env: EnvMap{
			"DISABLE_RUNNER_UPDATE":                  "true",
			"RUNNER_LABELS":                          runner.RunnerLabel,
			"DOCKER_ENABLED":                         "true",
			"DOCKERD_IN_RUNNER":                      "true",
			"GITHUB_URL":                             "https://github.com/",
			"RUNNER_WORKDIR":                         "/runner/_work",
			"RUNNER_EPHEMERAL":                       "true",
			"RUNNER_STATUS_UPDATE_HOOK":              "true",
			"GITHUB_ACTIONS_RUNNER_EXTRA_USER_AGENT": "actions-runner-broker/v0.0.1",
			"MTU":                                    "1400",
			"DOCKER_HOST":                            "tcp://localhost:2376",
			"DOCKER_TLS_VERIFY":                      "1",
			"DOCKER_CERT_PATH":                       "/certs/client",
		},
		labels: PrefixMap{
			runnerLabelKey: runner.RunnerLabel,
		},
		annotations: PrefixMap{},
	}

	return job
}
