package k8s

import (
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const prefixKey = "actions-job-dispatcher"

type PrefixMap map[string]string

func (m PrefixMap) prefixed(s string) string {
	if strings.HasPrefix(s, prefixKey+"/") {
		return s
	}

	return fmt.Sprintf("%s/%s", prefixKey, s)
}

func (m PrefixMap) unprefixed(s string) string {
	key, _ := strings.CutPrefix(s, prefixKey)
	return key
}

func (m PrefixMap) Add(key string, value any) {
	if value == nil {
		value = "unknown"
	}

	m[m.prefixed(key)] = fmt.Sprint(reflect.ValueOf(value))
}

// returns a map without prefixed keys
func (m PrefixMap) Extract() map[string]string {
	result := map[string]string{}
	for key, value := range m {
		result[m.unprefixed(key)] = value
	}

	return result
}

// creates a prefix map using only keys that have the matching prefix
func PrefixMapFromLabels(m map[string]string) PrefixMap {
	result := PrefixMap{}
	for key, value := range m {
		if strings.HasPrefix(key, prefixKey) {
			result.Add(key, value)
		}
	}

	return result
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
