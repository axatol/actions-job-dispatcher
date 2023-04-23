package config

import (
	"flag"
	"fmt"

	"github.com/rs/zerolog"
)

var (
	ConfigFile string
	LogLevel   = StringFlagValue{defaultValue: zerolog.InfoLevel.String(), validate: validateLogLevel}
	LogFormat  = StringFlagValue{defaultValue: LogFormatJSON, validate: validateLogFormat}
)

func registerGlobalFlags() {
	flag.StringVar(&ConfigFile, "config-file", "", "path to config")
	flag.Var(&LogLevel, "log-level", "logging level")
	flag.Var(&LogFormat, "log-format", `log format, one of: "text", "json"`)
}

const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

func validateLogLevel(s string) error {
	_, err := zerolog.ParseLevel(s)
	return err
}

func validateLogFormat(s string) error {
	if s != LogFormatJSON && s != LogFormatText {
		return fmt.Errorf(`value must be one of: "text", "json"`)
	}

	return nil
}
