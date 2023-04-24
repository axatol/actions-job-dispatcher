package job

import (
	"fmt"
	"reflect"

	"github.com/axatol/actions-runner-broker/pkg/config"
	"github.com/axatol/actions-runner-broker/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

func selectRunner(labels []string) *config.RunnerConfig {
	targetRunnerLabels := util.NewSet(labels...)
	for _, runner := range config.Runners {
		if targetRunnerLabels.EqualsStrs(runner.Labels) {
			return &runner
		}
	}

	return nil
}

type PrefixMap map[string]string

func (m PrefixMap) Add(key string, value any) {
	if value == nil {
		value = "unknown"
	}

	key = fmt.Sprintf("%s/%s", "actions-runner-broker", key)
	m[key] = fmt.Sprint(reflect.ValueOf(value))
}

type EnvMap map[string]string

func (m EnvMap) Add(key string, value string) {
	m[key] = value
}

func (m EnvMap) MaybeAdd(key string, value *string) {
	if value == nil {
		return
	}

	m.Add(key, *value)
}

func (m EnvMap) EnvVarList() []corev1.EnvVar {
	result := make([]corev1.EnvVar, len(m))

	i := 0
	for name, value := range m {
		result[i] = corev1.EnvVar{Name: name, Value: value}
		i += 1
	}

	return result
}
