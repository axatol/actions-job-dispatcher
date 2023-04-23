package config

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func configureLogger() {
	zerolog.SetGlobalLevel(LogLevel.value)

	if LogFormat.value == LogFormatText {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
}
