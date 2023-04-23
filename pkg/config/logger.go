package config

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func configureLogger() {
	level, _ := zerolog.ParseLevel(LogLevel.value)
	zerolog.SetGlobalLevel(level)

	if LogFormat.value == LogFormatText {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
}
