package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/axatol/actions-job-dispatcher/pkg/k8s"
	"github.com/axatol/actions-job-dispatcher/pkg/server"
	"github.com/axatol/actions-job-dispatcher/pkg/util"
	"github.com/rs/zerolog/log"
)

var (
	buildTime   string
	buildCommit string
)

func printVersion() {
	log.Info().
		Str("build_commit", buildCommit).
		Str("build_time", buildTime).
		Str("go_arch", runtime.GOARCH).
		Str("go_os", runtime.GOOS).
		Str("go_version", runtime.Version()).
		Send()
}

func main() {
	config.LoadConfig()

	if config.PrintVersion {
		printVersion()
		return
	}

	serverVersion, err := k8s.Version()
	if err != nil {
		log.Fatal().Err(fmt.Errorf("could not retrieve server details: %s", err)).Send()
	}

	server := server.NewServer(config.ServerPort)
	ctx, cancel := context.WithCancel(context.Background())

	// listen for interrupt
	go util.ListenForInterrupt(ctx, cancel, func(ctx context.Context) {
		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed to shut down server")
		}
	})

	// start the server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("server closed unexpectedly")
		}
	}()

	log.Info().
		Bool("dry_run", config.DryRun).
		Bool("github_token_auth", config.Github.IsToken()).
		Bool("github_app_auth", config.Github.IsApp()).
		Str("log_level", log.Logger.GetLevel().String()).
		Str("kubernetes_context", config.KubeContext).
		Str("kubernetes_namespace", config.Namespace).
		Str("kubernetes_server", serverVersion.GitVersion).
		Strs("serving_runner_labels", config.Runners.Strs()).
		Dur("sync_interval", config.SyncInterval).
		Msgf("server started at http://localhost:%d", config.ServerPort)

	// TODO first time reconcile
	// if err := controller.Reconcile(ctx); err != nil {
	// 	log.Fatal().Err(err).Msg("could not reconcile")
	// }

	ticker := time.NewTicker(config.SyncInterval)
	for loop := true; loop; {
		select {
		case <-ctx.Done():
			loop = false
		case <-ticker.C:
			// TODO regular reconciliation
			// if err := controller.Reconcile(ctx); err != nil {
			// 	log.Fatal().Err(err).Msg("could not reconcile")
			// }
		}
	}

	log.Info().Err(ctx.Err()).Msg("dispatcher exiting")
}
