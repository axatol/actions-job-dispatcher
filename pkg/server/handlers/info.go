package handlers

import (
	"fmt"
	"net/http"

	"github.com/axatol/actions-job-dispatcher/pkg/cache"
	"github.com/axatol/actions-job-dispatcher/pkg/config"
)

func DescribeHealth(w http.ResponseWriter, r *http.Request) {
	kubeClient := config.KubeClientFromContext(r.Context())
	if kubeClient == nil {
		ResponseErr(fmt.Errorf("no kubernetes client in context"), "").Write(w)
		return
	}

	_, err := kubeClient.ServerVersion()
	if err != nil {
		ResponseErr(err, "could not retrieve server details").Write(w)
		return
	}

	ResponseOK("").Write(w)
}

func ListRunners(w http.ResponseWriter, r *http.Request) {
	resp := ResponseOK("")
	resp.Data = config.Runners
	resp.Write(w)
}

func ListJobs(w http.ResponseWriter, r *http.Request) {
	resp := ResponseOK("")
	resp.Data = cache.List()
	resp.Write(w)
}
