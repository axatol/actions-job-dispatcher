package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/axatol/actions-job-dispatcher/pkg/job"
	"github.com/axatol/actions-job-dispatcher/pkg/server"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	config.LoadConfig()

	if len(config.Runners) < 1 {
		log.Fatal().Msg("no runners configured")
	}

	client, err := config.KubeClient()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	serverVersion, err := client.ServerVersion()
	if err != nil {
		log.Fatal().Err(fmt.Errorf("could not retrieve server details: %s", err)).Send()
	}

	log.Info().
		Str("kubernetes_server", serverVersion.GitVersion).
		Strs("serving_runner_labels", config.Runners.Strs()).
		Bool("dry_run", config.DryRun.Value()).
		Int("sync_interval", config.SyncInterval.Value()).
		Send()

	server := server.NewServer()
	ctx, cancel := context.WithCancel(ctx)

	// listen for interrupt
	go listenForInterrupt(ctx, cancel, server)

	// start the server
	go startHTTPServer(server)

	ticker := time.NewTicker(time.Second * time.Duration(config.SyncInterval.Value()))

	for loop := true; loop; {
		select {
		case <-ctx.Done():
			loop = false
		case <-ticker.C:
			if err := job.Reconcile(ctx, client); err != nil {
				log.Error().Err(err).Msg("could not reconcile")
			}
		}
	}
}

func listenForInterrupt(ctx context.Context, cancel context.CancelFunc, server *http.Server) {
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// wait for signal
	<-sig
	log.Info().Msg("waiting 5 seconds for server to shut down")

	// forcefully shut down if server is taking too long
	ctx, cancelTimeout := context.WithTimeout(ctx, 5*time.Second)
	go func() {
		<-ctx.Done()

		if err := ctx.Err(); err == context.DeadlineExceeded {
			log.Fatal().Err(err).Msg("graceful shutdown timed out... forcing exit")
		}
	}()

	// forcefully shut down if second signal received
	go func() {
		<-sig
		log.Fatal().Msg("forcefully shutting down")
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to shut down server")
	}

	cancelTimeout()
	cancel()
}

func startHTTPServer(server *http.Server) {
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server closed unexpectedly")
	}
}
