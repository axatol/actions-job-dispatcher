package handlers

import (
	"net/http"

	"github.com/axatol/actions-job-dispatcher/pkg/cache"
	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/axatol/actions-job-dispatcher/pkg/gh"
	"github.com/axatol/actions-job-dispatcher/pkg/k8s"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	results := struct {
		Kubernetes bool            `json:"kubernetes"`
		GitHub     map[string]bool `json:"github"`
	}{
		Kubernetes: false,
		GitHub:     map[string]bool{},
	}

	log := log.Logger

	if kubeClient, err := k8s.GetClient(); err != nil {
		log = log.With().AnErr("kubernetes_client_error", err).Logger()
		results.Kubernetes = false
	} else if kubeVersion, err := kubeClient.Version(); err != nil {
		log = log.With().AnErr("kubernetes_version_error", err).Logger()
		results.Kubernetes = false
	} else {
		log = log.With().Interface("kubernetes_version", kubeVersion).Logger()
		results.Kubernetes = true
	}

	ghResultDict := zerolog.Dict()
	for _, runner := range config.Runners {
		scope := runner.Scope.String()
		ghClient, err := gh.GetClient(r.Context(), runner.Scope)
		if err != nil {
			ghResultDict.AnErr(scope, err)
			results.GitHub[scope] = false
			continue
		}

		ghDescribe, err := ghClient.DescribeScope(r.Context())
		if err != nil {
			ghResultDict.AnErr(scope, err)
			results.GitHub[scope] = false
			continue
		}

		ghResultDict.Str(scope, ghDescribe)
		results.GitHub[scope] = true
	}

	log.Info().
		Dict("github", ghResultDict).
		Msg("health")

	ResponseOK().SetData(results).Write(w)
}

func ListRunners(w http.ResponseWriter, r *http.Request) {
	ResponseOK().SetData(config.Runners).Write(w)
}

func ListJobs(w http.ResponseWriter, r *http.Request) {
	ResponseOK().SetData(cache.List()).Write(w)
}
