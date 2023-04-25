package config

import (
	"fmt"

	"github.com/rs/zerolog"
)

func validateLogLevel(s string) error {
	_, err := zerolog.ParseLevel(s)
	return err
}

const (
	LogFormatJSON = "json"
	LogFormatText = "text"
)

func validateLogFormat(s string) error {
	if s != LogFormatJSON && s != LogFormatText {
		return fmt.Errorf(`value must be one of: "text", "json"`)
	}

	return nil
}

func validateSyncInterval(i int) error {
	if i < 30 {
		return fmt.Errorf("sync interval cannot be less than 30, got: %d", i)
	}

	return nil
}
