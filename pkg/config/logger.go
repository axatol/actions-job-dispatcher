package config

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func configureLogger() {
	level, _ := zerolog.ParseLevel(logLevel.value)
	zerolog.SetGlobalLevel(level)

	if logFormat.value == LogFormatText {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
}
