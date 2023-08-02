package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

func ListenForInterrupt(ctx context.Context, cancel context.CancelFunc, callback func(context.Context)) {
	// no matter what, if this function reaches an exit state, its because we need to shut down
	defer cancel()

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// first signal received, begin shutdown
	<-sig
	log.Info().Msg("waiting 5 seconds for server to shut down")

	ctx, cancelTimeout := context.WithTimeout(ctx, 5*time.Second)

	go func() {
		// second signal received, forcefully shut down
		<-sig
		log.Error().Err(ctx.Err()).Msg("context cancelled")
		cancelTimeout()
	}()

	callback(ctx)
}
